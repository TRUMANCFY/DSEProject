package gossiper

import (
	"fmt"
	"github.com/LiangweiCHEN/Peerster/message"
)

func (g *Gossiper) HandlePrivateMsg(wrapped_pkt *message.PacketIncome) {
	// 1. Decrement the hoplimit
	// 2. Accept the packet if self is the terminal
	// 3. Decrement and check for timeout, stop forwarding if timeout
	// 4. Send to next hop router with decremented HopLimit

	// Step 1
	pkt := wrapped_pkt.Packet
	pkt.Private.HopLimit -= 1

	if pkt.Private.Destination == g.Name {

		fmt.Printf("PRIVATE origin %s hop-limit %d contents %s\n",
			pkt.Private.Origin,
			int(pkt.Private.HopLimit),
			pkt.Private.Text)

		return
	}

	// Step 2
	if pkt.Private.HopLimit == 0 {
		return
	}

	// Step 3
	g.Dsdv.Mux.Lock()
	nextHop := g.Dsdv.Map[pkt.Private.Destination]
	g.Dsdv.Mux.Unlock()
	g.N.Send(pkt, nextHop)
}