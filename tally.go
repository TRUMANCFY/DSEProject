package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/TRUMANCFY/DSEProject/Peerster/message"
	"github.com/gorilla/mux"
)

type TallyContainer struct {
	Trustee *message.Trustee      `json:"trustee"`
	Vote    []*message.CastBallot `json:"vote"`
	Src     string                `json:"src"`
	Elec    message.Election      `json:"elec"`
}

type Tally struct {
	Record map[string](map[string]TallyContainer)
	Mux    *sync.Mutex
}

func (t *Tally) ReceiveTally(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic("wront type")
		return
	}

	var tallyObj struct {
		Tally TallyContainer `json:"tally"`
	}

	json.NewDecoder(r.Body).Decode(&tallyObj)

	fmt.Println(tallyObj.Tally)

	t.Mux.Lock()

	src := tallyObj.Tally.Src

	fmt.Println(tallyObj.Tally.Trustee)

	elecName := tallyObj.Tally.Elec.Uuid

	vote := tallyObj.Tally.Vote

	elec := tallyObj.Tally.Elec

	// put it in
	_, ok := t.Record[elecName]

	if !ok {
		t.Record[elecName] = make(map[string]TallyContainer)
	}

	t.Record[elecName][src] = tallyObj.Tally

	trustees := make([]*message.Trustee, 0)
	// check whether there are all
	if len(t.Record[elecName]) == 3 {
		for _, tallyo := range t.Record[elecName] {
			trustees = append(trustees, tallyo.Trustee)
		}
		res, err := elec.Tallier(vote, trustees)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(res)
	}

	t.Mux.Unlock()

}

func (t *Tally) ListenToGui() {
	r := mux.NewRouter()
	r.HandleFunc("/tally", t.ReceiveTally).Methods("POST")
	// r.HandleFunc("/createElection", s.CreateElection).Methods("POST")
	// r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./web/frontend/dist/"))))
	srv := &http.Server{
		Handler:           r,
		Addr:              "127.0.0.1:8082",
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func main() {
	t := Tally{
		Record: make(map[string](map[string]TallyContainer)),
		Mux:    &sync.Mutex{},
	}

	t.ListenToGui()
}
