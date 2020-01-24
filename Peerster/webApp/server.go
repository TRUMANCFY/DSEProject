package main

import (

	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {

	// Register router
	r := mux.NewRouter()

	// Register handlers
	r.HandleFunc("/message", MessageGetHandler).
		Methods("GET")
	r.HandleFunc("/node",  NodeGetHandler).
		Methods("GET")
	r.HandleFunc("/message", MessageSetHandler).
		Methods("POST")
	r.HandleFunc("/node", NodeSetHandler).
		Methods("POST")
	r.HandleFunc("/id", IDGetHandler).
		Methods("GET")

	// Get GUIPort
	var guiPort int
	flag.IntVar(&guiPort, "UIPort", 8080, "GUI port")
	flag.Parse()

	fmt.Printf("Starting webapp on address http://127.0.0.1:%d", guiPort)

	srv := &http.Server{

		Handler : r,
		Addr : fmt.Sprintf("127.0.0.1:%d", guiPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}