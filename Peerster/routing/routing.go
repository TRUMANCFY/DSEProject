package routing

import (
	"fmt"
	"sync"
)

type NextHop string 

type DSDV struct {

	Map map[string]string
	Mux sync.Mutex
	Ch chan *OriginRelayer
}

type OriginRelayer struct {

	Origin string
	Relayer string
	HeartBeat bool 
}

func (router *DSDV) StartRouting() {
	// Get entry from channel
	// Update DSDV

	go func() {

		for pair := range router.Ch {

			origin, relayer, heartbeat := pair.Origin, pair.Relayer, pair.HeartBeat

			router.Mux.Lock()
			router.Map[origin] = relayer
			router.Mux.Unlock()

			if !heartbeat {
				fmt.Printf("DSDV %s %s\n", origin, relayer)
			}
		}
	}()
}
