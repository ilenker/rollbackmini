package main

import (
	"fmt"
	"net"
	"log"
	"strings"
	"strconv"
	"golang.org/x/term"
)

var Debug bool

type TermInfo struct {
	fd int
	Cols int
	Rows int
	oldState *term.State
}

type PeerPacket struct {
	frameID uint16
	inputQ [4]input
}


func multiplayer(inbound, outbound chan PeerPacket) {
	version := "v0.1"

	fmt.Println("==== Rollback Microgame P2P ====")
	fmt.Printf("                           %s   \n", version)
	fmt.Println(" >> connecting to rendezvous")

	// Rendezvous server address
	rdvAddr := net.UDPAddr{
		IP: net.ParseIP("34.172.225.134"),
		Port: 55585,
	}

	Debug = false

	localIP := GetOutboundIP()
	laddr, err := net.ResolveUDPAddr("udp4", localIP.String() + ":0")
	if err != nil { fmt.Printf("(rdv)address parse failed: %v\n", err) }

	 // Bind Source Port (Listen)
	rdvConn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		fmt.Printf("(rdv)binding failed: %v\n", err)
		rdvConn.Close()
		return
	}

	 // Wait for peer connection + information from rendezvous 
	peerPubIP, _ := waitForRdvReply(rdvConn, &rdvAddr)
	inbound <-PeerPacket{frameID:6969}
	
	fmt.Printf(" >> Peer found: [%s]\n", peerPubIP)

	 // After Server Connect
     // Punch hole
	premote, err := net.ResolveUDPAddr("udp4", peerPubIP)
	if err != nil {
		fmt.Printf("(punch)address parse failed: %v\n", err)
		rdvConn.Close()
		return
	}

	fmt.Printf(" >> punching hole\n")

	rdvConn.WriteToUDP([]byte("6969=0000"), premote)

	go listenToPort(rdvConn, inbound)

	fmt.Printf(" >> Listening...\n\n")
	fmt.Println("--- Launching SnakeCycles ---")
	fmt.Println("---     <esc> to quit     ---")


	for {
		pP := <-outbound
		if pP.frameID < 50 {
			continue
		}
		input := peerPacketToBytes(pP)

		n, err := rdvConn.WriteToUDP([]byte(input), premote)
		if err != nil { fmt.Printf("(main)sending [%d bytes] failed: %v\n", n, err) }
	}

}


func listenToPort(conn *net.UDPConn, inbound chan PeerPacket) error {

	defer conn.Close()

	b := make([]byte, 512)

	for {
		n, _, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(listener)read error: %v\n", err) }

		assert(n, 9, "n", "expectedPacketLen")

		inbound <-bytesToPeerPacket(b[:n])
	}
}


func waitForRdvReply(conn *net.UDPConn, rdvAddr *net.UDPAddr) (string, string) {
	peerPublicEndpoint := ""
	peerPrivateEndpoint := ""
	fmt.Printf(" >> waiting for rendezvous\n")

	b := make([]byte, 65507)

	// Here, instead of sending funny number, we will send something *useful*
    // We need to send the endpoint we believe 
	// we are using to communicate with the server.
	privEndpoint := conn.LocalAddr().String()
	conn.WriteToUDP([]byte(privEndpoint), rdvAddr)

	for {
		if peerPublicEndpoint != "" && peerPrivateEndpoint != "" {
			fmt.Printf("if peerPublicEndpoint != \"\" && peerPrivateEndpoint != \"\"\n")
			return peerPublicEndpoint, peerPrivateEndpoint
		}

		n, _, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(rdv-reply)read error: %v\n", err) }

		if len(b) > 1 {
			fmt.Printf("if len(b) > 1\n")

			data, found := strings.CutPrefix(string(b[:n]), "peerPublicEndpoint:")
			fmt.Printf("####")
			if found {
				fmt.Printf("if found { PUB\n")
				peerPublicEndpoint = data
				continue
			}

			data, found = strings.CutPrefix(string(b[:n]), "peerPrivateEndpoint:") 
			if found {
				fmt.Printf("if found { PRIV\n")
				peerPrivateEndpoint = data
			}

		}

	}
}


func GetOutboundIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}

func peerPacketToBytes(p PeerPacket) []byte {
	s := fmt.Sprintf("%04d=%1d%1d%1d%1d",
		p.frameID,
		p.inputQ[0],
		p.inputQ[1],
		p.inputQ[2],
		p.inputQ[3])
	assert(len(s), 9, "len(s)", "9")
	return []byte(s)
}

func bytesToPeerPacket(b []byte) PeerPacket {
	assert(len(b), 9, "len(b)", "9")
	s := string(b)
	split := strings.Split(s, "=")

	frameID, err := strconv.Atoi(split[0]); F(err, "bytesToPeerPacket:ID")

	input0, err := strconv.Atoi(string(split[1][0])); F(err, "bytesToPeerPacket:0")
	input1, err := strconv.Atoi(string(split[1][1])); F(err, "bytesToPeerPacket:1")
	input2, err := strconv.Atoi(string(split[1][2])); F(err, "bytesToPeerPacket:2")
	input3, err := strconv.Atoi(string(split[1][3])); F(err, "bytesToPeerPacket:3")

	packet := PeerPacket{
		frameID: uint16(frameID),
		inputQ: [4]input{
			input(input0),
			input(input1),
			input(input2),
			input(input3),
		},
	}

	return packet
}

func makePeerPacket(frameID uint16, snake *Snake) PeerPacket {

	inputQ := [4]input{
		iNone,
		iNone,
		iNone,
		iNone,
	} 

	copy(inputQ[:], snake.inputQ)


	pP := PeerPacket{
		frameID: frameID,
		inputQ: inputQ, 
	}

	return pP
}
