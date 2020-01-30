package gossiper

import (
	"sync"
	"github.com/TRUMANCFY/DSEProject/Peerster/message"
	"go.dedis.ch/kyber/group/edwards25519"
	"go.dedis.ch/kyber"
)

type PeerStatus struct {

	Identifier string
	NextID uint32
}

type RumorBuffer struct {

	Rumors map[string][]*message.WrappedRumorTLCMessage
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

type PeerStatusAndSync struct {

	PeerStatus *message.PeerStatus
	IsSync bool
}

type Ack_chs struct {

	Chs map[string]chan *PeerStatusAndSync
	Mux sync.Mutex
}

type PeerStatuses struct {

	Map map[string]map[string]uint32
	Mux sync.Mutex
}

type TLCAckChs struct {

	Chs map[uint32]chan []string
	Mux sync.Mutex
}

type TLCClock struct {
	
	Clock map[string]int 
	Map map[string]map[uint32]int
	Mux sync.Mutex
}

type WrappedTLCMessage struct {
	TLCMessage *message.TLCMessage
	Round int
}

type MsgBuffer struct {
	Msg []string
	Mux sync.Mutex
}

type Auth struct {
	auth string			// Authentication message
	X kyber.Scalar		// Secret of NIZKF
	G kyber.Point
	H kyber.Point
	XG kyber.Point		// xG for NIZKF
	XH kyber.Point		// xH for NIZKF
	Suite *edwards25519.SuiteEd25519 // Suite for NIZKF
	Mux sync.Mutex
}