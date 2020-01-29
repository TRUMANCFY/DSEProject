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
	Res    map[string]message.Result
}

func (t *Tally) ReceiveTally(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Receive Tally")
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

	// elecName := tallyObj.Tally.Elec.Uuid

	vote := tallyObj.Tally.Vote

	elec := tallyObj.Tally.Elec

	// put it in
	_, ok := t.Record[elec.Name]

	if !ok {
		t.Record[elec.Name] = make(map[string]TallyContainer)
	}

	t.Record[elec.Name][src] = tallyObj.Tally

	trustees := make([]*message.Trustee, 0)
	// check whether there are all
	if len(t.Record[elec.Name]) == 3 {
		for _, tallyo := range t.Record[elec.Name] {
			trustees = append(trustees, tallyo.Trustee)
		}
		res, err := elec.Tallier(vote, trustees)

		if err != nil {
			fmt.Println(err)
		}

		// put the result into the container
		t.Res[elec.Name] = res

		fmt.Println(t.Res)
	}

	t.Mux.Unlock()

}

func (t *Tally) GetElectionResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic("Wrong Method")
	}

	// read the election name
	var comingElection struct {
		Elec string `json:"elec"`
	}

	json.NewDecoder(r.Body).Decode(&comingElection)

	elecName := comingElection.Elec

	res, ok := t.Res[elecName]

	fmt.Println(elecName)

	var ResultToSend struct {
		Res   message.Result `json:"res"`
		Exist bool           `json:"exist"`
	}

	if ok {
		ResultToSend.Exist = ok
		ResultToSend.Res = res
	} else {
		ResultToSend.Exist = ok
	}

	fmt.Println(res)

	json.NewEncoder(w).Encode(ResultToSend)
}

func (t *Tally) AckPost(success bool, w http.ResponseWriter) {
	var response struct {
		Success bool `json:"success"`
	}
	response.Success = success
	json.NewEncoder(w).Encode(response)
}

func (t *Tally) ListenToGui() {
	r := mux.NewRouter()
	r.HandleFunc("/tally", t.ReceiveTally).Methods("POST")
	r.HandleFunc("/getresult", t.GetElectionResult).Methods("POST")
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
	res := make(map[string]message.Result)
	res["test"] = make([][]int64, 0)
	res["test"] = append(res["test"], []int64{2, 0})
	res["test"] = append(res["test"], []int64{2, 5})

	t := Tally{
		Record: make(map[string](map[string]TallyContainer)),
		Mux:    &sync.Mutex{},
		Res:    res,
	}

	t.ListenToGui()
}
