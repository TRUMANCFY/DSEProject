package main

import (
	"fmt"
	"flag"
	"net"
	"os"
	"strings"
	"encoding/hex"
	"github.com/dedis/protobuf"
	"github.com/LiangweiCHEN/Peerster/message"
)


/* Struct definition */
func input() (UIPort string, msg string, dest string, file, request, keywords string, budget int,
				voterid, vote, electionName string) {

	// Set cmd flag value containers
	flag.StringVar(&UIPort, "UIPort", "8080", "UI port number")

	flag.StringVar(&msg, "msg", "", "Msg to be sent")

	flag.StringVar(&dest, "dest", "", "Private Msg Destination")

	flag.StringVar(&file, "file", "", "File to be indexed")

	flag.StringVar(&request, "request", "", "metahash of the file to be requested")

	flag.StringVar(&keywords, "keywords", "", "Keywords of file title to search")

	flag.IntVar(&budget, "budget", 2, "Initial budget of expanding ring search")

	// Parameter for votes
	flag.StringVar(&voterid, "Voterid", "", "Voter id in string")
	flag.StringVar(&vote, "Vote", "", "Vote in string")
	flag.StringVar(&electionName, "Election", "", "Election name")
	// Parse cmd values
	flag.Parse()

	return
}

func main() {

	UIPort, msg, dest, file, request, keywords, budget, voterid, vote, electionName := input()

	// Handle invalid user input
	switch {
	case msg != "" && dest != "" && file == "" && request == "" && keywords == "": // Private msg
	case msg == "" && dest == "" && file != "" && request == "" && keywords == "": // Share file instruction
	case msg == "" && dest != "" && file != "" && request != "" && keywords == "": // Download file instruction
	case msg != "" && dest == "" && file == "" && request == "" && keywords == "": // Gossip msg
	case msg == "" && dest == "" && file == "" && request == "" && keywords != "" && budget > 0: // Search file instruction
	case msg == "" && dest == "" && file != "" && request != "" && keywords == "": // Download searched file instruction?
	case voterid != "" && vote != "" && electionName != "":							// Blockchain
	default:
		fmt.Printf("ERROR (Bad argument combination)")
		os.Exit(1)
	}

	// Create dst address
	dst_addr, _ := net.ResolveUDPAddr("udp4", ":" + UIPort)

	// Create UDP 'connection'
	conn, _ := net.DialUDP("udp4", nil, dst_addr)

	defer conn.Close()

	// Create a gossiper msg
	var destPtr, filePtr *string
	var requestPtr *[]byte
	if dest == ""{
		destPtr = nil
	} else {
		destPtr = &dest
	}

	if file == "" {
		filePtr = nil
	} else {
		filePtr = &file
	}

	requestBytes := make([]byte, 32)
	// fmt.Println(request)
	requestBytes, err := hex.DecodeString(request)
	if err != nil {
		// fmt.Println(err)
		fmt.Printf("ERROR (Unable to decode hex hash)")
		os.Exit(1)
		
	}
	// fmt.Println(len(requestBytes))
	if request == "" {
		requestPtr = nil
	} else {
		requestPtr = &requestBytes
	}

	// Get the keywords slice
	var keyword_slice []string
	if len(keywords) > 0 {
		keyword_slice = strings.Split(keywords, ",")
	} else {
		keyword_slice = make([]string, 0)
	}
	fmt.Printf("Keywords are %s\n", strings.Join(keyword_slice, ","))
	fmt.Printf("Voterid %s Vote %s\n", voterid, vote)
	pkt := &message.Message{
		Text : msg,
		Destination : destPtr,
		File : filePtr,
		Request : requestPtr,
		Keywords : keyword_slice,
		Budget : uint64(budget),
		Voterid : voterid,
		Vote : vote,
		ElectionName : electionName,
	}

	// Encode the msg
	msg_bytes, err := protobuf.Encode(pkt)

	if err != nil {

		fmt.Println(err)
	}

	// Send the msg to the server
	fmt.Println("Sending to gossiper")
	conn.Write(msg_bytes)

	return
}