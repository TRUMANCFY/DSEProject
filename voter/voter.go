package voter

import (
	"crypto/dsa"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

type Voter struct{}

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

// NewKeyFromParams uses a given set of parameters to generate a public key.
func (v *Voter) NewKeyFromParams(g *big.Int, p *big.Int, q *big.Int) (*Key, *big.Int, error) {
	secret, err := rand.Int(rand.Reader, q)
	if err != nil {
		fmt.Println("Couldn't generate a secret for the key")
		return nil, nil, err
	}

	return &Key{g, p, q, new(big.Int).Exp(g, secret, p)}, secret, nil

}

// NewKey generates a fresh set of parameters and a public/private key pair in
// those parameters.
func (v *Voter) NewKey() (*Key, *big.Int, error) {
	// Use the DSA crypto code to generate a key pair. For testing
	// purposes, we'll use (2048,224) instead of (2048,160) as used by the
	// current Helios implementation
	params := new(dsa.Parameters)
	if err := dsa.GenerateParameters(params, rand.Reader, dsa.L2048N224); err != nil {
		fmt.Println("Couldn't generate DSA parameters for the ElGamal group")
		return nil, nil, err
	}

	return v.NewKeyFromParams(params.G, params.P, params.Q)
}

// A Ciphertext is an ElGamal ciphertext, where g is Key.Generator, r is a
// random value, m is a message, and y is Key.PublicValue.
type Ciphertext struct {
	// Alpha = g^r
	Alpha *big.Int `json:"alpha"`

	// Beta = g^m * y^r
	Beta *big.Int `json:"beta"`
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

// Encrypt encrypts the selection for an answer: either this value is
// selected or not. It also generates a DisjunctiveZKProof to show that
// the value is either selected or not. It returns the randomness it
// generated; this is useful for computing the OverallProof for a Question.
func (v *Voter) Encrypt(selected bool, pk *Key) (*Ciphertext, *big.Int) {
	// If this value is selected, then use g^1; otherwise, use g^0.
	var plaintext *big.Int
	//var realExp, fakeExp int64
	//if selected {
	//	plaintext = pk.Generator
	//	realExp = 1
	//	fakeExp = 0
	//} else {
	//	plaintext = big.NewInt(1)
	//	realExp = 0
	//	fakeExp = 1
	//}

	randomness, err := rand.Int(rand.Reader, pk.ExponentPrime)
	if err != nil {
		fmt.Println("Couldn't get randomness for an encryption")
		return nil, nil
	}

	a := new(big.Int).Exp(pk.Generator, randomness, pk.Prime)
	b := new(big.Int).Exp(pk.PublicValue, randomness, pk.Prime)
	b.Mul(b, plaintext)
	b.Mod(b, pk.Prime)
	c := &Ciphertext{a, b}

	//// Real proof of selected and a simulated proof of !selected
	//var proof DisjunctiveZKProof
	//proof = make([]*ZKProof, 2)
	//
	//if err = proof.CreateFakeProof(fakeExp, fakeExp, c, pk); err != nil {
	//	glog.Error("Couldn't create a simulated proof")
	//	return nil, nil, nil, err
	//}
	//
	//if err = proof.CreateRealProof(realExp, c, randomness, pk); err != nil {
	//	glog.Error("Couldn't create a real proof")
	//	return nil, nil, nil, err
	//}

	return c, randomness
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

// An EncryptedAnswer is part of a Ballot cast by a Voter. It is the answer to
// a given Question in an Election.
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
}

// NewBallot takes an Election and a set of responses as input and fills in a Ballot
func (v *Voter) NewBallot(election *Election, answers [][]int64) (*Ballot, error) {
	if len(answers) != len(election.Questions) {
		return nil, errors.New("wrong number of answers")
	}

	pk := election.PublicKey

	//vote.ElectionHash = election.ElectionHash
	//vote.ElectionUuid = election.Uuid

	ans := make([]*EncryptedAnswer, len(election.Questions))

	for i, q := range election.Questions {
		a := answers[i]
		results := make([]bool, len(q.Answers))
		//sum := int64(len(a))

		//min := q.Min
		//max := q.ComputeMax()
		//if sum < int64(min) || sum > int64(max) {
		//	glog.Errorf("Sum was %d, min was %d, and max was %d\n", sum, min, max)
		//	return nil, errors.New("invalid answers: sum must lie between min and max")
		//}

		ch := make([]*Ciphertext, len(results))
		//ip := make([]DisjunctiveZKProof, len(results))
		rs := make([]*big.Int, len(results))
		as := make([]int64, len(a))
		copy(as, a)

		// Mark each selected value as being voted for.
		for _, index := range a {
			results[index] = true
		}

		// Encrypt and create proofs for the answers, then create an overall proof if required
		tally := &Ciphertext{big.NewInt(1), big.NewInt(1)}
		randTally := big.NewInt(0)
		for j := range q.Answers {
			var err error
			if ch[j], rs[j] = v.Encrypt(results[j], pk); err != nil {
				fmt.Println("Couldn't encrypt choice %d for question %d\n", j, i)
				return nil, err
			}

			tally.MulCiphertexts(ch[j], pk.Prime)
			randTally.Add(randTally, rs[j])
			randTally.Mod(randTally, pk.ExponentPrime)
		}

		//var op DisjunctiveZKProof
		//if q.Max != 0 {
		//	op = make([]*ZKProof, q.Max-q.Min+1)
		//	for j := q.Min; j <= q.Max; j++ {
		//		if int64(j) != sum {
		//			// Create a simulated proof for the case where the
		//			// tally actually encrypts the value j.
		//			if err := op.CreateFakeProof(int64(j-q.Min), int64(j), tally, pk); err != nil {
		//				glog.Errorf("Couldn't create fake proof %d\n", j)
		//				return nil, err
		//			}
		//		}
		//	}
		//
		//	if err := op.CreateRealProof(sum-int64(q.Min), tally, randTally, pk); err != nil {
		//		glog.Errorf("Couldn't create the real proof")
		//		return nil, err
		//	}
		//}

		ans[i] = &EncryptedAnswer{ch, as, rs}
	}

	return &Ballot{ans, election.ElectionHash, election.Uuid}, nil
}

func main() {
	//implementation here
	fmt.Print("test")
	// var pk *Key
	// var secret *big.Int
	// pk, secret, _ = NewKey()
	// fmt.Println(pk)
	// fmt.Println(secret)

}
