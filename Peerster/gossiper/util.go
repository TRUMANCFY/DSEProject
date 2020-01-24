package gossiper

import (
	"fmt"
	"math/rand"
	"github.com/LiangweiCHEN/Peerster/message"
	"github.com/LiangweiCHEN/Peerster/routing"
)

func (g *Gossiper) ForwardPkt(pkt *message.GossipPacket, dest string) (err routing.RoutingErr) {
	// Find next hop for destination and forward the packet to next hop

	g.Dsdv.Mux.Lock()
	nextHop, ok := g.Dsdv.Map[dest]
	if !ok {
		err = routing.NewRoutingErr(dest)
		return
	}
	g.Dsdv.Mux.Unlock()
	g.N.Send(pkt, nextHop)
	return
}


func (g *Gossiper) SelectRandomPeer(excluded []string, n int) (rand_peer_addr []string, ok bool) {
	// Select min(n, max_possible) random peers with peers in excluded slice excluded

	g.Peers.Mux.Lock()
	defer g.Peers.Mux.Unlock()

	// Get slice of availble peers
	availablePeers := make([]string, 0)
	excludedMap := make(map[string]bool)
	for _, peer := range excluded {
		excludedMap[peer] = true
	}
	for _, peer := range g.Peers.Peers {
		if _, ok := excludedMap[peer]; !ok {
			availablePeers = append(availablePeers, peer)
		}
	}

	// Return if no available peer exists
	if len(availablePeers) == 0 {
		ok = false
		return
	}

	// Get min(n, len(available_peers)) random neighbour
	selected := make(map[int]bool)
	if n > len(availablePeers) {
		rand_peer_addr = availablePeers
		ok = true
	} else {
		count := 0
		for {
			next := rand.Intn(len(availablePeers))
			if _, ok := selected[next]; !ok {
				selected[next] = true
				count += 1
				rand_peer_addr = append(rand_peer_addr, availablePeers[next])
			}
			if count == n {
				ok = true
				return
			}
		}
	}
	return
}


func (gossiper *Gossiper) UpdatePeers(peerAddr string) {
	// Record new peer with given addr

	gossiper.Peers.Mux.Lock()
	defer gossiper.Peers.Mux.Unlock()

	// Try to find peer addr in self's buffer
	for _, addr := range gossiper.Peers.Peers {
		if peerAddr == addr {
			return
		}
	}

	// Put it in self's buffer if it is absent
	gossiper.Peers.Peers = append(gossiper.Peers.Peers, peerAddr)
}


func (g *Gossiper) MoreUpdated(peer_status message.StatusMap) (moreUpdated int) {
	// Check which of peer and self is more updated
	// Step 1. Loop through self's status to check whether self is more updated
	// Step 2. Loop thourgh peer's status to check whether peer is more updated
	g.StatusBuffer.Mux.Lock()
	defer g.StatusBuffer.Mux.Unlock()

	/* Step 1 */
	for k, v := range g.StatusBuffer.Status {
		peer_v, ok := peer_status[k]
		// Return Self more updated if not ok or peer_v < v
		if !ok || peer_v < v {
			moreUpdated = 1
			return
		}
	}

	/* Step 2 */
	for k, v   := range peer_status {
		self_v, ok := g.StatusBuffer.Status[k]
		// Return peer more updated if not ok or self_v < v
		if !ok || self_v < v {
			moreUpdated = -1
			return
		}
	}

	// Return zero if in sync state
	moreUpdated = 0
	return
}


func (rb *RumorBuffer) get(origin string, ID uint32) (rumor *message.WrappedRumorTLCMessage) {
	// Get the rumor or tlc with corresponding id from specified origin
	rb.Mux.Lock()
	rumor = rb.Rumors[origin][ID-1]
	rb.Mux.Unlock()
	return
}


func (sb *StatusBuffer) ToStatusPacket() (st *message.StatusPacket) {
	// Construct status packet from local status buffer
	// It basically convert map to slice of peer status
	Want := make([]message.PeerStatus, 0)
	sb.Mux.Lock()
	defer sb.Mux.Unlock()
	for k, v := range sb.Status {
		Want = append(Want, message.PeerStatus{
			Identifier: k,
			NextID:     v,
		})
	}
	st = &message.StatusPacket{
		Want: Want,
	}
	return
}


func (g *Gossiper) PrintPeers() {

	outputString := fmt.Sprintf("PEERS ")

	for i, s := range g.Peers.Peers {

		outputString += fmt.Sprintf(s)
		if i < len(g.Peers.Peers)-1 {
			outputString += fmt.Sprintf(",")
		}
	}
	outputString += fmt.Sprintf("\n")
	fmt.Print(outputString)
}


func (g *Gossiper) Update(wrappedMessage *message.WrappedRumorTLCMessage, sender string) (updated bool) {
	// This function attempt to update local cache of messages by comparing
	// the incoming msg's id with local expected.
	// It resolve the blockchain proposal update by checking the confirmed field of TLC message.

	g.StatusBuffer.Mux.Lock()
	defer g.StatusBuffer.Mux.Unlock()
	known_peer := false
	isRumor := wrappedMessage.RumorMessage != nil
	var inputID uint32
	var inputOrigin string
	if isRumor {
		inputID = wrappedMessage.RumorMessage.ID
		inputOrigin = wrappedMessage.RumorMessage.Origin
	} else if wrappedMessage.TLCMessage != nil {
		inputID = wrappedMessage.TLCMessage.ID
		inputOrigin = wrappedMessage.TLCMessage.Origin
	} else {
		inputID = wrappedMessage.BlockRumorMessage.ID
		inputOrigin = wrappedMessage.BlockRumorMessage.Origin
	}

	for origin, nextID := range g.StatusBuffer.Status {

		// Found rumor origin in statusbuffer
		if origin == inputOrigin {

			known_peer = true
			// Input is what is expected !!!
			if nextID == inputID {

				// Update rumor buffer
				g.RumorBuffer.Mux.Lock()
				g.RumorBuffer.Rumors[origin] = append(g.RumorBuffer.Rumors[origin],
					wrappedMessage)
				g.RumorBuffer.Mux.Unlock()

				// Update StatusBuffer
				g.StatusBuffer.Status[origin] += 1
				updated = true

				//fmt.Println("Receive rumor originated from " + rumor.Origin + " with ID " +
				// strconv.Itoa(int(rumor.ID)) + " relayed by " + sender)

				return
			}
		}
	}
	// Handle rumor originated from a new peer
	if inputID == 1 && !known_peer {

		// Put entry for origin into self's statusBuffer
		g.StatusBuffer.Status[inputOrigin] = 2
		// Buffer current rumor
		g.RumorBuffer.Mux.Lock()
		g.RumorBuffer.Rumors[inputOrigin] = []*message.WrappedRumorTLCMessage{wrappedMessage}
		g.RumorBuffer.Mux.Unlock()
		updated = true

		// fmt.Println("Receive rumor originated from " + rumor.Origin + " with ID " + strconv.Itoa(int(rumor.ID)) +
		//   " relayed by " + sender)
		return
	}
	// Fail to update, either out of date or too advanced
	updated = false
	return
}