package gossiper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/TRUMANCFY/DSEProject/Peerster/message"
)

type Blockchain struct {

	// Blocks
	Blocks   []*message.Block
	BlockMux sync.Mutex
	// Number of Peers
	N int

	// Next index of block to be added
	NextId int

	// Buffer of ballot to be added into blockchain
	Buffer []*message.CastBallot

	// Input channel for buffer
	InputCh chan *message.CastBallot

	// Send channel for candidate blocks
	SendCh chan *message.Block

	// Receive channel for candidate blocks
	ReceiveCh chan *message.Block

	// Mutexes for concurrent access of channels and buffers
	BufferMux sync.Mutex

	// Origin
	Origin string

	// Round map
	Map    map[string]map[int]bool
	MapMux sync.Mutex

	// Voter map
	VoterMap    map[string]bool
	VoterMapMux sync.Mutex

	// Election Name
	ElectionName string

	// Election prefix for testing purpose
	Prefix string

	// String record of blocks
	Records []string
}

func Rand() uint64 {
	buf := make([]byte, 8)
	rand.Read(buf) // Always succeeds, no need to check error
	return binary.LittleEndian.Uint64(buf)
}

func (g *Gossiper) NewBlockchain(electionName string) (bc *Blockchain) {
	/*
		This func create an instance of blockchain with genesis block
	*/
	// Create the channel
	bc = &Blockchain{
		Blocks:       make([]*message.Block, 0),
		NextId:       0,
		Buffer:       make([]*message.CastBallot, 0),
		InputCh:      make(chan *message.CastBallot, 0),
		SendCh:       make(chan *message.Block, 0),
		ReceiveCh:    make(chan *message.Block, 0),
		N:            g.NumPeers,
		Origin:       g.Name,
		Map:          make(map[string]map[int]bool),
		VoterMap:     make(map[string]bool),
		ElectionName: electionName,
		Records : make([]string, 0),
	}

	// For testing purpose
	if electionName == "China" {
		bc.Prefix = "\t\t"
	} else {
		bc.Prefix = "\t"
	}
	// Add genesis block
	genesisBlock := &message.Block{
		PrevHash:     sha256.Sum256(make([]byte, 0)),
		CurrentHash:  sha256.Sum256(make([]byte, 0)),
		CastBallot:   nil,
		ElectionName: bc.ElectionName,
	}
	bc.BlockMux.Lock()
	bc.Blocks = append(bc.Blocks, genesisBlock)
	bc.BlockMux.Unlock()
	bc.NextId = 1

	// Set random seed
	seedHash := sha256.Sum256([]byte(g.Name))
	seed := binary.BigEndian.Uint64(seedHash[:])
	rand.Seed(int64(seed))
	// Start working
	go bc.HandleRound()
	return
}

func (bc *Blockchain) CheckBlockValidty(b *message.Block) bool {
	/* This func returns true if the block's prevhash is the same as
	the end of current blockchain's hash */

	return bytes.Compare(b.PrevHash[:], bc.Blocks[len(bc.Blocks)-1].CurrentHash[:]) == 0
}

func (bc *Blockchain) HandleRound() {
	/*
		This function handle rounds of adding blocks into the blockchain
	*/
	voters := make(map[string]bool)
	for {
		bc.BufferMux.Lock()
		if len(bc.Buffer) > 0 {
			// Get the first Vote that has not been recorded in the blockchain to propagate

			var currentVote *message.CastBallot
			nextBufferIndex := 0
			var valid bool
			// Check whether there is valid vote to propogate
			for _, currentVote = range bc.Buffer {
				valid = true
				if _, ok := voters[currentVote.VoterUuid]; ok {
					valid = false
				}
				if valid {
					break
				} else {
					nextBufferIndex += 1
				}
			}
			if nextBufferIndex > len(bc.Buffer) {
				bc.Buffer = bc.Buffer[0:0]
			} else {
				bc.Buffer = bc.Buffer[nextBufferIndex:]
			}
			if !valid {
				// No valid block vote to propose
				bc.BufferMux.Unlock()
				time.Sleep(50 * time.Millisecond)
				continue
			}
			bc.BufferMux.Unlock()

			fmt.Printf("%s THE RECORD TO BE PROPOSED HAS VOTERID %s\n", bc.Prefix, currentVote.VoterUuid)
			// Create the block
			currentBlock := &message.Block{
				CastBallot:   currentVote,
				PrevHash:     bc.Blocks[len(bc.Blocks)-1].CurrentHash,
				Fitness:      Rand(),
				Round:        bc.NextId,
				Origin:       bc.Origin,
				ElectionName: bc.ElectionName,
			}
			currentBlock.CurrentHash = currentBlock.Hash()

			// Ask the gossiper to send the block
			bc.SendCh <- currentBlock

			// Wait for all peers' proposals
			count := 1
			// fmt.Printf("OUR FITNESS IS %d\n", currentBlock.Fitness)
			receivedMap := make(map[string]bool)
			for {
				// Check validity of proposal, here we actually don't need to as
				// all peers are trusted
				// Update self's block if peer's block has higher fitness value
				peerBlock := <-bc.ReceiveCh
				if _, ok := receivedMap[peerBlock.Origin]; !ok {
					receivedMap[peerBlock.Origin] = true
				} else {
					continue
				}
				// fmt.Printf("Peer fitness is %d for round %d\n", peerBlock.Fitness, peerBlock.Round)
				if bc.CheckBlockValidty(peerBlock) && peerBlock.Fitness > currentBlock.Fitness {
					currentBlock = peerBlock
				}
				count += 1
				fmt.Printf("%s RECEIVED %d proposals\n", bc.Prefix, count)
				if count == bc.N {
					break
				}
			}

			// Add the consensus block to the blockchain
			if _, ok := voters[currentBlock.CastBallot.VoterUuid]; ok {
				continue
			} else {
				voters[currentBlock.CastBallot.VoterUuid] = true
			}

			bc.BlockMux.Lock()
			bc.Blocks = append(bc.Blocks, currentBlock)
			bc.BlockMux.Unlock()

			fmt.Printf("%s    APPENDING BLOCK WITH VOTER UID %s, VOTE HASH %s FOR ELECTION %s\n",
				bc.Prefix,
				currentBlock.CastBallot.VoterUuid,
				currentBlock.CastBallot.VoteHash,
				currentBlock.ElectionName)
			bc.NextId += 1
			fmt.Printf("%s ENTERING ROUND %d FOR ELECTION %s\n\n", bc.Prefix, bc.NextId, bc.ElectionName)
		} else {
			bc.BufferMux.Unlock()
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (bc *Blockchain) CreateBallot(voterid, vote, electionName string) (v *message.CastBallot) {
	/* This function create a ballot from the voterid and vote */

	voteHash := sha256.Sum256([]byte(vote))
	voterHash := sha256.Sum256([]byte(voterid))
	voteHashStr := hex.EncodeToString(voteHash[:])
	voterHashStr := hex.EncodeToString(voterHash[:])
	v = &message.CastBallot{
		VoteHash:  voteHashStr,
		VoterHash: voterHashStr,
		VoterUuid: voterid,
		Vote: &message.Ballot{
			ElectionUuid: electionName,
		},
	}

	return
}

func (g *Gossiper) GetOrCreateBlockchain(electionName string) (bc *Blockchain) {
	/*
		This function get or create the blockchain corresponding to the election name
	*/

	g.BlockchainsMux.Lock()
	if _, ok := g.Blockchains[electionName]; !ok {
		bc = g.NewBlockchain(electionName)
		go g.HandleSendingBlocks(bc.SendCh)
		g.Blockchains[electionName] = bc
	}
	bc = g.Blockchains[electionName]
	g.BlockchainsMux.Unlock()
	return
}
func (g *Gossiper) HandleSendingBlocks(sendCh chan *message.Block) {
	/*
		This func receives blocks from underlying blockchain layer
		and send it using gossiper's rumor mongering
	*/

	for block := range sendCh {
		fmt.Println("SENDING SENDING SENDING")
		// Construct msg to be sent
		g.RumorBuffer.Mux.Lock()
		fmt.Println("OBTAIN RUMOR LOCK")
		wrappedMessage := &message.WrappedRumorTLCMessage{
			BlockRumorMessage: &message.BlockRumorMessage{
				Origin: g.Name,
				ID:     uint32(len(g.RumorBuffer.Rumors[g.Name]) + 1),
				Block:  block,
				Proof: g.Auth.Provide(),
			},
		}

		// Store msg
		g.RumorBuffer.Rumors[g.Name] = append(g.RumorBuffer.Rumors[g.Name], wrappedMessage)
		g.RumorBuffer.Mux.Unlock()
		// Update status
		g.StatusBuffer.Mux.Lock()
		fmt.Println("OBTAIN STATUS LOCK")
		if _, ok := g.StatusBuffer.Status[g.Name]; !ok {

			g.StatusBuffer.Status[g.Name] = 2
		} else {

			g.StatusBuffer.Status[g.Name] += 1
		}
		g.StatusBuffer.Mux.Unlock()

		// Monger block
		var prefix string
		if block.ElectionName == "China" {
			prefix = "\t\t"
		} else {
			prefix = "\t"
		}
		fmt.Printf("%s PROPOSING BLOCK WITH VOTER %s VOTE %s IN ROUND %d FOR ELECTION %s\n",
			prefix,
			block.CastBallot.VoterUuid,
			block.CastBallot.VoteHash,
			block.Round,
			block.ElectionName)
		g.MongerRumor(wrappedMessage, "", []string{})
	}
}

func (g *Gossiper) HandleReceivingBlock(wrapped_pkt *message.PacketIncome) {
	/*
		This func receive blocks from communication layer
		and inform blockchain layer with the right election name
		Step 0. Check validty of the block by authenticate the origin
		Step 1. Add the vote to corresponding blockchain buffer if it is empty
		Step 2. Inform the blockchain of the vote
		Step 3. Monger the block if necessary
	*/

	sender, blockRumor := wrapped_pkt.Sender, wrapped_pkt.Packet.BlockRumorMessage

	/* Step 0 */
	valid := g.Auth.Verify(blockRumor.Proof)
	if !valid {
		return
	}
	/* Step 1 */
	b := blockRumor.Block
	// Get or Create the corresponding blockchain
	bc := g.GetOrCreateBlockchain(b.ElectionName)

	// prefix := ""
	// if b.ElectionName == "China" {
	// 	prefix = "\t\t"
	// } else {
	// 	prefix = "\t"
	// }
	// fmt.Printf("%s NETWORK RECEVING BLOCK VOTER %s VOTING %s IN ELECTION %s FOR ROUND %d FROM PEER %s\n",
	// 	prefix,
	// 	b.CastBallot.VoterUuid,
	// 	b.CastBallot.VoteHash,
	// 	b.ElectionName,
	// 	b.Round,
	// 	b.Origin)

	// Reject block from future
	if b.Round > bc.NextId {
		return
	}

	// Reject block from self
	peerOrigin := blockRumor.Origin
	if peerOrigin == g.Name {
		return
	}

	// Check whether block has been seen before
	updated := g.Update(&message.WrappedRumorTLCMessage{
		BlockRumorMessage: blockRumor,
	}, sender)

	defer g.N.Send(&message.GossipPacket{
		Status: g.StatusBuffer.ToStatusPacket(),
	}, sender)

	if updated {

		// Monger it to peer if the block is in last round or this round
		if b.Round >= bc.NextId-1 {
			wrappedMessage := &message.WrappedRumorTLCMessage{
				BlockRumorMessage: blockRumor,
			}
			g.MongerRumor(wrappedMessage, "", []string{sender})
		}

		// Not update blockchain buffer if the block is not for current round
		if b.Round != bc.NextId {
			return
		}

		var prefix string
		if b.ElectionName == "China" {
			prefix = "\t\t"
		} else {
			prefix = "\t"
		}

		fmt.Printf("%s ACCEPT RECEVING BLOCK VOTER %s VOTING %s IN ELECTION %s FOR ROUND %d FROM PEER %s\n",
			prefix,
			b.CastBallot.VoterUuid,
			b.CastBallot.VoteHash,
			b.ElectionName,
			b.Round,
			b.Origin)

		// Check whether the record for the voter already existed in the blockchain
		bc.VoterMapMux.Lock()
		var existed bool
		if _, ok := bc.VoterMap[b.CastBallot.VoterUuid]; !ok {
			bc.VoterMap[b.CastBallot.VoterUuid] = true
			existed = false
		} else {
			existed = true
		}
		bc.VoterMapMux.Unlock()

		// Add it to buffer if not existed
		bc.BufferMux.Lock()
		if !existed {
			bc.Buffer = append(bc.Buffer, b.CastBallot)
		}
		bc.BufferMux.Unlock()

		// Step 2
		bc.ReceiveCh <- b
	}

	return
}

func (g *Gossiper) HandleReceivingVote(v *message.CastBallot) {
	/*
		This func add the vote to the corresponding blockchain's buffer
		Step 0. Convert big int to string in cast ballot
		Step 1. Get or Create the corresponding blockchain
		Step 2. Add the vote to the blockchain's buffer
	*/

	/* Step 0 */
	v.BigInt2Str()

	/* Step 1 */
	electionName := v.Vote.ElectionUuid
	bc := g.GetOrCreateBlockchain(electionName)

	/* Step 2 */
	bc.BufferMux.Lock()
	fmt.Printf("%s BUFFERING VOTER %s\n", bc.ElectionName, v.VoterUuid)
	bc.Buffer = append(bc.Buffer, v)
	bc.BufferMux.Unlock()
	return
}

func (bc *Blockchain) GetCastBallots() (castBallots []*message.CastBallot) {
	/*
		This func returns a slice of pointer to cast ballots
		The string representation of big.Int in cast ballots are converted back to big.Int
	*/

	castBallots = make([]*message.CastBallot, bc.NextId-1)
	bc.BlockMux.Lock()
	for i := 1; i < len(bc.Blocks); i += 1 {

		castBallots[i-1] = bc.Blocks[i].CastBallot
	}
	for _, cb := range castBallots {
		cb.Str2BigInt()
	}
	bc.BlockMux.Unlock()

	return
}

func (g *Gossiper) TerminateElection() {
	/*
		This function is triggered after receiving terminating signal for an election
		Step 1. Return the castBallots
		Step 2. Terminate the channels
		Step 3. Terminate blockchain
	*/
}
