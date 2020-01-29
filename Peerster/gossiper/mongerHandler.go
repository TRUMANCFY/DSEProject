package gossiper

import (
	"fmt"
	"time"
	"math/rand"
	"strconv"
	"github.com/LiangweiCHEN/Peerster/message"
)

func (g *Gossiper) MongerRumor(wrappedMessage *message.WrappedRumorTLCMessage, target string, excluded []string) {
	// 1. Select a random peer if no target specified, else set dest to target
	// 2. Register spoofing ack and timeout channel
	// 3. Send the rumor
	// 4. FlipCoinMonger if rumor is not needed or timeout

	/* Step 1 */
	var peerAddr string
	if target == "" {
		var ok bool
		peerAddrSlice, ok := g.SelectRandomPeer(excluded, 1)
		if !ok {
			return
		}
		peerAddr = peerAddrSlice[0]
	} else {
		peerAddr = target
	}
	if wrappedMessage.BlockRumorMessage != nil {
		// block := wrappedMessage.BlockRumorMessage.Block
		// fmt.Printf("PROPOSING BLOCK WITH VOTER %s VOTE %s TO PEER %s FROM %s IN ROUND %d\n", block.CastBallot.VoterUuid,
		// 																 block.CastBallot.VoteHash,
		// 																peerAddr,
		// 																block.Origin,
		// 																block.Round)
	}
	/* Step 2 & 4 */
	go func() {

		timeout := time.After(10 * time.Second)
		// Generate and store ack_ch
		ackCh := make(chan *PeerStatusAndSync)
		isRumor := wrappedMessage.RumorMessage != nil
		msgOrigin := wrappedMessage.GetOrigin()
		msgID := wrappedMessage.GetID()
		key := peerAddr + msgOrigin + strconv.Itoa(int(msgID))
		g.AckChs.Mux.Lock()
		g.AckChs.Chs[key] = ackCh
		g.AckChs.Mux.Unlock()

		// Start waiting
		for {
			select {
			case peerStatusAndSync := <-ackCh:
				// Receive a statusAndSync from ack ch
				// fmt.Printf("Origin %s Dst %s Rumor id %d rumor content %s \n", rumor.Origin, target, rumor.ID, rumor.Text)

				if peerStatusAndSync == nil && isRumor {
					fmt.Println("Peer status and sync nil")
					fmt.Printf("Origin %s Dst %s Rumor id %d rumor content %s \n", msgOrigin, target, msgID, wrappedMessage.RumorMessage.Text)
				}
				peerStatus, isSync := peerStatusAndSync.PeerStatus, peerStatusAndSync.IsSync
				g.AckChs.Mux.Lock()
				delete(g.AckChs.Chs, key)
				close(ackCh)
				g.AckChs.Mux.Unlock()

				if peerStatus.NextID == wrappedMessage.GetID()-1 && isSync {

					g.FlipCoinMonger(wrappedMessage, []string{peerAddr})
					return
				} else if !isSync && peerStatus.NextID > msgID-1 {

					return
				}

			case <-timeout:
				g.AckChs.Mux.Lock()
				delete(g.AckChs.Chs, key)
				close(ackCh)
				g.AckChs.Mux.Unlock()

				g.FlipCoinMonger(wrappedMessage, []string{peerAddr})
				return
			}
		}
	}()

	/* Step 3 */
	var toSendPkt *message.GossipPacket
	if wrappedMessage.RumorMessage != nil {
		toSendPkt = &message.GossipPacket{
			Rumor: wrappedMessage.RumorMessage,
		}
	} else if wrappedMessage.TLCMessage != nil {
		toSendPkt = &message.GossipPacket{
			TLCMessage: wrappedMessage.TLCMessage,
		}
	} else {
		toSendPkt = &message.GossipPacket{
			BlockRumorMessage: wrappedMessage.BlockRumorMessage,
		}
	}
	g.N.Send(toSendPkt, peerAddr)
}


func (g *Gossiper) ProvideMongering(peer_status message.StatusMap, sender string) {
	// Send to the request sender with the most urgent msg it requires
	g.StatusBuffer.Mux.Lock()
	defer g.StatusBuffer.Mux.Unlock()

	// Find the missing rumor with least ID to monger
	for k, v := range g.StatusBuffer.Status {
		switch peer_v, ok := peer_status[k]; {
		// Send the first rumor from current origin if peer have not heard
		// from it
		case !ok:
			g.MongerRumor(g.RumorBuffer.get(k, 1), sender, []string{})
			return
		case peer_v < v:
			g.MongerRumor(g.RumorBuffer.get(k, peer_v), sender, []string{})
			return
		}
	}
}

func (g *Gossiper) RequestMongering(peer_status message.StatusMap, sender string) {
	// Request mongering by send the status pkt to peer
	g.N.Send(&message.GossipPacket{
		Status: g.StatusBuffer.ToStatusPacket(),
	}, sender)
}

func (g *Gossiper) FlipCoinMonger(rumor *message.WrappedRumorTLCMessage, excluded []string) {

	// Flip a coin to decide whether to continue mongering
	continue_monger := rand.Int() % 2
	if continue_monger == 0 {
		return
	} else {
		if peer_addr_slice, ok := g.SelectRandomPeer(excluded, 1); !ok {
			return
		} else {
			// fmt.Printf("FLIPPED COIN sending rumor to %s\n", peer_addr)
			g.MongerRumor(rumor, peer_addr_slice[0], []string{})
		}
	}
}
