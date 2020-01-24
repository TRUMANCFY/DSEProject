package gossiper

import (
	"sync"
	"github.com/LiangweiCHEN/Peerster/message"

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