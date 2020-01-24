package gossiper

import (
	"fmt"
	"strconv"
	"github.com/LiangweiCHEN/Peerster/message"
	"github.com/LiangweiCHEN/Peerster/routing"
)

func (g *Gossiper) HandleRumor(wrapped_pkt *message.PacketIncome) {
	// This function handle incoming rumor
	// Step 1. Update local cache of rumors
	// Step 2. Update routing table if necessary
	// Step 3. Monger rumor if updated
	// Step 4. Send back self's status packet

	sender, rumor := wrapped_pkt.Sender, wrapped_pkt.Packet.Rumor

	/* Step 1 */
	updated := g.Update(&message.WrappedRumorTLCMessage{
		RumorMessage: rumor,
	}, sender)

	/* Step 4 */
	defer g.N.Send(&message.GossipPacket{
		Status: g.StatusBuffer.ToStatusPacket(),
	}, sender)

	// Trigger rumor mongering if it is new,
	// Trigger update routing table if it is new
	if updated {

		/* Step 3 */
		var heartbeat bool
		if rumor.Text == "" {
			heartbeat = true
		} else {
			heartbeat = false
		}
		g.Dsdv.Ch <- &routing.OriginRelayer{
			Origin:    rumor.Origin,
			Relayer:   sender,
			HeartBeat: heartbeat,
		}

		// Output rumor content only if it is not heartbeat rumor

		output := fmt.Sprintf("RUMOR origin %s from %s ID %s contents %s\n", rumor.Origin, sender, strconv.Itoa(int(rumor.ID)), rumor.Text)
		
		fmt.Print(output)
		if rumor.Text != "" {
			g.MsgBuffer.Mux.Lock()
			g.MsgBuffer.Msg = append(g.MsgBuffer.Msg, output)
			g.MsgBuffer.Mux.Unlock()
		}
		/* Step 4 */
		wrappedMessage := &message.WrappedRumorTLCMessage{
			RumorMessage: rumor,
		}
		g.MongerRumor(wrappedMessage, "", []string{sender})
	}

}

