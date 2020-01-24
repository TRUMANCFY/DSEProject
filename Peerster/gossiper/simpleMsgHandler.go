package gossiper

import (
	"fmt"
	"github.com/LiangweiCHEN/Peerster/message"
)

func (g *Gossiper) HandleSimple(wrapped_pkt *message.PacketIncome) {
	// 0. Handle simple flag case
	// 1. Construct and store rumor
	// 2. Update self's status
	// 3. Trigger rumor mongering

	// Step 0. Handle simple flag case
	packet := wrapped_pkt.Packet

	if g.Simple {

		if packet.Simple.OriginalName == "client" {

			// Output msg content
			fmt.Printf("CLIENT MESSAGE %s\n", packet.Simple.Contents)

			g.PrintPeers()
			// Broadcast
			packet.Simple.OriginalName = g.Name
			packet.Simple.RelayPeerAddr = g.Address

			g.Peers.Mux.Lock()
			for _, peer_addr := range g.Peers.Peers {

				g.N.Send(packet, peer_addr)
			}
			g.Peers.Mux.Unlock()
		} else {

			// Output msg content
			fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
				packet.Simple.OriginalName,
				packet.Simple.RelayPeerAddr,
				packet.Simple.Contents)
			g.PrintPeers()
			// Broadcast pkt to all peers apart from relayers
			g.Peers.Mux.Lock()
			relayPeerAddr := packet.Simple.RelayPeerAddr
			packet.Simple.RelayPeerAddr = g.Address

			for _, peer_addr := range g.Peers.Peers {

				if peer_addr != relayPeerAddr {

					// fmt.Printf("Sending simple message from %s to %s\n", packet.Simple.RelayPeerAddr, peer_addr)

					g.N.Send(packet, peer_addr)
				}
			}
			g.Peers.Mux.Unlock()
		}
		return
	}
}