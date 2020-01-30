package gossiper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"math/rand"

	"github.com/TRUMANCFY/DSEProject/Peerster/message"
	"github.com/TRUMANCFY/DSEProject/Peerster/routing"

	//"go.dedis.ch/kyber/util/random"
	"go.dedis.ch/kyber/group/edwards25519"
)

func (g *Gossiper) ForwardPkt(pkt *message.GossipPacket, dest string) (err routing.RoutingErr) {
	// Find next hop for destination and forward the packet to next hop

	g.Dsdv.Mux.Lock()
	nextHop, ok := g.Dsdv.Map[dest]
	if !ok {
		err = routing.NewRoutingErr(dest)
		return
	}
	g.Dsdv.Mux.Unlock()
	g.N.Send(pkt, nextHop)
	return
}

func (g *Gossiper) SelectRandomPeer(excluded []string, n int) (rand_peer_addr []string, ok bool) {
	// Select min(n, max_possible) random peers with peers in excluded slice excluded

	g.Peers.Mux.Lock()
	defer g.Peers.Mux.Unlock()

	// Get slice of availble peers
	availablePeers := make([]string, 0)
	excludedMap := make(map[string]bool)
	for _, peer := range excluded {
		excludedMap[peer] = true
	}
	for _, peer := range g.Peers.Peers {
		if _, ok := excludedMap[peer]; !ok {
			availablePeers = append(availablePeers, peer)
		}
	}

	// Return if no available peer exists
	if len(availablePeers) == 0 {
		ok = false
		return
	}

	// Get min(n, len(available_peers)) random neighbour
	selected := make(map[int]bool)
	if n > len(availablePeers) {
		rand_peer_addr = availablePeers
		ok = true
	} else {
		count := 0
		for {
			next := rand.Intn(len(availablePeers))
			if _, ok := selected[next]; !ok {
				selected[next] = true
				count += 1
				rand_peer_addr = append(rand_peer_addr, availablePeers[next])
			}
			if count == n {
				ok = true
				return
			}
		}
	}
	return
}

func (gossiper *Gossiper) UpdatePeers(peerAddr string) {
	// Record new peer with given addr

	gossiper.Peers.Mux.Lock()
	defer gossiper.Peers.Mux.Unlock()

	// Try to find peer addr in self's buffer
	for _, addr := range gossiper.Peers.Peers {
		if peerAddr == addr {
			return
		}
	}

	// Put it in self's buffer if it is absent
	gossiper.Peers.Peers = append(gossiper.Peers.Peers, peerAddr)
}

func (g *Gossiper) MoreUpdated(peer_status message.StatusMap) (moreUpdated int) {
	// Check which of peer and self is more updated
	// Step 1. Loop through self's status to check whether self is more updated
	// Step 2. Loop thourgh peer's status to check whether peer is more updated
	g.StatusBuffer.Mux.Lock()
	defer g.StatusBuffer.Mux.Unlock()

	/* Step 1 */
	for k, v := range g.StatusBuffer.Status {
		peer_v, ok := peer_status[k]
		// Return Self more updated if not ok or peer_v < v
		if !ok || peer_v < v {
			moreUpdated = 1
			return
		}
	}

	/* Step 2 */
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

func (rb *RumorBuffer) get(origin string, ID uint32) (rumor *message.WrappedRumorTLCMessage) {
	// Get the rumor or tlc with corresponding id from specified origin
	rb.Mux.Lock()
	rumor = rb.Rumors[origin][ID-1]
	rb.Mux.Unlock()
	return
}

func (sb *StatusBuffer) ToStatusPacket() (st *message.StatusPacket) {
	// Construct status packet from local status buffer
	// It basically convert map to slice of peer status
	Want := make([]message.PeerStatus, 0)
	sb.Mux.Lock()
	defer sb.Mux.Unlock()
	for k, v := range sb.Status {
		Want = append(Want, message.PeerStatus{
			Identifier: k,
			NextID:     v,
		})
	}
	st = &message.StatusPacket{
		Want: Want,
	}
	return
}

func (g *Gossiper) PrintPeers() {

	outputString := fmt.Sprintf("PEERS ")

	for i, s := range g.Peers.Peers {

		outputString += fmt.Sprintf(s)
		if i < len(g.Peers.Peers)-1 {
			outputString += fmt.Sprintf(",")
		}
	}
	outputString += fmt.Sprintf("\n")
	fmt.Print(outputString)
}

func (g *Gossiper) Update(wrappedMessage *message.WrappedRumorTLCMessage, sender string) (updated bool) {
	// This function attempt to update local cache of messages by comparing
	// the incoming msg's id with local expected.
	// It resolve the blockchain proposal update by checking the confirmed field of TLC message.

	g.StatusBuffer.Mux.Lock()
	defer g.StatusBuffer.Mux.Unlock()
	known_peer := false
	isRumor := wrappedMessage.RumorMessage != nil
	var inputID uint32
	var inputOrigin string
	if isRumor {
		inputID = wrappedMessage.RumorMessage.ID
		inputOrigin = wrappedMessage.RumorMessage.Origin
	} else if wrappedMessage.TLCMessage != nil {
		inputID = wrappedMessage.TLCMessage.ID
		inputOrigin = wrappedMessage.TLCMessage.Origin
	} else {
		inputID = wrappedMessage.BlockRumorMessage.ID
		inputOrigin = wrappedMessage.BlockRumorMessage.Origin
	}

	for origin, nextID := range g.StatusBuffer.Status {

		// Found rumor origin in statusbuffer
		if origin == inputOrigin {

			known_peer = true
			// Input is what is expected !!!
			if nextID == inputID {

				// Update rumor buffer
				g.RumorBuffer.Mux.Lock()
				g.RumorBuffer.Rumors[origin] = append(g.RumorBuffer.Rumors[origin],
					wrappedMessage)
				g.RumorBuffer.Mux.Unlock()

				// Update StatusBuffer
				g.StatusBuffer.Status[origin] += 1
				updated = true

				//fmt.Println("Receive rumor originated from " + rumor.Origin + " with ID " +
				// strconv.Itoa(int(rumor.ID)) + " relayed by " + sender)

				return
			}
		}
	}
	// Handle rumor originated from a new peer
	if inputID == 1 && !known_peer {

		// Put entry for origin into self's statusBuffer
		g.StatusBuffer.Status[inputOrigin] = 2
		// Buffer current rumor
		g.RumorBuffer.Mux.Lock()
		g.RumorBuffer.Rumors[inputOrigin] = []*message.WrappedRumorTLCMessage{wrappedMessage}
		g.RumorBuffer.Mux.Unlock()
		updated = true

		// fmt.Println("Receive rumor originated from " + rumor.Origin + " with ID " + strconv.Itoa(int(rumor.ID)) +
		//   " relayed by " + sender)
		return
	}
	// Fail to update, either out of date or too advanced
	updated = false
	return
}

func (g *Gossiper) StartAuthentication(auth string) {
	/*
		This func build DES for gossiper
	*/
	fmt.Println("start authentiation")
	g.Auth = NewAuth(auth)
}

func NewAuth(auth string) (authenticator *Auth) {
	/*
		This function create a new authenticator for NIZKF
	*/

	// Initialize random stream
	//rng := random.New()

	// Initialize random stream for G and H
	plainText := []byte(auth[:16])
	block, err := aes.NewCipher(plainText)
	if err != nil {
		fmt.Println(err)
		return
	}
	ivString := "aaaaaaaaaaaaaaaa"
	iv := []byte(ivString)
	stream := cipher.NewCFBEncrypter(block, iv)

	// Create new suite
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create secret
	authBytes := []byte(auth)
	scal := sha256.Sum256(authBytes[:])
	x := suite.Scalar().SetBytes(scal[:32])

	// Create G and H for NIZKF
	G := suite.Point().Pick(stream)
	H := suite.Point().Pick(stream)
	fmt.Printf("G IS %s\n", G)
	// Create authenticator
	authenticator = &Auth{
		auth:  auth,
		X:     x,
		G:     G,
		H:     H,
		XG:    suite.Point().Mul(x, G),
		XH:    suite.Point().Mul(x, H),
		Suite: suite,
	}

	return
}

func (a *Auth) Provide() (proof *message.Proof) {
	/*
		This function provide a proof for holding the secret in a non-interative manner
		Step 1. Select random scalr v and compute vG, vH
		Step 2. Create challange c
		Step 3. Compute r = v - xc where v is a randomly selected scalar
						rG
						rH
	*/
	fmt.Println("check point 0")
	/* Step 1 */
	v := a.Suite.Scalar().Pick(a.Suite.RandomStream())
	vG := a.Suite.Point().Mul(v, a.G)
	vH := a.Suite.Point().Mul(v, a.H)
	fmt.Println("check point 1")
	/* Step 2 */
	h := a.Suite.Hash()
	a.XG.MarshalTo(h)
	a.XH.MarshalTo(h)
	vG.MarshalTo(h)
	vH.MarshalTo(h)
	cb := h.Sum(nil)
	c := a.Suite.Scalar().Pick(a.Suite.XOF(cb))
	fmt.Println("check point 2")
	/* Step 3 */
	r := a.Suite.Scalar()
	r.Mul(a.X, c).Sub(v, r)

	// Encode point and scalar to bytes

	rBytes, _ := r.MarshalBinary()
	vGBytes, _ := vG.MarshalBinary()
	vHBytes, _ := vH.MarshalBinary()
	proof = &message.Proof{
		R:  rBytes,
		VG: vGBytes,
		VH: vHBytes,
	}
	fmt.Println("check point 3")
	return
}

func (a *Auth) Verify(proof *message.Proof) (valid bool) {
	/*
		This func check the validity of proof
		Step 1. Compute c from xG, xH, vG, vH
		Step 2. Compute cxG and cxH, rG and rH
		Step 3. Verify vG = cxG + rG and vH = cxH + rH
	*/

	if proof == nil {
		return false
	}
	a.Mux.Lock()
	r := a.Suite.Scalar()
	r.UnmarshalBinary(proof.R)
	vG := a.Suite.Point()
	vG.UnmarshalBinary(proof.VG)
	vH := a.Suite.Point()
	vH.UnmarshalBinary(proof.VH)

	/* Step 1 */
	h := a.Suite.Hash()
	a.XG.MarshalTo(h)
	a.XH.MarshalTo(h)
	vG.MarshalTo(h)
	vH.MarshalTo(h)
	cb := h.Sum(nil)
	c := a.Suite.Scalar().Pick(a.Suite.XOF(cb))
	/* Step 2 */
	cxG := a.Suite.Point().Mul(c, a.XG)
	cxH := a.Suite.Point().Mul(c, a.XH)
	rG := a.Suite.Point().Mul(r, a.G)
	rH := a.Suite.Point().Mul(r, a.H)

	/* Step 3 */
	resultG := a.Suite.Point().Add(rG, cxG)
	resultH := a.Suite.Point().Add(rH, cxH)

	if !(vG.Equal(resultG) && vH.Equal(resultH)) {
		fmt.Printf("Incorrect proof!\n")
		valid = false
	} else {
		fmt.Printf("Correct proof")
		valid = true
	}
	a.Mux.Unlock()
	return
}
