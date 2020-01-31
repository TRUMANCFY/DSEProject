// implemented by Fengyu

package voter

import (
	"errors"
	"fmt"
	"math/big"
)

type Trustee struct {
	DecryptionFactors [][]*big.Int `json:"decryption_factors"`

	PublicKey *Key `json:"public_key"`

	PublicKeyHash string `json:"public_key_hash"`

	// Uuid is the unique identifier for this Trustee.
	Uuid string `json:"uuid"`

	// Address of the trustee
	Address string `json:"address"`

	// Name of election
	Election string `json:"election"`
}

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

type Key struct {
	Generator *big.Int `json:"g"`

	Prime *big.Int `json:"p"`

	ExponentPrime *big.Int `json:"q"`

	PublicValue *big.Int `json:"y"`
}

type Ciphertext struct {
	// Alpha = g^r
	Alpha *big.Int `json:"alpha"`

	// Beta = g^m * y^r
	Beta *big.Int `json:"beta"`
}

// A Question is part of an Election and specifies a question to be voted on.
type Question struct {
	AnswerUrls []string `json:"answer_urls"`

	Answers []string `json:"answers"`

	ChoiceType string `json:"choice_type"`

	Max int `json:"max"`

	Min int `json:"min"`

	Question string `json:"question"`

	ResultType string `json:"result_type"`

	ShortName string `json:"short_name"`

	TallyType string `json:"tally_type"`
}

// An EncryptedAnswer is part of a Ballot cast by a Voter. It is the answer to
// a given Question in an Election.
type EncryptedAnswer struct {
	Choices []*Ciphertext `json:"choices"`

	Answer []int64 `json:"answer,omitempty"`

	Randomness []*big.Int `json:"randomness,omitempty"`
}

// A Ballot is a cryptographic vote in an Election.
type Ballot struct {
	Answers []*EncryptedAnswer `json:"answers"`

	ElectionHash string `json:"election_hash"`

	ElectionUuid string `json:"election_uuid"`
}

// An Election contains all the information about a Helios election.
type Election struct {
	JSON []byte `json:"-"`

	ElectionHash string `json:"-"`

	CastURL string `json:"cast_url"`

	// Description is a plaintext description of the election.
	Description string `json:"description"`

	// FrozenAt is the date at which the election was fully specified and
	// frozen.
	FrozenAt string `json:"frozen_at"`

	// Name is the full name of the election.
	Name string `json:"name"`

	Openreg bool `json:"openreg"`

	PublicKey *Key `json:"public_key"`

	// Questions is the list of questions to be voted on in this election.
	Questions []*Question `json:"questions"`

	// ShortName provides a short plaintext name for this election.
	ShortName string `json:"short_name"`

	UseVoterAliases bool `json:"use_voter_aliases"`

	Uuid string `json:"uuid"`

	// VotersHash provides the hash of the list of voters.
	VotersHash string `json:"voters_hash"`

	VotingEndsAt   string `json:"voting_ends_at"`
	VotingStartsAt string `json:"voting_starts_at"`

	Secret *big.Int

	Trustees []*Trustee `json:"trustee"`
}

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
