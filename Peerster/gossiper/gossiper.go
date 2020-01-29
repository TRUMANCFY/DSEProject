package gossiper

import (
	"fmt"
	"github.com/LiangweiCHEN/Peerster/fileSharing"
	"github.com/LiangweiCHEN/Peerster/message"
	"github.com/LiangweiCHEN/Peerster/network"
	"github.com/LiangweiCHEN/Peerster/routing"
	"net"
	"sync"
	"time"
)

type Gossiper struct {
	Address            string
	Conn               *net.UDPConn
	Name               string
	UIPort             string
	GuiPort            string
	StubbornTimeout    int
	NumPeers           int
	Peers              *PeersBuffer
	Simple             bool
	N                  *network.NetworkHandler
	RumorBuffer        *RumorBuffer
	StatusBuffer       *StatusBuffer
	AckChs             *Ack_chs
	PeerStatuses       *PeerStatuses
	AntiEntropyPeriod  int
	Dsdv               *routing.DSDV
	RTimer             int
	HopLimit           uint32
	SharedFilePath     string
	FileSharer         *fileSharing.FileSharer
	SearchDistributeCh chan *message.SearchRequestRelayer
	TLCAckChs          *TLCAckChs
	TLCAckCh           chan *message.PacketIncome
	TLCClock           *TLCClock
	TLCRoundCh			chan struct{}
	WrappedTLCCh		chan *WrappedTLCMessage
	ConfirmedMessageCh	chan *message.TLCMessage
	TransactionSendCh	chan *message.TxPublish
	Hw3ex2				bool
	Hw3ex3				bool
	Round 				int
	AckAll				bool
	MsgBuffer			MsgBuffer
	// Stuff for Que sera consensus
	Rand 				func() int64 // Function to generate random ticket for QSC
	QSCMessage			QSCMessage // QSC Message holder
	Acks				int // Number of acks for QSC proposal
	Wits				int // Number of threshold witnessed messages

	// Stuff for blockchain
	Blockchains			map[string]*Blockchain
	BlockchainsMux		sync.Mutex
}

// Gossiper start working
func (gossiper *Gossiper) StartWorking() {

	// Start network working
	gossiper.N.StartWorking()

	// Start Receiving
	gossiper.StartHandling()

	// Start search sending
	gossiper.StartSearching()

	// Start antiEntropy sending
	if !gossiper.Simple {
		gossiper.StartAntiEntropy()
	}

	// Start routing
	gossiper.Dsdv.StartRouting()

	// Start heartbing
	gossiper.StartHeartbeat()

	// Start gui handling
	gossiper.HandleGUI()

	// Start handling downloading after search
	go gossiper.FileSharer.HandleSearchedFileDownload()

	// Start handling tlc ack
	go gossiper.HandleTLCAck()

	// Start handling sending candidate blocks
	// go gossiper.HandleSendingBlocks()

	// Start round tlc ack if hw3ex3
	if gossiper.Hw3ex3 {
		go gossiper.RoundTLCAck()
		confirmCh := make(chan struct{})
		sendCh := make(chan struct{})
		go gossiper.HandleConfirmedMessage(confirmCh, sendCh)
		go gossiper.HandleRoundSend(confirmCh, sendCh)
	}
}

func (gossiper *Gossiper) StartHandling() {
	// This function start go routine to trigger different handlers for incoming message
	// from either gossiper or client
	go func() {

		for pkt := range gossiper.N.Listen_ch {

			pkt := pkt
			// Start handling packet content
			switch {

			case pkt.Packet.Simple != nil:
				// Handle simple message
				if gossiper.Simple && pkt.Packet.Simple.OriginalName != "client" {
					gossiper.UpdatePeers(pkt.Packet.Simple.RelayPeerAddr)
				}
				gossiper.HandleSimple(pkt)

			case pkt.Packet.Rumor != nil:
				// Handle rumor message
				// Update peers
				gossiper.UpdatePeers(pkt.Sender)
				// Print peers
				// gossiper.PrintPeers()
				go gossiper.HandleRumor(pkt)

			case pkt.Packet.Status != nil:
				// Handle status pkt
				// Update peers
				gossiper.UpdatePeers(pkt.Sender)
				// Print peers
				// gossiper.PrintPeers()
				go gossiper.HandleStatus(pkt)

			case pkt.Packet.Private != nil:
				// Handle private message
				// TODO: Decide whether need to update peers here
				go gossiper.HandlePrivateMsg(pkt)

			case pkt.Packet.DataRequest != nil:
				// Handle data request
				// Forward pkt if dest is not self
				if pkt.Packet.DataRequest.Destination != gossiper.Name {
					pkt.Packet.DataRequest.HopLimit -= 1
					if pkt.Packet.DataRequest.HopLimit > 0 {
						gossiper.ForwardPkt(pkt.Packet, pkt.Packet.DataRequest.Destination)
					}
				} else {
					// Handle pkt if dest is self
					go gossiper.FileSharer.HandleRequest(pkt)
				}

			case pkt.Packet.DataReply != nil:
				// Handle data reply
				// Forward pkt if dest is not self
				if pkt.Packet.DataReply.Destination != gossiper.Name {
					pkt.Packet.DataReply.HopLimit -= 1
					if pkt.Packet.DataReply.HopLimit > 0 {
						gossiper.ForwardPkt(pkt.Packet, pkt.Packet.DataReply.Destination)
					}
				} else {
					go gossiper.FileSharer.HandleReply(pkt)
				}

			case pkt.Packet.SearchRequest != nil:
				// Handle search request in sharer
				// Do nothing if self is the origin
				if pkt.Packet.SearchRequest.Origin == gossiper.Name {
					continue
				}
				go gossiper.FileSharer.HandleSearch(pkt.Packet.SearchRequest, pkt.Sender)

			case pkt.Packet.SearchReply != nil:
				// Handle search reply
				// Forward pkt if dest is not self
				if pkt.Packet.SearchReply.Destination != gossiper.Name {
					pkt.Packet.SearchReply.HopLimit -= 1
					if pkt.Packet.SearchReply.HopLimit > 0 {
						gossiper.ForwardPkt(pkt.Packet, pkt.Packet.SearchReply.Destination)
					}
				} else {
					if gossiper.FileSharer.Searcher.ReplyCh != nil {
						gossiper.FileSharer.Searcher.ReplyCh <- pkt.Packet.SearchReply
					}
				}
			case pkt.Packet.TLCMessage != nil:
				// Handle transaction proposal
				if pkt.Packet.TLCMessage.Origin != gossiper.Name {
					go gossiper.HandleTLCMessage(pkt)
				}

			case pkt.Packet.ACK != nil:
				// Handle transaction acknowledgement
				// Forward pkt if dest is not self
				if pkt.Packet.ACK.Destination != gossiper.Name {
					pkt.Packet.ACK.HopLimit -= 1
					if pkt.Packet.ACK.HopLimit > 0 {
						gossiper.ForwardPkt(pkt.Packet, pkt.Packet.ACK.Destination)
					}
				} else {
					gossiper.TLCAckCh <- pkt
				}

			case pkt.Packet.BlockRumorMessage != nil:
				// Handle blockRumorMessage
				// Pass the block to blockchain handler
				go gossiper.HandleReceivingBlock(pkt)
			}

		}
	}()

	go func() {

		for msg := range gossiper.N.Client_listen_ch {

			gossiper.HandleClient(msg)
		}
	}()
}

// Begin antiEntropy to ensure the delivery of packets
func (gossiper *Gossiper) StartAntiEntropy() {

	go func() {

		// Set up ticker
		ticker := time.NewTicker(time.Duration(gossiper.AntiEntropyPeriod) * time.Second)

		for _ = range ticker.C {
			// Get random peer
			if randPeerSlice, ok := gossiper.SelectRandomPeer([]string{}, 1); ok {
				// Send status to selected peer
				gossiper.N.Send(&message.GossipPacket{
					Status: gossiper.StatusBuffer.ToStatusPacket(),
				}, randPeerSlice[0])
			}
		}
	}()
}


func (g *Gossiper) StartSearching() {
	go g.TriggerSearch()
	go g.DistributeSearch()
}


func (g *Gossiper) StartHeartbeat() {

	go func() {

		// Stop heartbeat if RTime is zero
		if g.RTimer == 0 {

			return
		}

		// Start initial heart beat
		/* Need modification if heartbeat start from init */

		g.StatusBuffer.Mux.Lock()
		g.RumorBuffer.Mux.Lock()
		g.StatusBuffer.Status[g.Name] = 2

		rumor := &message.RumorMessage{
			Origin: g.Name,
			ID:     uint32(1),
			Text:   "",
		}
		wrappedMessage := &message.WrappedRumorTLCMessage{
			RumorMessage: rumor,
		}
		// fmt.Printf("Initial rumor is %d", 1)

		g.RumorBuffer.Rumors[g.Name] = append(g.RumorBuffer.Rumors[g.Name], wrappedMessage)
		g.RumorBuffer.Mux.Unlock()
		g.StatusBuffer.Mux.Unlock()

		g.MongerRumor(wrappedMessage, "", nil)

		// Periodically heartbeat

		ticker := time.NewTicker(time.Duration(g.RTimer) * time.Second)

		for _ = range ticker.C {

			// Periodically heartbeat
			g.StatusBuffer.Mux.Lock()
			g.RumorBuffer.Mux.Lock()
			id := g.StatusBuffer.Status[g.Name]
			g.StatusBuffer.Status[g.Name] += 1

			rumor := &message.RumorMessage{
				Origin: g.Name,
				ID:     uint32(id),
				Text:   "",
			}
			wrappedMessage := &message.WrappedRumorTLCMessage{
				RumorMessage: rumor,
			}
			g.RumorBuffer.Rumors[g.Name] = append(g.RumorBuffer.Rumors[g.Name], wrappedMessage)
			g.StatusBuffer.Mux.Unlock()
			g.RumorBuffer.Mux.Unlock()
			g.MongerRumor(wrappedMessage, "", nil)
			fmt.Println("Heartbeating......", id)
		}

	}()
}