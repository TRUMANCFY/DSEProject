package message

import (

	"net"
	"math/big"
	"crypto/sha256"
)
/* Struct definition */
type Message struct {

	// TODO: Figure out why need ptr here
	Text string
	Destination *string
	File *string
	Request *[]byte
	Keywords []string
	Budget uint64

	// Attributes for blockchain
	Voterid string
	Vote string
}

type SimpleMessage struct {

	OriginalName string
	RelayPeerAddr string
	Contents string
}

type RumorMessage struct {

	Origin string
	ID uint32
	Text string
}

type PeerStatus struct {

	Identifier string
	NextID uint32
}

type StatusPacket struct {

	Want []PeerStatus
}

type  PrivateMessage struct {

	Origin string
	ID uint32
	Text string
	Destination string
	HopLimit uint32
}

type DataRequest struct {

	Origin string
	Destination string
	HopLimit uint32
	HashValue []byte
}

type DataReply struct {

	Origin string
	Destination string
	HopLimit uint32
	HashValue []byte
	Data []byte
}

type SearchRequest struct {

	Origin string
	Budget uint64
	Keywords []string
}

type SearchRequestRelayer struct {

	SearchRequest *SearchRequest
	Relayer string
}
type SearchReply struct {

	Origin string
	Destination string
	HopLimit uint32
	Results []*SearchResult
}

type SearchResult struct {

	FileName string
	MetafileHash []byte
	ChunkMap []uint64
	ChunkCount uint64
}

type TxPublish struct {
	Name string
	Size int64
	MetafileHash []byte
}

type BlockPublish struct {
	PrevHash [32]byte
	Transaction TxPublish
}

type TLCMessage struct {
	Origin string
	ID uint32
	Confirmed int
	TxBlock BlockPublish
	VectorClock *StatusPacket
	Fitness float32
}

type WrappedRumorTLCMessage struct {
	RumorMessage *RumorMessage
	TLCMessage *TLCMessage
	BlockRumorMessage *BlockRumorMessage
}

/************************ Message for blockchain ************************/
type CastBallot struct {

	// CasAt gives the time of the vote
	CastAt string

	// Vote is the cast Ballot itself
	Vote *Ballot

	// Vote Hash 
	VoteHash string

	// VoterHash is the hash of the voter uuid	
	VoterHash string

	// VoterUuid is the unique identifier of the voter
	VoterUuid string
}

// A Ballot is a cryptographic vote in an Election.
type Ballot struct {
	// Answers is a list of answers to the Election specified by
	// ElectionUuid and ElectionHash.
	Answers []*EncryptedAnswer `json:"answers"`

	// ElectionHash is the SHA-256 hash of the Election specified by
	// ElectionUuid.
	ElectionHash string `json:"election_hash"`

	// ElectionUuid is the unique identifier for the Election that Answers
	// apply to.
	ElectionUuid string `json:"election_uuid"`
}

type EncryptedAnswer struct {
	// Choices is a list of votes for each choice in a Question. Each choice
	// is encrypted with the Election.PublicKey.
	Choices []*Ciphertext `json:"choices"`

	// IndividualProofs gives a proof that each corresponding entry in
	// Choices is well formed: this means that it is either 0 or 1. So, each
	// DisjunctiveZKProof is a list of two ZKProofs, the first proving the 0
	// case, and the second proving the 1 case. One of these proofs is
	// simulated, and the other is real: see the comment for ZKProof for the
	// algorithm and the explanation.
	//IndividualProofs []DisjunctiveZKProof `json:"individual_proofs"`

	// OverallProof shows that the set of choices sum to an acceptable
	// value: one that falls between Question.Min and Question.Max. If there
	// is no Question.Max, then OverallProof will be empty and does not need
	// to be checked.
	//OverallProof DisjunctiveZKProof `json:"overall_proof"`

	// Answer is the actual answer that is supposed to be encrypted in
	// EncryptedAnswer. This is not serialized/deserialized if not present.
	// This must only be present in a spoiled ballot because SECRECY.
	Answer []int64 `json:"answer,omitempty"`

	// Randomness is the actual randomness that is supposed to have been
	// used to encrypt Answer in EncryptedAnswer. This is not serialized or
	// deserialized if not present. This must only be present in a spoiled
	// ballot because SECRECY.
	Randomness []*big.Int `json:"randomness,omitempty"`
}

// A Ciphertext is an ElGamal ciphertext, where g is Key.Generator, r is a
// random value, m is a message, and y is Key.PublicValue.
type Ciphertext struct {
	// Alpha = g^r
	Alpha *big.Int `json:"alpha"`

	// Beta = g^m * y^r
	Beta *big.Int `json:"beta"`
}

type Block struct {

	// Hash of previous block
	PrevHash [32]byte

	// Hash of current block
	CurrentHash [32]byte

	// Fitness which decide the priority to be added into blockchain
	Fitness uint64

	// Round
	Round int

	// Source 
	Origin string
	
	// Ballot
	CastBallot *CastBallot
}

type BlockRumorMessage struct {
	Origin string
	ID uint32
	Block *Block
}
func (b *Block) Hash() (out [32]byte) {
	/*
	This func provide the hash of block
	*/

	h := sha256.New()
	h.Write(b.PrevHash[:])

	// Hash the ballot data
	referenceString := b.CastBallot.VoteHash + b.CastBallot.VoterHash
	voteHashBytes := sha256.Sum256([]byte(referenceString))
	//voteHashStr := hex.EncodeToString(voteHashBytes[:])

	// Hash current block with prev block's hash
	h.Write(voteHashBytes[:])
	copy(out[:], h.Sum(nil))

	return
}

/***************************************************************************/
func (m *WrappedRumorTLCMessage) GetOrigin() (origin string) {
	if m.RumorMessage != nil {
		origin = m.RumorMessage.Origin
	} else if m.TLCMessage != nil {
		origin = m.TLCMessage.Origin
	} else {
		origin = m.BlockRumorMessage.Origin
	}
	return
}

func (m *WrappedRumorTLCMessage) GetID() (ID uint32) {
	if m.RumorMessage != nil {
		ID = m.RumorMessage.ID
	} else if m.TLCMessage != nil {
		ID = m.TLCMessage.ID
	} else {
		ID = m.BlockRumorMessage.ID
	}
	return
}
type TLCAck PrivateMessage

type GossipPacket struct {

	Simple *SimpleMessage
	Rumor *RumorMessage
	Status *StatusPacket
	Private *PrivateMessage
	DataRequest *DataRequest
	DataReply *DataReply
	SearchRequest *SearchRequest
	SearchReply *SearchReply
	TLCMessage *TLCMessage
	ACK *TLCAck
	BlockRumorMessage *BlockRumorMessage
}

type Gossiper struct {

	address *net.UDPAddr
	conn *net.UDPConn
	Name string
}

type PacketToSend struct {

	Packet *GossipPacket
	Addr string
	Timeout chan struct{}
}

type PacketIncome struct {

	Packet *GossipPacket
	Sender string
}

type ClientMsgIncome struct {

	Msg *Message
	Sender string
}

type StatusMap map[string]uint32 

/* Convert a status packet to map */
func (status *StatusPacket) ToMap() (statusMap StatusMap) {

	statusMap = make(StatusMap)

	for _, peer_status := range status.Want {

		statusMap[peer_status.Identifier] = peer_status.NextID
	}

	return
}