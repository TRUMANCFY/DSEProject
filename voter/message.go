package voter

import "math/big"

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
}

type CastBallot struct {
	// JSON is the JSON string corresponding to this type. This is not part
	// of the original JSON structure (obviously).
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

// A Ciphertext is an ElGamal ciphertext, where g is Key.Generator, r is a
// random value, m is a message, and y is Key.PublicValue.
type Ciphertext struct {
	// Alpha = g^r
	Alpha *big.Int `json:"alpha"`

	// Beta = g^m * y^r
	Beta *big.Int `json:"beta"`
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

	Secret *big.Int
}
