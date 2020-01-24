package gossiper

import (
	//"github.com/LiangweiCHEN/Peerster/message"
)

func (g *Gossiper) Advance() {
	/* This function advance the tlc to next round of QSC */

	// Reset proposal for next round
	g.QSCMessage.Step++
	g.QSCMessage.Type = Raw
	g.Acks = 0
	g.Wits = 0

	// Trigger advance of QSC
	//g.AdvanceQSC()

	// TODO: Trigger broadcast of new raw proposal
	//g.BroadcastTLC()
}