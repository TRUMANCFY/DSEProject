package main

import (
	"fmt"
	"flag"
	"strings"
	"net"
	"strconv"
	"sync"
	"time"
	"math/rand"
	"github.com/DecentralizedSystem/Peerster/network"
	"github.com/DecentralizedSystem/Peerster/message"
)

type Gossiper struct {

	Address string
	Conn *net.UDPConn
	Name string
	UIPort string
	Peers *PeersBuffer
	Simple bool
	N *network.NetworkHandler
	RumorBuffer *RumorBuffer
	StatusBuffer *StatusBuffer
	AntiEntropyPeriod int
}

type PeerStatus struct {

	Identifier string
	NextID uint32
}

type RumorBuffer struct {

	Rumors map[string][]*message.RumorMessage
	Mux sync.Mutex
}

type StatusBuffer struct {

	Status message.StatusMap
	Mux sync.Mutex
}

type PeersBuffer struct {

	Peers []string
	Mux sync.Mutex
}
/** Global variable **/
var UIPort, gossipAddr, name string
var peers []string
var simple bool
var antiEntropy int

func input() (UIPort string, gossipAddr string, name string, peers []string, simple bool, antiEntropy int) {

	// Set flag value containers
	flag.StringVar(&UIPort, "UIPort", "8080", "UI port num")

	flag.StringVar(&gossipAddr, "gossipAddr", "127.0.0.1:5000",
					"gossip addr")

	flag.StringVar(&name, "name", "", "name of gossiper")

	var peers_str string

	flag.StringVar(&peers_str, "peers", "", "list of peers")

	flag.BoolVar(&simple, "simple", true, "Simple broadcast or not")

	flag.IntVar(&antiEntropy, "antiEntropy", 1, "antiEntroypy trigger period")
	// Conduct parameter retreival
	flag.Parse()

	// Convert peers to slice
	peers = strings.Split(peers_str, ",")

	return
}

func InitGossiper(UIPort, gossipAddr, name string, simple bool, peers []string, antiEntropy int) (g *Gossiper) {

	// Establish gossiper addr and conn
	addr, _ := net.ResolveUDPAddr("udp", gossipAddr)
	conn, _ := net.ListenUDP("udp", addr)

	// Establish client addr and conn
	client_addr, _ := net.ResolveUDPAddr("udp", UIPort)
	client_conn, _ := net.ListenUDP("udp", client_addr)
	// Create gossiper
	g = &Gossiper{
		Address : gossipAddr,
		Conn : conn,
		Client_conn : client_conn
		Name : name,
		UIPort : UIPort,
		Peers : &PeersBuffer{

			Peers : peers,
		},
		Simple : simple,
		N : &network.NetworkHandler{

			Conn : conn,
			Addr : addr,
			Send_ch : make(chan *message.PacketToSend),
			Listen_ch : make(chan *message.PacketIncome),
			Done_chs : &network.Done_chs{

				Chs : make(map[string]chan struct{}),
			},
		},
		RumorBuffer : &RumorBuffer{

			Rumors : make(map[string][]*message.RumorMessage),
		},
		StatusBuffer : &StatusBuffer{

			Status : make(message.StatusMap),
		},
		AntiEntropyPeriod : antiEntropy,
	}

	return 
}

func main() {

	// Get input parameters
	UIPort, gossipAddr, name, peers, simple, antiEntropy = input()

	// Set up gossiper
	g := InitGossiper(UIPort, gossipAddr, name, simple, peers, antiEntropy)
	
	// Start gossiper's work
	g.Start_working()

	// Send sth using network
	/*
	msgs := make([]*message.GossipPacket, 0)

	for i := 1; i < 2; i += 1 {

		msgs = append(msgs, &message.GossipPacket{
			Rumor : &message.RumorMessage{
				Origin : name,
				ID : uint32(i),
				Text : "greeting from " + name + " " + strconv.Itoa(i), 
			},
		})
	}

	g.RumorBuffer.Rumors[g.Name] = make([]*message.RumorMessage, 0)
	for i, msg := range msgs {

		g.StatusBuffer.Mux.Lock()
		g.StatusBuffer.Status[g.Name] = uint32(i + 2)
		g.RumorBuffer.Mux.Lock()
		g.RumorBuffer.Rumors[g.Name] = append(g.RumorBuffer.Rumors[g.Name], msg.Rumor)
		g.RumorBuffer.Mux.Unlock()
		g.StatusBuffer.Mux.Unlock()

		g.Peers.Mux.Lock()
		fmt.Println("Sending " + msg.Rumor.Text + " to peer " + g.Peers.Peers[0])
		g.N.Send(msg, g.Peers.Peers[0])
		g.Peers.Mux.Unlock()
	}
	*/

	// TODO: Set terminating condition
	for {
		time.Sleep(10 * time.Second)
	}
	return
}


// Gossiper start working 
func (gossiper *Gossiper) Start_working() {

	// Start network working
	gossiper.N.Start_working()

	// Start Receiving
	gossiper.Start_handling()

	// Start antiEntropy sending
	gossiper.Start_antiEntropy()
}


// Handle receiving msg
func (gossiper *Gossiper) Start_handling() {

	go func() {

		// Print peers
		fmt.Print("PEERS ")
		for i, s := range gossiper.Peers.Peers {

			fmt.Print(s)
			if i < len(gossiper.Peers.Peers) - 1 {
				fmt.Print(',')
			}
		}
		fmt.Println()
		for pkt := range gossiper.N.Listen_ch {

			// Update peers
			gossiper.UpdatePeers(pkt.Sender)

			// Start handling packet content
			switch {

			case pkt.Packet.Simple != nil:
				gossiper.HandleSimple(pkt)
			case pkt.Packet.Rumor != nil:
				gossiper.HandleRumor(pkt)
			case pkt.Packet.Status != nil:
				gossiper.HandleStatus(pkt)
			}
		}
	}()
}


// Begin antiEntropy to ensure the delivery of packets
func (gossiper *Gossiper) Start_antiEntropy() {

	go func() {

		// Set up ticker
		ticker := time.NewTicker(time.Duration(gossiper.AntiEntropyPeriod) * time.Second)

		for t := range ticker.C {

			// Drop useless tick
			_ = t

			// Conduct antiEntropy job
			gossiper.Peers.Mux.Lock()
			rand_peer := gossiper.Peers.Peers[rand.Intn(len(gossiper.Peers.Peers))]
			gossiper.Peers.Mux.Unlock()

			// Send status to selected peer
			gossiper.N.Send(&message.GossipPacket{
				Status : gossiper.StatusBuffer.ToStatusPacket(),
			}, rand_peer)
		}
	}()
}

// Update peer with given address
func (gossiper *Gossiper) UpdatePeers(peer_addr string) {

	gossiper.Peers.Mux.Lock()
	defer gossiper.Peers.Mux.Unlock()

	// Try to find peer addr in self's buffer
	for _, addr := range gossiper.Peers.Peers {

		if peer_addr == addr {
			return
		}
	}

	// Put it in self's buffer if it is absent
	gossiper.Peers.Peers = append(gossiper.Peers.Peers, peer_addr)
}
// Handle Rumor msg
func (g *Gossiper) HandleRumor(wrapped_pkt *message.PacketIncome) {

	// Print rumor received
	// fmt.Println("Rumor from " + wrapped_pkt.Sender + " with Contents: " + wrapped_pkt.Packet.Rumor.Text)
	// Decode wrapped pkt
	sender, rumor := wrapped_pkt.Sender, wrapped_pkt.Packet.Rumor

	// Update status and rumor buffer
	updated := g.Update(rumor, sender)

	// Trigger rumor mongering if it is new
	if updated {

		// Output rumor content
		fmt.Printf("RUMOR origin %s from %s ID %s contents %s\n", rumor.Origin, sender, strconv.Itoa(int(rumor.ID)), rumor.Text)
		g.MongerRumor(rumor, sender)
	}

	// Send status back
	g.N.Send(&message.GossipPacket{
		Status : g.StatusBuffer.ToStatusPacket(),
	}, sender)

}

func (g *Gossiper) Update(rumor *message.RumorMessage, sender string) (updated bool) {

	/** Compare origin and seq_id with that in self.status **/

	// Lock Status
	g.StatusBuffer.Mux.Lock()
	defer g.StatusBuffer.Mux.Unlock()

	// Initialize flag showing whether peer is known
	known_peer := false

	for origin, nextID := range g.StatusBuffer.Status {

		// Found rumor origin in statusbuffer
		if origin == rumor.Origin {

			known_peer = true

			// Update rumor buffer if current rumor is needed
			if nextID == rumor.ID {
				g.RumorBuffer.Mux.Lock()
				g.RumorBuffer.Rumors[origin] = append(g.RumorBuffer.Rumors[origin], rumor)
				g.RumorBuffer.Mux.Unlock()
				// Update StatusBuffer
				g.StatusBuffer.Status[origin] += 1

				updated = true

				fmt.Println("Receive rumor originated from " + rumor.Origin + " with ID " +
				 strconv.Itoa(int(rumor.ID)) + " relayed by " + sender)
				return
			}
		}
	}

	// Handle rumor originated from a new peer
	if rumor.ID == 1 && !known_peer {

		// Put entry for origin into self's statusBuffer
		g.StatusBuffer.Status[rumor.Origin] = 2
		// Buffer current rumor
		g.RumorBuffer.Mux.Lock()
		g.RumorBuffer.Rumors[rumor.Origin] = []*message.RumorMessage{rumor}
		g.RumorBuffer.Mux.Unlock()
		updated = true

		fmt.Println("Receive rumor originated from " + rumor.Origin + " with ID " + strconv.Itoa(int(rumor.ID)) +
		  " relayed by " + sender)
		return
	}

	// Fail to update, either out of date or too advanced
	updated = false

	return
}

func (g *Gossiper) MongerRumor(rumor *message.RumorMessage, sender string) {

	// Start rumor mongering to a neighbour apart from peer
	num_trials := 0

	// Output monger rumor signal
	fmt.Printf("MONGERING with %s\n", sender)

	for {

		// Select a random neighbour
		g.Peers.Mux.Lock()
		rand_peer_addr := g.Peers.Peers[rand.Intn(len(g.Peers.Peers))]
		g.Peers.Mux.Unlock()

		// Monger Rumor to selected peer if it is not the relayer of the rumor
		if rand_peer_addr != sender {

			//fmt.Printf("Mongering to %s with rumor originates from %s of seq id %d\n", rand_peer_addr, rumor.Origin, rumor.ID)
			g.N.Send(&message.GossipPacket{Rumor : rumor,}, rand_peer_addr)

			break
		}

		// Don't allow trying too many times
		num_trials += 1

		if num_trials > 100 {

			break
		}
	}
}

// Handle status pkt
func (g *Gossiper) HandleStatus(wrapped_pkt *message.PacketIncome) {

	// fmt.Println("Handling Status Packet")
	// Decode sender and pkt
	sender, peer_status := wrapped_pkt.Sender, wrapped_pkt.Packet.Status

	// fmt.Printf("Inside handlestatus, packet address is %p, packet details %v\n", wrapped_pkt.Packet, wrapped_pkt.Packet)
	// Convert peer_status from []PeerStatus to a map
	peer_status_map := peer_status.ToMap()

	// Ouput peer status information
	
	fmt.Printf("STATUS from %s ", sender)
	for k, v := range peer_status_map {

		fmt.Printf("peer %s nextID %s ", k, strconv.Itoa(int(v)))
	}
	fmt.Println()
	
	// TODO: stop all useless rumor mongering to this peer 
	g.StopMongering(peer_status_map, sender)

	// Case 2: send self status to peer if self is out of date

	// Case 1: send the oldest missing rumor back if peer is out of date
	moreUpdated := g.MoreUpdated(peer_status_map)

	switch moreUpdated {

	case 1:
		g.ProvideMongering(peer_status_map, sender)
	case -1:
		g.RequestMongering(peer_status_map, sender)
	default:
		// Shall filpCoinMonger, but don't know which Rumor the Status msg is replying
		fmt.Printf("IN SYNC WITH %s\n", sender)
	}


}


// Handle stoping useless mongering
func (g *Gossiper) StopMongering(peer_status_map message.StatusMap, sender string) {

	// Require lock of Done_chs 
	g.N.Done_chs.Mux.Lock()
	defer g.N.Done_chs.Mux.Unlock()

	for origin, nextID := range peer_status_map {

		// Stop mongering all pkt with ID less than nextID
		for id := 0; id < int(nextID); id += 1 {

			key := sender + origin + strconv.Itoa(id) 

			done_ch, ok := g.N.Done_chs.Chs[key]

			if ok {
				fmt.Printf("Closing %s\n", key) 
				close(done_ch)
				delete(g.N.Done_chs.Chs, key)
			}
		}
	}
}

func (g *Gossiper) MoreUpdated(peer_status message.StatusMap) (moreUpdated int) {

	// Check which of peer and self is more updated

	g.StatusBuffer.Mux.Lock()
	defer g.StatusBuffer.Mux.Unlock()

	// Loop through self status first
	for k, v := range g.StatusBuffer.Status {

		peer_v, ok := peer_status[k]
		// Return Self more updated if not ok or peer_v < v
		if !ok || peer_v < v {

			moreUpdated = 1
			return
		}
	}

	// Loop through peer to check if peer more update
	for k, v := range peer_status {

		self_v, ok := g.StatusBuffer.Status[k]

		// Return peer more updated if not ok or self_v < v
		if !ok || self_v < v {

			moreUpdated = -1
			return
		}
	}

	// Return zero if in sync state
	moreUpdated = 0
	return
}

func (rb *RumorBuffer) get(origin string, ID uint32) (rumor *message.RumorMessage) {

	rb.Mux.Lock()
	rumor = rb.Rumors[origin][ID - 1]
	rb.Mux.Unlock()
	return
}


/* Construct Status Packet from local Status Buffer */
func (sb *StatusBuffer) ToStatusPacket() (st *message.StatusPacket) {

	Want := make([]message.PeerStatus, 0)

	// fmt.Println("Requesting status buffer lock")
	sb.Mux.Lock()
	defer sb.Mux.Unlock()
	// defer fmt.Println("Releasing status buffer lock")
	for k, v := range sb.Status {

		Want = append(Want, message.PeerStatus{
			Identifier : k,
			NextID : v,
		})
	}

	st = &message.StatusPacket{
		Want : Want,
	}

	return
}


func (g *Gossiper) ProvideMongering(peer_status message.StatusMap, sender string) {

	//fmt.Println("Provide Mongering")
	// Lock status buffer for comparison
	g.StatusBuffer.Mux.Lock()
	defer g.StatusBuffer.Mux.Unlock()

	// Find the missing rumor with least ID to monger
	for k, v := range g.StatusBuffer.Status {

		switch peer_v, ok := peer_status[k]; {

		// Send the first rumor from current origin if peer have not heard
		// from it
		case !ok:
			g.N.Send(&message.GossipPacket{
				Rumor : g.RumorBuffer.get(k, 1),
			}, sender)
			// fmt.Printf("Mongering to %s with rumor originates from %s of seq id %d\n", sender, k, 1)
			return
		case peer_v < v:
			g.N.Send(&message.GossipPacket{
				Rumor : g.RumorBuffer.get(k, peer_v),
			}, sender)
			// fmt.Printf("Mongering to %s with rumor originates from %s of seq id %d\n", sender, k, peer_v)
			return
		}
	}
}


func (g *Gossiper) RequestMongering(peer_status message.StatusMap, sender string) {

	g.N.Send(&message.GossipPacket{
		Status : g.StatusBuffer.ToStatusPacket(),
	}, sender)
}


// Handle Simple Message
func (gossiper *Gossiper) HandleSimple(wrapped_pkt *message.PacketIncome) {

	// Construct rumor message
	packet := wrapped_pkt.Packet

	// Output msg content
	fmt.Printf("CLIENT MESSAGE %s\n", packet.Simple.Contents)
	gossiper.RumorBuffer.Mux.Lock()
	defer gossiper.RumorBuffer.Mux.Unlock()

	gossiper.RumorBuffer.Rumors[gossiper.Name] = append(gossiper.RumorBuffer.Rumors[gossiper.Name], 
														&message.RumorMessage{

															Origin : gossiper.Name,
															ID : uint32(len(gossiper.RumorBuffer.Rumors[gossiper.Name]) + 1),
															Text : packet.Simple.Contents,
														})

	// Trigger rumor mongering
	gossiper.Peers.Mux.Lock()
	defer gossiper.Peers.Mux.Unlock()

	rand_peer := gossiper.Peers.Peers[rand.Intn(len(gossiper.Peers.Peers))]
	gossiper.MongerRumor(gossiper.RumorBuffer.Rumors[gossiper.Name][len(gossiper.RumorBuffer.Rumors[gossiper.Name]) - 1],
						rand_peer)

}
