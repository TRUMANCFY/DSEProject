package message

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net"
)

// An Election contains all the information about a Helios election.
type Election struct {
	// JSON stores the original JSON for the election. This is not part of
	// the Helios JSON structure but is added here for convenience.
	JSON []byte `json:"-"`

	// ElectionHash stores the SHA256 hash of the JSON value, since this is
	// needed to verify each ballot. This is not part of the original
	// Helios JSON structure but is added here for convenience.
	ElectionHash string `json:"-"`

	// CastURL is the url that can be used to cast ballots; casting ballots
	// is not currently supported by this go package. Ballots must still be
	// cast using the online Helios service.
	CastURL string `json:"cast_url"`

	// Description is a plaintext description of the election.
	Description string `json:"description"`

	// FrozenAt is the date at which the election was fully specified and
	// frozen.
	FrozenAt string `json:"frozen_at"`

	// Name is the full name of the election.
	Name string `json:"name"`

	// Openreg specifies whether or not voters can be added after the
	// election has started.
	Openreg bool `json:"openreg"`

	// PublicKey is the ElGamal public key associated with the election.
	// This is the key used to encrypt all ballots and to create and verify
	// proofs.
	PublicKey *Key `json:"public_key"`

	// Questions is the list of questions to be voted on in this election.
	Questions []*Question `json:"questions"`

	// ShortName provides a short plaintext name for this election.
	ShortName string `json:"short_name"`

	// UseVoterAliases specifies whether or not voter names are replaced by
	// alises (like V153) that leak no information about the voter
	// identities. This can be used instead of encrypting voter names if the
	// election creators want to be sure that voter identities will remain
	// secret forever, even in the face of future cryptanalytic advances.
	UseVoterAliases bool `json:"use_voter_aliases"`

	// Uuid is a unique identifier for this election. This uuid is used in
	// the URL of the election itself: the URL of the JSON version of this
	// Election data structure is
	// https://vote.heliosvoting.org/helios/elections/<uuid>
	Uuid string `json:"uuid"`

	// VotersHash provides the hash of the list of voters.
	VotersHash string `json:"voters_hash"`

	VotingEndsAt   string `json:"voting_ends_at"`
	VotingStartsAt string `json:"voting_starts_at"`

	Secret *big.Int

	Trustees []*Trustee `json:"trustee"`
}

// A Question is part of an Election and specifies a question to be voted on.
type Question struct {
	// AnswerUrls can provide urls with information about answers. These
	// urls can be empty.
	AnswerUrls []string `json:"answer_urls"`

	// Answers is the list of answer choices for this question.
	Answers []string `json:"answers"`

	// ChoiceType specifies the possible ways to evaluate responses. It can
	// currently only be set to 'approval'.
	ChoiceType string `json:"choice_type"`

	// Maximum specifies the maximum value of a vote for this Question. If
	// Max is not specified in the JSON structure, then there will be no
	// OverallProof, since any number of values is possible, up to the
	// total number of answers. This can be detected by looking at
	// OverallProof in the given Ballot.
	Max int `json:"max"`

	// Min specifies the minimum number of answers. This can be as low as
	// 0.
	Min int `json:"min"`

	// Question gives the actual question to answer
	Question string `json:"question"`

	// ResultType specifies the way in which results should be calculated:
	// 'absolute' or 'relative'.
	ResultType string `json:"result_type"`

	// ShortName gives a short representation of the Question.
	ShortName string `json:"short_name"`

	// TallyType specifies the kind of tally to perform. The only valid
	// value here is 'homomorphic'.
	TallyType string `json:"tally_type"`
}

/* Struct definition */
type Message struct {

	// TODO: Figure out why need ptr here
	Text        string
	Destination *string
	File        *string
	Request     *[]byte
	Keywords    []string
	Budget      uint64

	// Attributes for blockchain
	Voterid string
	Vote    string
}

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

type SearchRequestRelayer struct {
	SearchRequest *SearchRequest
	Relayer       string
}
type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
	ChunkCount   uint64
}

type TxPublish struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

type BlockPublish struct {
	PrevHash    [32]byte
	Transaction TxPublish
}

type TLCMessage struct {
	Origin      string
	ID          uint32
	Confirmed   int
	TxBlock     BlockPublish
	VectorClock *StatusPacket
	Fitness     float32
}

type WrappedRumorTLCMessage struct {
	RumorMessage      *RumorMessage
	TLCMessage        *TLCMessage
	BlockRumorMessage *BlockRumorMessage
}

// A Trustee represents the public information for one of the keys used to
// tally and decrypt the election results.
type Trustee struct {
	// DecryptionFactors are the partial decryptions of each of the
	// homomorphic tally results.
	DecryptionFactors [][]*big.Int `json:"decryption_factors"`

	// DecryptionProofs are the proofs of correct partial decryption for
	// each of the DecryptionFactors.
	//DecryptionProofs [][]*ZKProof `json:"decryption_proofs"`

	// PoK is a proof of knowledge of the private key share held by this
	// Trustee and used to create the DecryptionFactors.
	//PoK *SchnorrProof `json:"pok"`

	// PublicKey is the ElGamal public key of this Trustee.
	PublicKey *Key `json:"public_key"`

	// PublicKeyHash is the SHA-256 hash of the JSON representation of
	// PublicKey.
	PublicKeyHash string `json:"public_key_hash"`

	// Uuid is the unique identifier for this Trustee.
	Uuid string `json:"uuid"`

	// Address of the trustee
	Address string `json:"address"`

	// Name of election
	Election string `json:"election"`
}

// A Key is an ElGamal public key. There is one Key in each Election, and it
// specifies the group in which computations are to be performed. Encryption of
// a value m is performed as (g^r, g^m * y^r) mod p.
type Key struct {
	// Generator is the generator element g used in ElGamal encryptions.
	Generator *big.Int `json:"g"`

	// Prime is the prime p for the group used in encryption.
	Prime *big.Int `json:"p"`

	// ExponentPrime is another prime that specifies the group of exponent
	// values in the exponent of Generator. It is used in challenge
	// generation and verification.
	ExponentPrime *big.Int `json:"q"`

	// PublicValue is the public-key value y used to encrypt.
	PublicValue *big.Int `json:"y"`
}

/************************ Message for blockchain ************************/
type CastBallot struct {
	JSON []byte `json:"-"`
	// CastAt gives the time at which Vote was cast.
	CastAt string `json:"cast_at"`

	// Vote is the cast Ballot itself.
	Vote *Ballot `json:"vote"`

	// VoteHash is the SHA-256 hash of the JSON corresponding to Vote.
	VoteHash string `json:"vote_hash"`

	// VoterHash is the SHA-256 hash of the Voter JSON corresponding to
	// VoterUuid.
	VoterHash string `json:"voter_hash"`

	// VoterUuid is the unique identifier for the Voter that cast Vote.
	VoterUuid string `json:"voter_uuid"`
}

func (cb *CastBallot) BigInt2Str() {
	/* This func convert bigint in ballot to string */

	for _, answer := range cb.Vote.Answers {
		// Convert all big int to string
		answer.ChoicesStr = make([]*CiphertextStr, len(answer.Choices))
		answer.RandomnessStr = make([]*string, len(answer.Randomness))

		for i, choice := range answer.Choices {
			answer.ChoicesStr[i] = NewCiphertextStr(choice.Alpha, choice.Beta)
		}
		for i, r := range answer.Randomness {
			// Convert all big int to string in randomness
			result := r.String()
			answer.RandomnessStr[i] = &result
		}

		// Remove all big int pointers
		answer.Choices = make([]*Ciphertext, 0)
		answer.Randomness = make([]*big.Int, 0)
	}
}

func (cb *CastBallot) Str2BigInt() {
	/* This func convert string to big int */

	for _, answer := range cb.Vote.Answers {
		// Convert all string to big int
		answer.Choices = make([]*Ciphertext, len(answer.ChoicesStr))
		answer.Randomness = make([]*big.Int, len(answer.RandomnessStr))

		for i, choiceStr := range answer.ChoicesStr {
			answer.Choices[i] = NewCiphertext(choiceStr.Alpha, choiceStr.Beta)
		}
		for i, r := range answer.RandomnessStr {
			result := new(big.Int)
			result, ok := result.SetString(*r, 10)
			if !ok {
				fmt.Println("Cannot convert randomness str to big int")
				return
			}
			answer.Randomness[i] = result
		}
		// Remove all string pointers
		answer.ChoicesStr = make([]*CiphertextStr, 0)
		answer.RandomnessStr = make([]*string, 0)
	}

	return
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
	Choices    []*Ciphertext `json:"choices"`
	ChoicesStr []*CiphertextStr

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
	Randomness    []*big.Int `json:"randomness,omitempty"`
	RandomnessStr []*string
}

// A Ciphertext is an ElGamal ciphertext, where g is Key.Generator, r is a
// random value, m is a message, and y is Key.PublicValue.
type Ciphertext struct {
	// Alpha = g^r
	Alpha *big.Int `json:"alpha"`

	// Beta = g^m * y^r
	Beta *big.Int `json:"beta"`
}

type CiphertextStr struct {
	Alpha *string
	Beta  *string
}

func NewCiphertextStr(alpha *big.Int, beta *big.Int) (cs *CiphertextStr) {

	alphaStr := alpha.String()
	betaStr := beta.String()
	cs = &CiphertextStr{
		Alpha: &alphaStr,
		Beta:  &betaStr,
	}
	return
}

func NewCiphertext(alpha, beta *string) (ct *Ciphertext) {

	alphaBigInt := new(big.Int)
	betaBigInt := new(big.Int)

	alphaBigInt, ok := alphaBigInt.SetString(*alpha, 10)
	if !ok {
		fmt.Println("Convert alpha string to big int err")
		return
	}
	betaBigInt, ok = betaBigInt.SetString(*beta, 10)
	if !ok {
		fmt.Println("Convert beta string to big int err")
		return
	}

	ct = &Ciphertext{
		Alpha: alphaBigInt,
		Beta:  betaBigInt,
	}

	return
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
	ID     uint32
	Block  *Block
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
	Simple            *SimpleMessage
	Rumor             *RumorMessage
	Status            *StatusPacket
	Private           *PrivateMessage
	DataRequest       *DataRequest
	DataReply         *DataReply
	SearchRequest     *SearchRequest
	SearchReply       *SearchReply
	TLCMessage        *TLCMessage
	ACK               *TLCAck
	BlockRumorMessage *BlockRumorMessage
}

type Gossiper struct {
	address *net.UDPAddr
	conn    *net.UDPConn
	Name    string
}

type PacketToSend struct {
	Packet  *GossipPacket
	Addr    string
	Timeout chan struct{}
}

type PacketIncome struct {
	Packet *GossipPacket
	Sender string
}

type ClientMsgIncome struct {
	Msg    *Message
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

// Tally computes the tally of an election and returns the result.
// In the process, it generates partial decryption proofs for each of
// the partial decryptions computed by the trustee.
func (e *Election) Tally(votes []*CastBallot, t *Trustee, trusteeSecrets *big.Int) {
	tallies, _ := e.AccumulateTallies(votes)
	// TODO(tmroeder): maybe we should just skip votes that don't pass verification?
	// What does the spec say?
	//if len(voteFingerprints) == 0 {
	//  glog.Error("Couldn't tally the votes")
	//  return nil, errors.New("couldn't tally the votes")
	//}

	df := make([][]*big.Int, len(e.Questions))
	//dp := make([][]*ZKProof, len(e.Questions))
	for i, q := range e.Questions {
		df[i] = make([]*big.Int, len(q.Answers))
		//dp[i] = make([]*ZKProof, len(q.Answers))
		for j := range q.Answers {
			df[i][j] = new(big.Int).Exp(tallies[i][j].Alpha, trusteeSecrets, t.PublicKey.Prime)
			//if dp[i][j], err = NewPartialDecryptionProof(tallies[i][j], df[i][j], trusteeSecrets[k], t.PublicKey); err != nil {
			//  glog.Errorf("Couldn't create a proof for (%d, %d) for trustee %d\n", i, j, k)
			//  return nil, err
			//}
		}
	}

	t.DecryptionFactors = df
	//t.DecryptionProofs = dp
}

func (election *Election) AccumulateTallies(votes []*CastBallot) ([][]*Ciphertext, []string) {
	// Initialize the tally structures for homomorphic accumulation.

	tallies := make([][]*Ciphertext, len(election.Questions))
	fingerprints := make([]string, len(votes))
	for i := range tallies {
		tallies[i] = make([]*Ciphertext, len(election.Questions[i].Answers))
		for j := range tallies[i] {
			// Each tally must start at 1 for the multiplicative
			// homomorphism to work.
			tallies[i][j] = &Ciphertext{big.NewInt(1), big.NewInt(1)}
		}
	}

	// Verify the votes and accumulate the tallies.
	//resp := make(chan bool)
	for i := range votes {
		// Shadow i as a new variable for the goroutine.
		i := i
		//go func(c chan bool) {
		//	glog.Infof("Verifying vote from %s\n", votes[i].VoterUuid)
		//	c <- votes[i].Vote.Verify(election)
		//	return
		//}(resp)

		h := sha256.Sum256(votes[i].JSON)
		encodedHash := base64.StdEncoding.EncodeToString(h[:])
		fingerprint := encodedHash[:len(encodedHash)-1]
		fingerprints = append(fingerprints, fingerprint)

		for j, q := range election.Questions {
			for k := range q.Answers {
				// tally_j_k = (tally_j_k * ballot_i_j_k) mod p
				tallies[j][k].MulCiphertexts(votes[i].Vote.Answers[j].Choices[k], election.PublicKey.Prime)
			}
		}
	}

	// Make sure all the votes passed verification.
	//for _ = range votes {
	//	if !<-resp {
	//		glog.Error("Vote verification failed")
	//		return nil, nil
	//	}
	//}

	return tallies, fingerprints
}

// MulCiphertexts multiplies an ElGamal Ciphertext value element-wise into an
// existing Ciphertext. This has the effect of adding the value encrypted in the
// other Ciphertext to the prod Ciphertext. The prime specifies the group in
// which these multiplication operations are to be performed.
func (prod *Ciphertext) MulCiphertexts(other *Ciphertext, prime *big.Int) *Ciphertext {
	prod.Alpha.Mul(prod.Alpha, other.Alpha)
	prod.Alpha.Mod(prod.Alpha, prime)
	prod.Beta.Mul(prod.Beta, other.Beta)
	prod.Beta.Mod(prod.Beta, prime)
	return prod
}

type Result [][]int64

func (e *Election) Tallier(votes []*CastBallot, trustees []*Trustee) (Result, error) {
	tallies, _ := e.AccumulateTallies(votes)

	// For each question and each answer, reassemble the tally and search for its value.
	// Then put this in the results.
	maxValue := len(votes)
	result := make([][]int64, len(e.Questions))
	for i, q := range e.Questions {
		result[i] = make([]int64, len(q.Answers))
		for j := range q.Answers {
			alpha := big.NewInt(1)
			for k := range trustees {
				alpha.Mul(alpha, trustees[k].DecryptionFactors[i][j])
				alpha.Mod(alpha, trustees[k].PublicKey.Prime)
			}

			beta := new(big.Int).ModInverse(alpha, e.PublicKey.Prime)
			beta.Mul(beta, tallies[i][j].Beta)
			beta.Mod(beta, e.PublicKey.Prime)

			// This decrypted value can be anything between g^0 and g^maxValue.
			// Try all values until we find it.
			temp := new(big.Int)
			val := new(big.Int)
			var v int
			for v = 0; v <= maxValue; v++ {
				val.SetInt64(int64(v))
				temp.Exp(e.PublicKey.Generator, val, e.PublicKey.Prime)
				if temp.Cmp(beta) == 0 {
					result[i][j] = int64(v)
					break
				}
			}

			if v > maxValue {
				fmt.Printf("Couldn't decrypt value (%d, %d)\n", i, j)
				return nil, errors.New("couldn't decrypt part of the tally")
			}
		}
	}

	return result, nil
}
