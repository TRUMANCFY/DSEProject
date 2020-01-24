package main

import (
	"flag"
	"fmt"
	"github.com/LiangweiCHEN/Peerster/fileSharing"
	"github.com/LiangweiCHEN/Peerster/gossiper"
	"github.com/LiangweiCHEN/Peerster/message"
	"github.com/LiangweiCHEN/Peerster/network"
	"github.com/LiangweiCHEN/Peerster/routing"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

/** Global variable **/
var UIPort, GuiPort, gossipAddr, name string
var peers []string
var simple bool
var antiEntropy int
var rtimer int
var sharedFilePath string
var stubbornTimeout int
var numPeers int
var hw3ex2 bool
var hw3ex3 bool
var ackAll bool
func input() (UIPort string, GuiPort string, gossipAddr string, name string, peers []string, simple, hw3ex2, hw3ex3 bool, antiEntropy int, rtimer int, sharedFilePath string,
	stubbornTimeout int, numPeers int, ackAll bool) {

	// Set flag value containers
	flag.StringVar(&UIPort, "UIPort", "8080", "UI port num")

	flag.StringVar(&gossipAddr, "gossipAddr", "127.0.0.1:5000",
		"gossip addr")

	flag.StringVar(&name, "name", "", "name of gossiper")

	flag.StringVar(&GuiPort, "GuiPort", "", "GUI port, default to be UIPort + GossipPort")
	var peers_str string

	flag.StringVar(&peers_str, "peers", "", "list of peers")

	flag.BoolVar(&simple, "simple", false, "Simple broadcast or not")

	flag.BoolVar(&hw3ex2, "hw3ex2", false, "Whether run in mode for homework 3 exercise 2")
	
	flag.BoolVar(&hw3ex3, "hw3ex3", false, "Whether run in mode for homework 3 exercise 3")

	flag.IntVar(&antiEntropy, "antiEntropy", 10, "antiEntroypy trigger period")

	flag.IntVar(&rtimer, "rtimer", 0, "Routing heartbeat period")

	flag.StringVar(&sharedFilePath, "file", "_SharedFiles", "shared file path")

	flag.IntVar(&stubbornTimeout, "stubbornTimeout", 5, "timeout between two continous blockchain proposal")

	flag.IntVar(&numPeers, "N", 1, "number of peers in the system")

	flag.BoolVar(&ackAll, "ackAll", false, "whether to ack all incoming tlc message")

	// Conduct parameter retreival
	flag.Parse()
	
	if hw3ex3 { fmt.Printf("We are in hw3ex3") }
	fmt.Printf("Number of peers is %d\n", numPeers)
	// Convert peers to slice
	peers = strings.Split(peers_str, ",")
	if peers[0] == "" {
		peers = peers[1:]
	}
	return
}

func InitGossiper(UIPort, gossipAddr, name string, simple bool, peers []string, antiEntropy, rtimer int, sharedFilePath string) (g *gossiper.Gossiper) {

	// Establish gossiper addr and conn
	addr, _ := net.ResolveUDPAddr("udp", gossipAddr)
	conn, _ := net.ListenUDP("udp", addr)

	// Establish client addr and conn
	client_addr, _ := net.ResolveUDPAddr("udp", ":"+UIPort)
	client_conn, _ := net.ListenUDP("udp", client_addr)

	// Check whether need to use default GUIPort
	if GuiPort == "" {
		GuiPortInt, _ := strconv.Atoi(UIPort)
		offset, _ := strconv.Atoi(strings.Split(gossipAddr, ":")[1])
		GuiPortInt += offset
		GuiPort = strconv.Itoa(GuiPortInt)
	}

	/*
	if !hw3ex2 && !hw3ex3 {
		fmt.Println("We are not in hw3ex2 or hw3ex3")
		numPeers = 1
	}
	*/
	// Create gossiper
	g = &gossiper.Gossiper{
		Address: gossipAddr,
		Conn:    conn,
		Name:    name,
		UIPort:  UIPort,
		GuiPort: GuiPort,
		Peers: &gossiper.PeersBuffer{

			Peers: peers,
		},
		Simple: simple,
		N: &network.NetworkHandler{

			Conn:             conn,
			Addr:             addr,
			Client_conn:      client_conn,
			Send_ch:          make(chan *message.PacketToSend),
			Listen_ch:        make(chan *message.PacketIncome),
			Client_listen_ch: make(chan *message.Message),
			Done_chs: &network.Done_chs{

				Chs: make(map[string]chan struct{}),
			},
			RumorTimeoutCh: make(chan *message.PacketToSend),
		},
		RumorBuffer: &gossiper.RumorBuffer{

			Rumors: make(map[string][]*message.WrappedRumorTLCMessage),
		},
		StatusBuffer: &gossiper.StatusBuffer{

			Status: make(message.StatusMap),
		},
		AckChs: &gossiper.Ack_chs{
			Chs: make(map[string]chan *gossiper.PeerStatusAndSync),
		},
		PeerStatuses: &gossiper.PeerStatuses{
			Map: make(map[string]map[string]uint32),
		},
		AntiEntropyPeriod: antiEntropy,
		Dsdv: &routing.DSDV{
			Map: make(map[string]string),
			Ch:  make(chan *routing.OriginRelayer),
		},
		RTimer:             rtimer,
		HopLimit:           uint32(10),
		SharedFilePath:     sharedFilePath,
		SearchDistributeCh: make(chan *message.SearchRequestRelayer),
		TLCAckChs: &gossiper.TLCAckChs{
			Chs: make(map[uint32]chan []string),
		},
		TLCAckCh:        make(chan *message.PacketIncome, 100),
		StubbornTimeout: stubbornTimeout,
		NumPeers:        numPeers,
		TLCClock: &gossiper.TLCClock{
			Clock : make(map[string]int),
			Map : make(map[string]map[uint32]int),
		},
		TLCRoundCh : make(chan struct{}),
		WrappedTLCCh : make(chan *gossiper.WrappedTLCMessage),
		ConfirmedMessageCh : make(chan *message.TLCMessage, 1000),
		TransactionSendCh : make(chan *message.TxPublish),
		Hw3ex2 : hw3ex2,
		Hw3ex3 : hw3ex3,
		AckAll : ackAll,
		MsgBuffer : gossiper.MsgBuffer{
			Msg : make([]string, 0),
		},
	}

	g.FileSharer = &fileSharing.FileSharer{

		N: g.N,
		Indexer: &fileSharing.FileIndexer{
			SharedFolder: g.SharedFilePath,
		},
		RequestReplyChMap: &fileSharing.RequestReplyChMap{
			Map: make(map[string]chan *message.DataReply),
		},
		HopLimit:       g.HopLimit,
		Origin:         g.Name,
		RequestTimeout: 15,
		IndexFileMap: &fileSharing.IndexFileMap{
			Map: make(map[string]*fileSharing.IndexFile),
		},
		ChunkHashMap: &fileSharing.ChunkHashMap{
			Map: make(map[string]bool),
		},
		Dsdv: g.Dsdv,
		Downloading: &fileSharing.Downloading{
			Map: make(map[string]chan *message.DataReply),
		},
		FileLocker: &fileSharing.FileLocker{
			Map: make(map[string]*sync.Mutex),
		},
		MetaFileMap: &fileSharing.MetaFileMap{
			MetaFile: make(map[string][]byte),
		},
		ChunkMap: &fileSharing.ChunkMap{
			Chunks: make(map[string][]byte),
		},
		SearchDistributeCh: g.SearchDistributeCh,
		SearchReqMap: &fileSharing.SearchReqMap{
			Map: make(map[string]bool),
		},
		Searcher: &fileSharing.Searcher{
			SendCh:  make(chan *message.SearchRequest),
			ReplyCh: make(chan *message.SearchReply),
			Target: &fileSharing.Target{
				Map: make(map[string][]fileSharing.ChunkDest),
			},
			TargetMetahash: &fileSharing.TargetMetahash{
				Map: make(map[string][]byte),
			},
			TargetMetaFile: &fileSharing.TargetMetaFile{
				Map: make(map[string][]byte),
			},
			Threshold:              2,
			InitBudget:             2,
			MaxBudget:              32,
			SearchedFileDownloadCh: make(chan *fileSharing.WrappedDownloadRequest),
		},
	}
	g.Blockchain = g.NewBlockchain()
	return
}

func main() {

	// Get input parameters
	UIPort, GuiPort, gossipAddr, name, peers, simple, hw3ex2, hw3ex3, antiEntropy, rtimer, sharedFilePath, stubbornTimeout, numPeers, ackAll = input()

	// Set up gossiper
	g := InitGossiper(UIPort, gossipAddr, name, simple, peers, antiEntropy, rtimer, sharedFilePath)

	// Start gossiper's work
	g.StartWorking()

	// TODO: Set terminating condition
	for {
		time.Sleep(10 * time.Second)
	}
	return
}
