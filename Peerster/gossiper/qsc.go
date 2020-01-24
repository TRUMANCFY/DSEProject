package gossiper

import (
	"github.com/LiangweiCHEN/Peerster/message" 
)
func (g *Gossiper) QSCRound(block *message.BlockPublish) {
	/*
		1. Send tlc message with block embedded to all the peers (t_0)
		2. Select the confirmed BlockPublish with highest fitness values (t_1)
		3. Send the tlc message with the block obtained from above
		4. Select the confirmed BlockPublish with highest fitness values (t_2)
		5. Send the tlc message with the block obtained from above 
		6. Select the confiremd blockpublish with highest fitness value (t_3)
		5. Add the block to the blockchain
	*/

}
