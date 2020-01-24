package gossiper

import (
	"fmt"
	"strings"
	"github.com/LiangweiCHEN/Peerster/message"
)

func (g *Gossiper) HandleClient(msg *message.Message) {

	switch {
	case g.Simple:
		fmt.Printf("CLIENT MESSAGE %s\n", msg.Text)
		g.PrintPeers()

		// Broadcast the simple msg to all the peers 
		pkt := &message.GossipPacket{
			Simple: &message.SimpleMessage{
				OriginalName:  g.Name,
				RelayPeerAddr: g.Address,
				Contents:      msg.Text,
			},
		}
		g.Peers.Mux.Lock()
		for _, peer_addr := range g.Peers.Peers {

			g.N.Send(pkt, peer_addr)
		}
		g.Peers.Mux.Unlock()

	case msg.Destination == nil && msg.File == nil && len(msg.Text) != 0:
		// Handle mongering rumor case

		// Step 1. Construct rumor message

		fmt.Printf("CLIENT MESSAGE %s\n", msg.Text)
		g.RumorBuffer.Mux.Lock()
		defer g.RumorBuffer.Mux.Unlock()
		rumor := &message.RumorMessage{
			Origin: g.Name,
			ID:     uint32(len(g.RumorBuffer.Rumors[g.Name]) + 1),
			Text:   msg.Text,
		}

		// Store rumor
		wrappedMessage := &message.WrappedRumorTLCMessage{
			RumorMessage: rumor}
		g.RumorBuffer.Rumors[g.Name] = append(g.RumorBuffer.Rumors[g.Name], wrappedMessage)

		// Step 2. Update status
		g.StatusBuffer.Mux.Lock()
		defer g.StatusBuffer.Mux.Unlock()

		if _, ok := g.StatusBuffer.Status[g.Name]; !ok {

			g.StatusBuffer.Status[g.Name] = 2
		} else {

			g.StatusBuffer.Status[g.Name] += 1
		}

		// Step 3. Trigger rumor mongering
		g.MongerRumor(wrappedMessage, "", []string{})

	case msg.File == nil && msg.Destination != nil:
		// Handle private message sending
		// 1. Find next hop router
		// 2. Create private message with init hop limit and zero id
		// 3. Directly send private msg to next hop
		// TODO: Solve pkt lost problem

		// Step 1
		fmt.Printf("CLIENT MESSAGE %s dest %s\n", msg.Text, *msg.Destination)
		g.Dsdv.Mux.Lock()
		nextHop := g.Dsdv.Map[*msg.Destination]
		g.Dsdv.Mux.Unlock()

		// Step 2
		privatePkt := &message.GossipPacket{
			Private: &message.PrivateMessage{
				Origin:      g.Name,
				ID:          0,
				Destination: *msg.Destination,
				Text:        msg.Text,
				HopLimit:    g.HopLimit,
			},
		}

		// Step 3
		g.N.Send(privatePkt, nextHop)
	case msg.File != nil && msg.Request == nil:
		// Handle File indexing
		// 1. Trigger fileSharing obj indexing
		// 2. Trigger broadcasting the proposed name to the peers
		if !g.Hw3ex3 {
			go func(fileName *string) {
				tx, _ := g.FileSharer.CreateIndexFile(fileName)
				round := g.Round
				g.SendTLC(*tx, round)
			}(msg.File)
		} else {
			go func(fileName *string) {
				tx, _ := g.FileSharer.CreateIndexFile(fileName)
				g.TransactionSendCh<- tx
			}(msg.File)
		}

	case msg.File != nil && msg.Request != nil && msg.Destination != nil:
		// Handle file request
		// 1. Trigger fileSharing obj requesting

		// fmt.Printf("SEND REQUEST FOR %s TO %v\n", hex.EncodeToString(*msg.Request), msg.Destination)
		go g.FileSharer.RequestFile(msg.File, msg.Request, msg.Destination)

	case msg.File != nil && msg.Request != nil && msg.Destination == nil:
		// Handle downloading searched file
		go g.FileSharer.Searcher.RequestSearchedFile(*msg.File, *msg.Request)

	case len(msg.Keywords) > 0:
		// Handle file search
		fmt.Printf("CLIENT WANT TO SEARCH FOR %s\n", strings.Join(msg.Keywords, ","))
		g.FileSharer.Searcher.Search(msg.Keywords, int(msg.Budget))

	case msg.Voterid != "" && msg.Vote != "":
		// Handle vote
		fmt.Printf("CLIENT SEND VOTE FROM %s WITH CONTENT %s\n", msg.Voterid, msg.Vote)
		v := g.Blockchain.CreateBallot(msg.Voterid, msg.Vote)
		go g.HandleReceivingVote(v)
	}
}