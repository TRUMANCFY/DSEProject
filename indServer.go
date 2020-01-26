package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	. "github.com/TRUMANCFY/DSEProject/voter"
	"github.com/gorilla/mux"
)

type Server struct {
	listElection []Election
}

type PKContainer struct {
	Name      string `json:"name"`
	PublicKey Key    `json:"publickey"`
}

type PartialKeyContainer struct {
	Name       string   `json:"name"`
	PartialKey *big.Int `json:"partialkey"`
	Trust      *Trustee `json:"trust"`
	Elec       Election `json:"elec"`
}

const PYTHON_SERVER = "127.0.0.1:4000"

func (s Server) ReceiveElection(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic("wrong method")
		return
	}

	var comingElection struct {
		Elec Election `json:"elec"`
	}

	json.NewDecoder(r.Body).Decode(&comingElection)

	var pk *Key
	var secret *big.Int
	pk, secret, err := NewKey()
	comingElection.Elec.PublicKey = pk
	comingElection.Elec.Secret = secret

	// create trustees from election
	trusteeCount := 3
	var trustees []*Trustee
	var trusteeSecrets []*big.Int
	trustees, trusteeSecrets, _ = SplitKey(comingElection.Elec.Secret, comingElection.Elec.PublicKey, trusteeCount)

	// hardcoded addresses of the trusteeSecrets
	trustees[0].Address = "http://127.0.0.1:8000/partialkey"
	trustees[1].Address = "http://127.0.0.1:8001/partialkey"
	trustees[2].Address = "http://127.0.0.1:8002/partialkey"

	// add those trustees to the election
	comingElection.Elec.Trustees = trustees

	var sendVal map[string]PartialKeyContainer

	var jsonVal []byte

	// append the election to the public elecrions
	s.listElection = append(s.listElection, comingElection.Elec)

	fmt.Println(comingElection.Elec.Questions)

	elecSend := comingElection.Elec

	// comingElection.Elec.Name
	pkContainer := PKContainer{
		Name:      comingElection.Elec.Name,
		PublicKey: *pk,
	}

	values := map[string]PKContainer{"pkcontainer": pkContainer}
	jsonValue, _ := json.Marshal(values)
	// target, _ := url.Parse("127.0.0.1:8081/election")
	resp, err := http.Post("http://127.0.0.1:4000/publickey", "application/json", bytes.NewBuffer(jsonValue))
	fmt.Print(resp)
	if err != nil {
		panic(err)
	}

	//send PM POST to each trustee with its secret
	for i, t := range trustees {

		partialKeyCon := PartialKeyContainer{
			Name:       comingElection.Elec.Name,
			PartialKey: trusteeSecrets[i],
			Trust:      trustees[i],
			Elec:       elecSend,
		}

		fmt.Println("=======")
		fmt.Println(partialKeyCon)

		sendVal = map[string]PartialKeyContainer{"partial": partialKeyCon}
		jsonVal, _ = json.Marshal(sendVal)

		http.Post(t.Address, "application/json", bytes.NewBuffer(jsonVal))

	}
}

func (s *Server) GetElectionInfo(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) AckPost(key Key, w http.ResponseWriter) {
	var response struct {
		PublicKey Key `json:"publickey"`
	}
	response.PublicKey = key
	json.NewEncoder(w).Encode(response)
}

func (s *Server) ListenToGui() {
	r := mux.NewRouter()
	r.HandleFunc("/election", s.ReceiveElection).Methods("POST")
	r.HandleFunc("/getElection", s.GetElectionInfo).Methods("POST")
	// r.HandleFunc("/createElection", s.CreateElection).Methods("POST")
	// r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./web/frontend/dist/"))))
	srv := &http.Server{
		Handler:           r,
		Addr:              "127.0.0.1:8081",
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func main() {
	listElection := make([]Election, 0)
	s := &Server{
		listElection: listElection,
	}

	s.ListenToGui()
}
