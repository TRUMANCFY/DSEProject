package gossiper

import (
	"os"
	"fmt"
	"log"
	"time"
	"strings"
	"net/http"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/LiangweiCHEN/Peerster/message"
)
/*****************************************************/
// GUI Handling

func (g *Gossiper) HandleGUI() {

	// Register router
	go func() {
		r := mux.NewRouter()

		// Register handlers
		r.HandleFunc("/message", g.MessageGetHandler).
			Methods("GET", "OPTIONS")
		r.HandleFunc("/node",  g.NodeGetHandler).
			Methods("GET", "OPTIONS")
		r.HandleFunc("/message", g.MessagePostHandler).
			Methods("POST", "OPTIONS")
		r.HandleFunc("/node", g.NodePostHandler).
			Methods("POST", "OPTIONS")
		r.HandleFunc("/id", g.IDGetHandler).
			Methods("GET", "OPTIONS")
		r.HandleFunc("/routing", g.RoutableGetHandler).
			Methods("GET", "OPTIONS")
		r.HandleFunc("/search", g.SearchedGetHandler).
			Methods("GET", "OPTIONS")
		r.HandleFunc("/routing", g.PrivateMsgSendHandler).
			Methods("POST", "OPTIONS")
		r.HandleFunc("/sharing", g.ShareFileHandler).
			Methods("POST", "OPTIONS")
		r.HandleFunc("/request", g.RequestFileHandler).
			Methods("POST", "OPTIONS")
		r.HandleFunc("/search", g.SearchHandler).
			Methods("POST", "OPTIONS")
		r.HandleFunc("/download", g.DownloadHandler).
			Methods("POST", "OPTIONS")
		r.HandleFunc("/vote", g.VoteHandler).
			Methods("POST", "OPTIONS")
		fmt.Printf("Starting webapp on address http://127.0.0.1:%s\n", g.GuiPort)

		srv := &http.Server{

			Handler : r,
			Addr : fmt.Sprintf("127.0.0.1:%s", g.GuiPort),
			WriteTimeout: 15 * time.Second,
			ReadTimeout: 15 * time.Second,
		}

		log.Fatal(srv.ListenAndServe())
	}()
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (g *Gossiper) MessageGetHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var messages struct {
		Messages []string `json:"messages"`
	}

	messages.Messages = g.GetMessages()

	json.NewEncoder(w).Encode(messages)
}

func (g *Gossiper) GetMessages() ([]string){
	// Return all rumors

	buffer := g.MsgBuffer.Msg

	return buffer
}

func (g *Gossiper) MessagePostHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)
	var message struct {

		Text string `json:"text"`
	}

	json.NewDecoder(r.Body).Decode(&message)

	fmt.Printf("Receive new msg from GUI%v",message)
	g.PostNewMessage(message.Text)

	g.AckPost(true, w)
}

func (g *Gossiper) PostNewMessage(text string) {

	// Create Simple msg
	message := &message.Message{
		Text : text,
	}

	// Trigger handle simple msg
	go g.HandleClient(message)
}

func (g *Gossiper) NodeGetHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var peers struct {
		Nodes []string `json:"nodes"`
	}

	peers.Nodes = g.GetPeers()
    
	json.NewEncoder(w).Encode(peers)
}

func (g *Gossiper) GetPeers() ([]string) {
     
	// TODO: Decide whether need to lock
	return g.Peers.Peers
}

func (g *Gossiper) NodePostHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)
	var peer struct {
		Addr string `json:"addr"`
	}

	json.NewDecoder(r.Body).Decode(&peer)

	g.AddNewNode(peer.Addr)

	g.AckPost(true, w)
}

func (g *Gossiper) AddNewNode(addr string) {

	g.Peers.Mux.Lock()
	g.Peers.Peers = append(g.Peers.Peers, addr)
	fmt.Println("After adding new node, our peers are ", g.Peers.Peers)
	g.Peers.Mux.Unlock()
}

func (g *Gossiper) IDGetHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)
	var ID struct {

		ID string `json:"id"`
	}

	ID.ID = g.GetPeerID()

	json.NewEncoder(w).Encode(ID)
}

func (g *Gossiper) GetPeerID() (ID string) {

	ID = g.Name
	return
}

func (g *Gossiper) RoutableGetHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var peers struct {
		Nodes []string `json:"nodes"`
	}

	peers.Nodes = g.GetRoutable()
    
	json.NewEncoder(w).Encode(peers)
}

func (g *Gossiper) GetRoutable() ([]string) {
     
	routable := make([]string, 0)
	for k, _ := range g.Dsdv.Map {

		routable = append(routable, k)
	}

	return routable
}

func (g *Gossiper) SearchedGetHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	var matched struct {
		Matches []string `json:"matches"`
	}

	matched.Matches = g.GetSearchedFileName()

	json.NewEncoder(w).Encode(matched)
}

func (g *Gossiper) GetSearchedFileName() []string {

	matched := make([]string, 0)
	g.FileSharer.Searcher.TargetMetahash.Mux.Lock()
	for k, _ := range g.FileSharer.Searcher.TargetMetahash.Map {
		matched = append(matched, k)
	}
	g.FileSharer.Searcher.TargetMetahash.Mux.Unlock()

	return matched
}
func (g *Gossiper) PrivateMsgSendHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var messageReceived struct {

		Text string
		Dest string
	}

	json.NewDecoder(r.Body).Decode(&messageReceived)

	msg := &message.Message{

		Text : messageReceived.Text,
		Destination : &messageReceived.Dest,
	}
	fmt.Println("TRIGGER HANDLING PRIVATE")
	go g.HandleClient(msg)

	g.AckPost(true, w)
}

func (g *Gossiper) ShareFileHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var fileName struct {
		Name string
	}

	json.NewDecoder(r.Body).Decode(&fileName)

	// Trigger fileSharer to index that file
	// TODO: May need to change this func to be concurrent
	go func(fileName *string) {
		if !g.Hw3ex3 {
			tx, _ := g.FileSharer.CreateIndexFile(fileName)
			round := g.Round
			g.SendTLC(*tx, round)
		
		} else {
			tx, _ := g.FileSharer.CreateIndexFile(fileName)
			g.TransactionSendCh<- tx
		}
	}(&fileName.Name)

	g.AckPost(true, w)
}

func (g *Gossiper) AckPost(success bool, w http.ResponseWriter) {

	enableCors(&w)
	var response struct {
		Success bool `json:"success"`
	}
	response.Success = success 
	json.NewEncoder(w).Encode(response)
}

func (g *Gossiper) RequestFileHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var msg struct {

		Dest string
		FileName string
		MetaHash string
	}

	json.NewDecoder(r.Body).Decode(&msg)

	// Trigger file request sending
	metaHash, err := hex.DecodeString(msg.MetaHash)
	if err != nil {

		fmt.Printf("ERROR (Unable to decode hash)")
		os.Exit(1)
	}
	g.FileSharer.RequestFile(&msg.FileName, &metaHash, &msg.Dest)

	g.AckPost(true, w)
}

func (g *Gossiper) SearchHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var msg struct {
		Keywords string
	}

	json.NewDecoder(r.Body).Decode(&msg)

	// Trigger keyword search
	keywords := strings.Split(msg.Keywords, ",")
	g.FileSharer.Searcher.Search(keywords, 2)

	g.AckPost(true, w)
}

func (g *Gossiper) DownloadHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var fileName struct {
		Name string
	}

	json.NewDecoder(r.Body).Decode(&fileName)

	// Get the metahash of the file
	g.FileSharer.Searcher.TargetMetahash.Mux.Lock()
	metahash := g.FileSharer.Searcher.TargetMetahash.Map["_SharedFiles/" + fileName.Name]
	fmt.Println(fileName)
	g.FileSharer.Searcher.TargetMetahash.Mux.Unlock()
	fmt.Printf("Requesting %s\n", hex.EncodeToString(metahash))
	// Trigger download of the file
	go g.FileSharer.Searcher.RequestSearchedFile(fileName.Name, metahash)

	g.AckPost(true, w)
}

func (g *Gossiper) VoteHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	var castBallot message.CastBallot

	json.NewDecoder(r.Body).Decode(&castBallot)

	fmt.Printf("GET VOTE FROM %s VOTING FOR %s", castBallot.VoterUuid, castBallot.VoteHash)

	go g.HandleReceivingVote(&castBallot)
}