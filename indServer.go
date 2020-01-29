package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"time"

	. "github.com/TRUMANCFY/DSEProject/voter"
	"github.com/gorilla/mux"
)

type Server struct {
	listElection []ElectionStruct
}

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

func ConvertBigIntToStr(key *Key) KeyStr {
	return KeyStr{
		Generator:     key.Generator.String(),
		Prime:         key.Prime.String(),
		ExponentPrime: key.ExponentPrime.String(),
		PublicValue:   key.PublicValue.String(),
	}
}

type PKContainer struct {
	Name      string `json:"name"`
	PublicKey KeyStr `json:"publickey"`
}

type PartialKeyContainer struct {
	Name       string   `json:"name"`
	PartialKey *big.Int `json:"partialkey"`
	Trust      *Trustee `json:"trust"`
	Elec       Election `json:"elec"`
}

type PrivateKeyMap struct {
	Src        string `json:"src"`
	PrivateKey string `json:"privatekey"`
}

type ElectionStruct struct {
	Elec        string          `json:"elec"`
	PublicKey   KeyStr          `json:"publickey"`
	PrivateKeys []PrivateKeyMap `json:"privatekeys"`
}

const PYTHON_SERVER = "127.0.0.1:4000"

func (s *Server) ReceiveElection(w http.ResponseWriter, r *http.Request) {
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
	// s.listElection = append(s.listElection, comingElection.Elec)

	fmt.Println(comingElection.Elec.Questions)

	elecSend := comingElection.Elec

	// comingElection.Elec.Name
	pkContainer := PKContainer{
		Name:      comingElection.Elec.Name,
		PublicKey: ConvertBigIntToStr(pk),
	}

	values := map[string]PKContainer{"pkcontainer": pkContainer}
	jsonValue, _ := json.Marshal(values)
	// target, _ := url.Parse("127.0.0.1:8081/election")
	resp, err := http.Post("http://127.0.0.1:4000/publickey", "application/json", bytes.NewBuffer(jsonValue))
	fmt.Print(resp)
	if err != nil {
		panic(err)
	}

	privateKeys := make([]PrivateKeyMap, 0)

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

		privateKeys = append(privateKeys, PrivateKeyMap{
			Src:        t.Address,
			PrivateKey: trusteeSecrets[i].String(),
		})

	}

	elecStruct := ElectionStruct{
		Elec:        comingElection.Elec.Name,
		PublicKey:   ConvertBigIntToStr(pk),
		PrivateKeys: privateKeys,
	}

	s.listElection = append(s.listElection, elecStruct)

	fmt.Println("+++++")
	fmt.Println(s.listElection)

}

func (s *Server) GetElectionInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		panic("wrong method")
		return
	}

	var messages struct {
		Messages []ElectionStruct `json:"messages"`
	}

	messages.Messages = s.listElection

	json.NewEncoder(w).Encode(messages)
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
	r.HandleFunc("/getElection", s.GetElectionInfo).Methods("GET")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./web/indserver/dist/"))))
	srv := &http.Server{
		Handler:           r,
		Addr:              "127.0.0.1:8081",
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (s *Server) SendAuth() {
	ranStr := RandStringRunes(32)

	var trustees = make([]string, 3)

	trustees[0] = "http://127.0.0.1:8000/auth"
	trustees[1] = "http://127.0.0.1:8001/auth"
	trustees[2] = "http://127.0.0.1:8002/auth"

	var sendVal map[string]string
	var jsonVal []byte

	for _, t := range trustees {

		sendVal = map[string]string{"auth": ranStr}
		jsonVal, _ = json.Marshal(sendVal)

		http.Post(t, "application/json", bytes.NewBuffer(jsonVal))
	}

}

func main() {
	listElection := make([]ElectionStruct, 0)
	s := &Server{
		listElection: listElection,
	}

	s.SendAuth()

	s.ListenToGui()
}
