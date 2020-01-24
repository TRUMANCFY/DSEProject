package network

import (

	"net"
	"fmt"
	"sync"
	//"strconv"
	//"time"
	// "encoding/hex"
	"github.com/dedis/protobuf"
	// "encoding/base64"
	"github.com/LiangweiCHEN/Peerster/message"
)

type NetworkHandler struct {

	Conn *net.UDPConn
	Client_conn *net.UDPConn
	Addr *net.UDPAddr
	Send_ch chan *message.PacketToSend
	Listen_ch chan *message.PacketIncome
	Client_listen_ch chan *message.Message
	Done_chs *Done_chs
	RumorTimeoutCh chan *message.PacketToSend
}

type Done_chs struct {

	Chs map[string]chan struct{}
	Mux sync.Mutex
}

func (n *NetworkHandler) Send(pkt *message.GossipPacket, dst string) {

	// Build pkt to send
	pkt_to_send := message.PacketToSend{

		Packet: pkt,
		Addr: dst,
	}

	// Pass pkt to send to send_ch
	// fmt.Printf("%p\n", pkt_to_send)
	n.Send_ch <- &pkt_to_send

	return
}


func (n *NetworkHandler) StartSending() {
	
	// Get pkt to send from send_ch and 
	// create go routine to send it

	for pkt_to_send := range n.Send_ch {

		pkt_to_send := pkt_to_send

		// Localize pkt and addr
		pkt, err := protobuf.Encode(pkt_to_send.Packet)

		if err != nil {

			fmt.Print(err)
			return
		}

		addr, err := net.ResolveUDPAddr("udp", pkt_to_send.Addr)

		if err != nil {

			fmt.Print(err)
			return
		} 

		n.Conn.WriteToUDP(pkt, addr)
		
	}
	
}


func (n *NetworkHandler) StartListening() {

	// Create buffer and pkt container

	// Listen
	for {
		buffer := make([]byte, 9 * 1024)
		packet := new(message.GossipPacket)
		// fmt.Println("Listening")
		// Try to collect encoded pkt
		size, addr, err := n.Conn.ReadFromUDP(buffer)

		// fmt.Println("Receiving " + strconv.Itoa(size))
		if err != nil {

			fmt.Print(err)
			return
		}

		// Decode pkt
		protobuf.Decode(buffer[: size], packet)

		// Output packet for testing
		//fmt.Printf("CLIENT MESSAGE %s\n", packet.Rumor.Text)

		/*
		if packet.DataRequest != nil {

			fmt.Printf("RECEIVE REQUEST %s", base64.URLEncoding.EncodeToString(packet.DataRequest.HashValue))
		}
		*/
		// Put pkt into listen channel
		n.Listen_ch <- &message.PacketIncome{
			Packet : packet,
			Sender : addr.String(),
		}
	}

	fmt.Println("Finish listening")
}


func (n *NetworkHandler) StartListeningClient() {
	// Create buffer and pkt container

	// Listen
	for {
		buffer := make([]byte, 8 * 1024)
		packet := new(message.Message)
		// fmt.Println("Listening")
		// Try to collect encoded pkt
		size, _, err := n.Client_conn.ReadFromUDP(buffer)

		// fmt.Println("Receiving " + strconv.Itoa(size))
		if err != nil {

			fmt.Print(err)
			return
		}

		// Decode pkt
		protobuf.Decode(buffer[: size], packet)

		// Output packet for testing
		// fmt.Println(packet.Request)
		// fmt.Printf("CLIENT MESSAGE %s\n", packet.Simple.Contents)

		// Put pkt into listen channel
		n.Client_listen_ch <- packet
		// fmt.Println("Successfully put client msg into channel")
	}

	// fmt.Println("Finish listening")

}
func (n *NetworkHandler) StartWorking() {

	go n.StartListening()
	go n.StartListeningClient()
	go n.StartSending()
}
