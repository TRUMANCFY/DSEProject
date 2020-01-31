package voter

// implemented by Fengyu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type KeyStr struct {
	// Generator is the generator element g used in ElGamal encryptions.
	Generator string `json:"g"`

	// Prime is the prime p for the group used in encryption.
	Prime string `json:"p"`

	// ExponentPrime is another prime that specifies the group of exponent
	// values in the exponent of Generator. It is used in challenge
	// generation and verification.
	ExponentPrime string `json:"q"`

	// PublicValue is the public-key value y used to encrypt.
	PublicValue string `json:"y"`
}

type QAndA struct {
	Question string   `json:"question"`
	Answers  []string `json:"choices"`
}

func (v *Voter) ConvertStrToBigInt(s string) *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(s, 10)
	if !ok {
		fmt.Println("SetString: error")
		return nil
	}
	return n
}

func (v *Voter) CollectVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic("Wrong Methods")
	}

	var answers struct {
		Voter      int       `json:"voter"`
		Election   string    `json:"election"`
		Answers    [][]int64 `json:"answers"`
		PublicKey  KeyStr    `json:"publickey"`
		QuesAndAns []QAndA   `json:"qanda"`
	}

	json.NewDecoder(r.Body).Decode(&answers)

	fmt.Println(answers)

	pk := &Key{
		Generator:     v.ConvertStrToBigInt(answers.PublicKey.Generator),
		Prime:         v.ConvertStrToBigInt(answers.PublicKey.Prime),
		ExponentPrime: v.ConvertStrToBigInt(answers.PublicKey.ExponentPrime),
		PublicValue:   v.ConvertStrToBigInt(answers.PublicKey.PublicValue),
	}

	// we need to get the election with the corresponding name

	electionPk := &Election{}

	electionPk.PublicKey = pk

	electionPk.Name = answers.Election

	electionPk.Questions = make([]*Question, 0)

	fmt.Println("====Q&A====")
	fmt.Println(answers.QuesAndAns)

	for _, q := range answers.QuesAndAns {
		electionPk.Questions = append(electionPk.Questions, &Question{
			Answers:  q.Answers,
			Question: q.Question,
		})
	}

	fmt.Println("=====election=====")
	fmt.Println(electionPk)

	fmt.Println("======answers=====")
	fmt.Println(answers.Answers)

	// encode
	vote, _ := NewCastBallot(electionPk, answers.Answers)

	// who vote it. election, vote

	vote.VoterUuid = strconv.Itoa(answers.Voter)
	vote.VoterHash = vote.VoterUuid
	vote.VoteHash = answers.Election + vote.VoterUuid

	vote.Vote.ElectionUuid = answers.Election

	fmt.Println(vote.Vote.Answers[0].Answer)

	v.SendEncrypted(vote)

	v.AckPost(true, w)
}

func (v *Voter) SendEncrypted(vote *CastBallot) {
	trustees := make([]string, 3)

	trustees[0] = "http://127.0.0.1:8000/vote"
	trustees[1] = "http://127.0.0.1:8001/vote"
	trustees[2] = "http://127.0.0.1:8002/vote"

	fmt.Println(vote.VoterUuid)
	fmt.Println(vote)

	sendVal := map[string]CastBallot{"vote": *vote}
	jsonVal, _ := json.Marshal(sendVal)

	for _, t := range trustees {
		resp, _ := http.Post(t, "application/json", bytes.NewBuffer(jsonVal))
		fmt.Println(resp)
	}
}

func (v *Voter) CreateElection(w http.ResponseWriter, r *http.Request) {
	type Qlist struct {
		Question string   `json:"question"`
		Choices  []string `json:"choices"`
	}

	var election struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Questions   []Qlist `json:"questions"`
		Creator     string  `json:"creator"`
	}

	json.NewDecoder(r.Body).Decode(&election)

	// generate the list of question
	questionList := make([]*Question, 0)

	for _, d := range election.Questions {
		q := &Question{
			Max:      1,
			Min:      1,
			Question: d.Question,
			Answers:  d.Choices,
		}
		questionList = append(questionList, q)
	}

	newElection, _, _ := NewElection("https://example.com", election.Description, time.Now().String(),
		election.Name, false, questionList, "Fake",
		false, "Fake hash", time.Now().String(), time.Now().String(), nil)

	values := map[string]Election{"elec": *newElection}
	jsonValue, _ := json.Marshal(values)
	// target, _ := url.Parse("127.0.0.1:8081/election")
	resp, err := http.Post("http://127.0.0.1:8081/election", "application/json", bytes.NewBuffer(jsonValue))

	fmt.Println(resp)

	if err != nil {
		panic(err)
	}

	v.AckPost(true, w)
}

func (v *Voter) EndVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic("Wrong method")
	}
	var election struct {
		Electionend string `json:"electionend"`
	}
	json.NewDecoder(r.Body).Decode(&election)

	fmt.Println("===========")
	fmt.Println(election.Electionend)

	trustees := make([]string, 3)

	trustees[0] = "http://127.0.0.1:8000/endvote"
	trustees[1] = "http://127.0.0.1:8001/endvote"
	trustees[2] = "http://127.0.0.1:8002/endvote"

	for _, target := range trustees {
		values := map[string]string{"elec": election.Electionend}
		jsonValue, _ := json.Marshal(values)
		http.Post(target, "application/json", bytes.NewBuffer(jsonValue))
	}

	v.AckPost(true, w)
}

func (v *Voter) AckPost(success bool, w http.ResponseWriter) {
	var response struct {
		Success bool `json:"success"`
	}
	response.Success = success
	json.NewEncoder(w).Encode(response)
}

func (v *Voter) ListenToGui() {
	r := mux.NewRouter()
	r.HandleFunc("/vote", v.CollectVote).Methods("POST")
	r.HandleFunc("/createElection", v.CreateElection).Methods("POST")
	r.HandleFunc("/endvote", v.EndVote).Methods("POST")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./web/frontend/dist/"))))
	srv := &http.Server{
		Handler:           r,
		Addr:              "127.0.0.1:" + v.Port,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
