package main

import (
	"encoding/json"
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

	s.listElection = append(s.listElection, comingElection.Elec)

}

func (s *Server) ListenToGui() {
	r := mux.NewRouter()
	r.HandleFunc("/election", s.ReceiveElection).Methods("POST")
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
