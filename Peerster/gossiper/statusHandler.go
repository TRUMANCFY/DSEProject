package gossiper


import (
	"strconv"
	"github.com/LiangweiCHEN/Peerster/message"
)


func (g *Gossiper) HandleStatus(wrappedPkt *message.PacketIncome) {
	// Step 1. Convert peer_status to map
	// Step 2. Ack rumors sended to peer that are not needed as indicated by its status
	// Step 3. Check whether need to provide or request for mongering

	/* Step 1 */
	sender, peerStatus := wrappedPkt.Sender, wrappedPkt.Packet.Status
	peerStatusMap := peerStatus.ToMap()

	// Ouput peer status information
	/*
		outputString := fmt.Sprintf("STATUS from %s ", sender)
		for k, v := range peer_status_map {
			outputString += fmt.Sprintf("peer %s nextID %s ", k, strconv.Itoa(int(v)))
		}
		outputString += fmt.Sprintf("\n")
		fmt.Print(outputString)
	*/

	/* Step 2. Ack all pkts received or not needed by peer */
	moreUpdated := g.MoreUpdated(peerStatusMap)
	go g.Ack(peerStatusMap, sender, moreUpdated == 0)

	// Step 3. Provide new mongering or Request mongering
	switch moreUpdated {
	case 1:
		g.ProvideMongering(peerStatusMap, sender)
	case -1:
		g.RequestMongering(peerStatusMap, sender)
	default:
		// Already handled in Ack
		// fmt.Printf("IN SYNC WITH %s\n", sender)
	}
}


func (g *Gossiper) Ack(peer_status_map message.StatusMap, sender string, isSync bool) {
	// Step 1. Ack pkts being sent to peer but not needed as shown by its status map
	// Step 2. Update PeerStatuses

	g.PeerStatuses.Mux.Lock()
	if _, ok := g.PeerStatuses.Map[sender]; !ok {
		g.PeerStatuses.Map[sender] = make(map[string]uint32)
	}
	g.AckChs.Mux.Lock()

	for origin, nextid := range peer_status_map {

		if _, ok := g.PeerStatuses.Map[sender][origin]; !ok {
			g.PeerStatuses.Map[sender][origin] = uint32(1)
		}

		// Construct newest peerstatus for current origin
		peerStatus := &message.PeerStatus{
			Identifier: origin,
			NextID:     nextid,
		}

		/* Step 1 */
		for id := g.PeerStatuses.Map[sender][origin]; id <= nextid; id += 1 {
			if ack_ch, ok := g.AckChs.Chs[sender+origin+strconv.Itoa(int(nextid))]; ok {
				ack_ch <- &PeerStatusAndSync{
					PeerStatus: peerStatus,
					IsSync:     isSync,
				}
			}
		}

		/* Step 2 */
		g.PeerStatuses.Map[sender][origin] = nextid
	}

	g.AckChs.Mux.Unlock()
	g.PeerStatuses.Mux.Unlock()
}