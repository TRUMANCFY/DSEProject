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
	trusteeCount := 4
	var trustees []*Trustee
	var trusteeSecrets []*big.Int
	trustees, trusteeSecrets, _ = SplitKey(comingElection.Elec.Secret, comingElection.Elec.PublicKey, trusteeCount)

	// hardcoded addresses of the trusteeSecrets
	trustees.Address[0] = "00000000"
	trustees.Address[1] = "00000000"
	trustees.Address[2] = "00000000"
	trustees.Address[3] = "00000000"

	// add those trustees to the election
	comingElection.Trustees = trustees

	//send PM POST to each trustee with its secret

	// destroy the secret key so that the independant party no longer has knowledge of it
	comingElection.Elec.Secret = ""

	// append the election to the public elecrions
	s.listElection = append(s.listElection, comingElection.Elec)

	// send the public key to python server

	// comingElection.Elec.Name
	pkContainer := PKContainer{
		Name:      comingElection.Elec.Name,
		PublicKey: *pk,
	}

	values := map[string]PKContainer{"pkContainer": pkContainer}
	jsonValue, _ := json.Marshal(values)
	// target, _ := url.Parse("127.0.0.1:8081/election")
	resp, err := http.Post("http://127.0.0.1:4000/publickey", "application/json", bytes.NewBuffer(jsonValue))

	fmt.Println(resp)

	if err != nil {
		panic(err)
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
	r.HandleFunc("/getElection", s.GetElectionInfo.Methods("POST"))
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
