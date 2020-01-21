package voter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (v *Voter) CollectVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic("Wrong Methods")
	}

	var answers struct {
		User     string  `json:"user`
		Election string  `json:"election"`
		Answers  [][]int `json:"answers"`
	}

	json.NewDecoder(r.Body).Decode(&answers)

	fmt.Println(answers)

	v.AckPost(true, w)
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
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./web/frontend/dist/"))))
	srv := &http.Server{
		Handler:           r,
		Addr:              "127.0.0.1:8080",
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
