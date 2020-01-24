package gossiper

import (
	"fmt"
	"encoding/hex"
	"time"
	"math"
	"strings"
	"github.com/LiangweiCHEN/Peerster/message"
	"github.com/LiangweiCHEN/Peerster/routing"
)

func (g *Gossiper) HandleTLCMessage(wrapped_pkt *message.PacketIncome) {
	/*	Step 1. Update records if the TLC message is unseen
		Step 2. Send back ACK if unseen and not confirmed
		Step 3. Trigger same process as normal rumor message 
	*/
	
	/* Step 1 */
	// Decode wrapped pkt
	sender, tlc := wrapped_pkt.Sender, wrapped_pkt.Packet.TLCMessage

	// Update status and rumor buffer
	updated := g.Update(&message.WrappedRumorTLCMessage{
		TLCMessage : tlc,
	}, sender)

	// Defer sending status back
	defer g.N.Send(&message.GossipPacket{
			Status : g.StatusBuffer.ToStatusPacket(),
			}, sender)

	if updated {
		/* Step 2 */
		if tlc.Confirmed == -1 {
			outputStr := fmt.Sprintf("UNCONFIRMED GOSSIP origin %s ID %d file name %s size %d metahash %s\n",
			tlc.Origin,
			tlc.ID,
			tlc.TxBlock.Transaction.Name,
			tlc.TxBlock.Transaction.Size,
			hex.EncodeToString(tlc.TxBlock.Transaction.MetafileHash),
			)
			fmt.Printf(outputStr)
			if (!g.Hw3ex3) { g.ACK(tlc.ID, tlc.Origin) } else {
				wrappedTLC := g.ComputeRound(tlc)
				g.WrappedTLCCh<- wrappedTLC
			}
		} else {
			outputStr := fmt.Sprintf("CONFIRMED GOSSIP origin %s ID %d file name %s size %d metahash %s\n",
				tlc.Origin,
				tlc.ID,
				tlc.TxBlock.Transaction.Name,
				tlc.TxBlock.Transaction.Size,
				hex.EncodeToString(tlc.TxBlock.Transaction.MetafileHash),
				)
				fmt.Printf(outputStr)
				if g.Hw3ex2 || g.Hw3ex3 {
					g.MsgBuffer.Mux.Lock()
					g.MsgBuffer.Msg = append(g.MsgBuffer.Msg, outputStr)
					g.MsgBuffer.Mux.Unlock()
				}
			if g.Hw3ex3 {
				g.ConfirmedMessageCh<- tlc
			}
		}

		/* Step 3 */
		// Triger update routing
		heartbeat := false
		g.Dsdv.Ch<- &routing.OriginRelayer{
			Origin : tlc.Origin,
			Relayer : sender,
			HeartBeat : heartbeat,
		}

		// Output rumor content only if it is not heartbeat rumor
	
		wrappedMessage := &message.WrappedRumorTLCMessage {
			TLCMessage : tlc,
		}
		g.MongerRumor(wrappedMessage, "", []string{sender})
	}
}

func (g *Gossiper) ACK(ID uint32, destination string) {
	// Step 1. Get next hop of destination
	// Step 2. Construct ack msg
	// Step 3. Send ACK back to destination

	fmt.Printf("SENDING ACK origin %s ID %d\n", destination, ID)
	/* Step 1 */
	g.Dsdv.Mux.Lock()
	nextHop := g.Dsdv.Map[destination]
	g.Dsdv.Mux.Unlock()

	/* Step 2 */
	ack := &message.TLCAck{
		Origin : g.Name,
		ID : ID,
		Text : "",
		Destination : destination,
		HopLimit : g.HopLimit,
	}

	/* Step 3 */
	g.N.Send(&message.GossipPacket{
		ACK : ack,
	}, nextHop)
}

func (g *Gossiper) SendTLC(tx message.TxPublish, round int) {
	// Step 1. Create TLCMessage
	// Step 2. Register unconfirmed TLC message
	// Step 3. Periodically monger TLC message till 
	// receive terminating signal
	// Step 4. Broadcast confirmation to all peers after receiving majority ack

	/* Step 1 */
	g.RumorBuffer.Mux.Lock()
	ID := uint32(len(g.RumorBuffer.Rumors[g.Name]) + 1)
	outputStr := fmt.Sprintf("UNCONFIRMED GOSSIP origin %s ID %d file name %s size %d metahash %s\n",
					g.Name,
					ID,
					tx.Name,
					tx.Size,
					hex.EncodeToString(tx.MetafileHash))

	var tlc *message.TLCMessage
	if !g.Hw3ex3 {
		tlc = &message.TLCMessage{
			Origin : g.Name,
			ID : ID,
			Confirmed : -1,
			TxBlock : message.BlockPublish{
				Transaction : tx,
			},
			VectorClock : nil,
			Fitness : 0,
		}
	} else {
		tlc = &message.TLCMessage{
			Origin : g.Name,
			ID : ID,
			Confirmed : -1,
			TxBlock : message.BlockPublish{
				Transaction : tx,
			},
			VectorClock : g.StatusBuffer.ToStatusPacket(),
			Fitness : 0,
		}
	}
	// Increment self's round
	g.ComputeRound(tlc)
	wrappedMessage := &message.WrappedRumorTLCMessage{
		TLCMessage : tlc,
	}

	// Store new tlc message into rumor buffer
	g.RumorBuffer.Rumors[g.Name] = append(g.RumorBuffer.Rumors[g.Name], &message.WrappedRumorTLCMessage{
		TLCMessage : tlc,
	})
	g.RumorBuffer.Mux.Unlock()

	// Update status
	g.StatusBuffer.Mux.Lock()
	if _, ok := g.StatusBuffer.Status[g.Name]; !ok {
		g.StatusBuffer.Status[g.Name] = 2
	} else {
		g.StatusBuffer.Status[g.Name] += 1
	}
	g.StatusBuffer.Mux.Unlock()

	// Directly confirm if less than three peers in the system
	var witnesses []string
	if g.NumPeers >= 3 {

		/* Step 2 */
		terminateCh := make(chan []string)
		g.TLCAckChs.Mux.Lock()
		g.TLCAckChs.Chs[ID] = terminateCh
		g.TLCAckChs.Mux.Unlock()

		/* Step 3 */
		// Trigger initial broadcast of tlc
		g.MongerRumor(wrappedMessage, "", []string{})
		fmt.Printf(outputStr)
		ticker := time.NewTicker(time.Duration(g.StubbornTimeout) * time.Second)

		consensus := false
		for {
			select {
			case <-ticker.C:
				// Timeout before receiving enough ack
				// Trigger mongering again
				if g.Hw3ex3 && g.Round > round {
					fmt.Println("WE HAVE ALREADY ADVANCED TO NEXT ROUND")
					return
				}
				g.MongerRumor(wrappedMessage, "", []string{})
				fmt.Println(len(witnesses))
				fmt.Printf("RE-BROADCAST ID %d WITNESSES %s\n", ID, strings.Join(witnesses, ","))
			case witnesses = <-terminateCh:
				// Receive ack from some peer
				if (len(witnesses) >= int(math.Ceil(float64(g.NumPeers) / 2))) {
					ticker.Stop()
					consensus = true
					break
				}
			}
			if consensus {
				break
			}
		}
	}

	/* Step 4 */
	g.RumorBuffer.Mux.Lock()
	confirmedMsgID := uint32(len(g.RumorBuffer.Rumors[g.Name]) + 1)

	if !g.Hw3ex3 {
		tlc = &message.TLCMessage{
			Origin : g.Name,
			ID : confirmedMsgID,
			Confirmed : int(ID),
			TxBlock : message.BlockPublish{
				Transaction : tx,
			},
			VectorClock : nil,
			Fitness : 0,
		}
	} else {
		tlc = &message.TLCMessage{
			Origin : g.Name,
			ID : confirmedMsgID,
			Confirmed : int(ID),
			TxBlock : message.BlockPublish{
				Transaction : tx,
			},
			VectorClock : g.StatusBuffer.ToStatusPacket(),
			Fitness : 0,
		}
	}
	if g.Hw3ex3 {
		g.ConfirmedMessageCh<- tlc
	}
	wrappedMessage = &message.WrappedRumorTLCMessage{
		TLCMessage : tlc,
	}

	// Store new tlc message into rumor buffer
	g.RumorBuffer.Rumors[g.Name] = append(g.RumorBuffer.Rumors[g.Name], &message.WrappedRumorTLCMessage{
		TLCMessage : tlc,
	})
	g.RumorBuffer.Mux.Unlock()

	// Update status
	g.StatusBuffer.Mux.Lock()
	if _, ok := g.StatusBuffer.Status[g.Name]; !ok {
		g.StatusBuffer.Status[g.Name] = 2
	} else {
		g.StatusBuffer.Status[g.Name] += 1
	}
	g.StatusBuffer.Mux.Unlock()

	// Print confirmation from itself
	fmt.Printf("RECEIVE MAJORITY ACK FOR %d th PROPOSAL\n", ID)
	selfOutStr := fmt.Sprintf("CONFIRMED GOSSIP origin %s ID %d file name %s size %d metahash %s\n",
		tlc.Origin,
		tlc.ID,
		tlc.TxBlock.Transaction.Name,
		tlc.TxBlock.Transaction.Size,
		hex.EncodeToString(tlc.TxBlock.Transaction.MetafileHash),
	)
	fmt.Printf(selfOutStr)
	if g.Hw3ex2 || g.Hw3ex3 {
		g.MsgBuffer.Mux.Lock()
		g.MsgBuffer.Msg = append(g.MsgBuffer.Msg, selfOutStr)
		g.MsgBuffer.Mux.Unlock()
	}
	g.MongerRumor(wrappedMessage, "", []string{})
}


func (g *Gossiper) HandleTLCAck() {
	// Step 0. Initialize map holding ack info
	// Step 1. Get ack from TLCAckChs
	// Step 2. Update local ack map
	// Step 3. Trigger termination if some proposal obtain majority ack

	/* Step 0 */
	ackMap := make(map[uint32]map[string]bool)
	finishMap := make(map[uint32]bool)

	/* Step 1 */
	for wrappedPkt := range g.TLCAckCh {
		
		peer := wrappedPkt.Packet.ACK.Origin
		ID := wrappedPkt.Packet.ACK.ID

		// Stop handling if this round has already finished
		if _, ok := finishMap[ID]; ok {
			continue
		}

		fmt.Printf("RECEIVE ACK OF %d FROM %s\n", ID, peer)
		/* Step 2 */
		if _, ok := ackMap[ID]; !ok {
			ackMap[ID] = make(map[string]bool)

			// ACK SELF
			ackMap[ID][g.Name] = true
		}

		if _, ok := ackMap[ID][peer]; !ok {
			ackMap[ID][peer] = true
		} else {
			continue
		}

		/* Step 3 */
		g.TLCAckChs.Mux.Lock()
		terminateCh := g.TLCAckChs.Chs[ID]
		g.TLCAckChs.Mux.Unlock()

		// Put the witness to the channel
		witnesses := make([]string, 0)
		for k, _ := range ackMap[ID] {
			witnesses = append(witnesses, k)
		}
		terminateCh<- witnesses

		if len(ackMap[ID]) >= int(math.Ceil(float64(g.NumPeers) / 2)) {
			g.TLCAckChs.Mux.Lock()
			delete(g.TLCAckChs.Chs, ID)
			g.TLCAckChs.Mux.Unlock()
			close(terminateCh)
			finishMap[ID] = true
		} 
	}
}

func (g *Gossiper) RoundTLCAck() {
	// This function ack proposals in current round. Used only when hw3ex3 flag set
	// Step 1. Initialize local cache of proposals
	// Step 2. When enter a new round, ack all the proposals of current round

	/* Step 1 */
	tlcCache := make(map[int][]*message.TLCMessage)
	round := 0

	/* Step 2 */
	for {
		select {
		case <-g.TLCRoundCh:
			round += 1
			// Ack the proposals of current round stored in cache
			if cache, ok := tlcCache[round]; ok {
				for _, tlc := range cache {
					fmt.Println("EXIST IMPOSSIBLE ELEMENT IN CACHE")
					g.ACK(tlc.ID, tlc.Origin)
				}
			}
		case wrappedTLC := <-g.WrappedTLCCh:
			// Ack tlc if it is the one required by tlcclock, else store it in cache
			if g.AckAll {
				g.ACK(wrappedTLC.TLCMessage.ID, wrappedTLC.TLCMessage.Origin)
				
			} else {
				origin := wrappedTLC.TLCMessage.Origin
				fmt.Printf("RECEIVE TLC OF ROUND %d\n", wrappedTLC.Round)
				g.TLCClock.Mux.Lock()
				if wrappedTLC.Round >= round && wrappedTLC.Round == g.TLCClock.Clock[origin] - 1{
					g.ACK(wrappedTLC.TLCMessage.ID, wrappedTLC.TLCMessage.Origin)
				} else if wrappedTLC.Round > round {
					// Impossible since the tlc delivery is causal order 
					if _, ok := tlcCache[wrappedTLC.Round]; !ok {
						tlcCache[wrappedTLC.Round] = make([]*message.TLCMessage, 0)
					}
					tlcCache[wrappedTLC.Round] = append(tlcCache[wrappedTLC.Round], wrappedTLC.TLCMessage)
				}
				g.TLCClock.Mux.Unlock()
			}
		}
	}
}


func (g *Gossiper) ComputeRound(tlc *message.TLCMessage) (wrappedTLC *WrappedTLCMessage) {
	// Return the wrappedTLCMessage containing the round of input tlc

	// Get round of input tlc and increment it
	origin := tlc.Origin
	g.TLCClock.Mux.Lock()
	if _, ok := g.TLCClock.Clock[origin]; !ok {
		g.TLCClock.Clock[origin] = 0
		g.TLCClock.Map[origin] = make(map[uint32]int)
	}
	round := g.TLCClock.Clock[origin]
	g.TLCClock.Map[origin][tlc.ID] = round
	g.TLCClock.Clock[origin] += 1
	g.TLCClock.Mux.Unlock()

	// Construct wrappedTLCMessage
	wrappedTLC = &WrappedTLCMessage{
		TLCMessage : tlc,
		Round : round,
	}

	return
}

func (g *Gossiper) HandleConfirmedMessage(confirmCh chan struct{}, sendCh chan struct{}) {
	// Increment round when receives majority confirmed message from current round
	// Step 1. Increment number of confirmed message of current round
	// Step 2. Trigger increment of round if number of confirmed message of current 
	// round becomes greater than N / 2

	/* Step 0 Initialization */
	roundCountMap := make(map[int]int)
	roundFinishMap := make(map[int]bool)
	roundWitnessMap := make(map[int]map[string]int)

	for tlc := range g.ConfirmedMessageCh {

		/* Step 1 */
		// Get round of the tlc message being confirmed
		origin := tlc.Origin
		ID := tlc.Confirmed
		g.TLCClock.Mux.Lock()
		round := g.TLCClock.Map[origin][uint32(ID)]
		g.TLCClock.Mux.Unlock()
		fmt.Printf("RECEIVE CONFIRMED MSG FROM %s IN ROUND %d\n", origin, round)
		// Increment roundCount if round has not finished
		if _, ok := roundFinishMap[round]; !ok {
			if _, roundOk := roundCountMap[round]; !roundOk {
				roundCountMap[round] = 0
				roundWitnessMap[round] = make(map[string]int)
			}
			roundCountMap[round] += 1
			roundWitnessMap[round][origin] = ID
			/* Step 2 */
			if roundCountMap[round] >= int(math.Ceil(float64(g.NumPeers) / 2)) {
				<-sendCh
				roundFinishMap[round] = true
				g.Round += 1
				g.TLCRoundCh<- struct{}{}
				confirmCh<- struct{}{}
				witnessOutputSlice := make([]string, roundCountMap[round])
				index := 0
				for k, v := range roundWitnessMap[round] {
					witnessOutputSlice[index] = fmt.Sprintf("origin%d %s ID%d %d", index + 1, k, index + 1, v)
					index += 1
				}
				outputStr := fmt.Sprintf("ADVANCING TO round %d BASED ON CONFIRMED MESSAGES %s",
											g.Round, strings.Join(witnessOutputSlice, ","))
				outputStr += "\n"
				g.MsgBuffer.Mux.Lock()
				g.MsgBuffer.Msg = append(g.MsgBuffer.Msg, outputStr)
				g.MsgBuffer.Mux.Unlock()
				fmt.Printf(outputStr)
			}
		}
	}
}

func (g *Gossiper) HandleRoundSend(confirmCh chan struct{}, sendCh chan struct{}) {
	// Send msg in TLC round order

	round := 0
	for tx := range g.TransactionSendCh {
		go g.SendTLC(*tx, round)
		round += 1
		sendCh<- struct{}{}
		<-confirmCh
	}
}