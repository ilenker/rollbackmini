package main

import (
	"fmt"
	"net"
	"log"
	"strings"
	"strconv"
)

var Debug bool

type PeerPacket struct {
	frameID uint16
	content [4]byte
}

/*

r u l d, h, p

H M

123456:LRS_

123457:L___

123458:____

123450:H___
123450:M___

*/


func multiplayer(inboundInputs, inboundReplies, outboundPackets chan PeerPacket) {
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

	// This is just a local signal for now
	// to unblock main thread when both peers are ready
	inboundInputs <-PeerPacket{frameID:6969}
	
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
	rdvConn.WriteToUDP([]byte("65535=6969"), premote)


	// Inbound loop thread
	go listenToPort(rdvConn, inboundInputs, inboundReplies)

	fmt.Printf(" >> Listening...\n\n")
	fmt.Println("--- Launching SnakeCycles ---")
	fmt.Println("---     <esc> to quit     ---")


	// Outbound loop
	for {
		// From main thread...
		pP := <-outboundPackets
		if pP.frameID < 50 {
			continue
		}

		input := peerPacketToBytes(pP)

		// ...out to peer.
		n, err := rdvConn.WriteToUDP([]byte(input), premote)
		if err != nil { fmt.Printf("(main)sending [%d bytes] failed: %v\n", n, err) }
	}

}


func listenToPort(conn *net.UDPConn, inboundInputs, inboundReplies chan PeerPacket) error {

	defer conn.Close()

	b := make([]byte, 64)

	for {
		n, _, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(listener)read error: %v\n", err) }

		assert(n, 10, "n", "expectedPacketLen")

		peerPacket := bytesToPeerPacket(b[:n])

		inboundInputs  <-peerPacket
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
			return peerPublicEndpoint, peerPrivateEndpoint
		}

		n, _, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(rdv-reply)read error: %v\n", err) }

		if len(b) > 1 {

			data, found := strings.CutPrefix(string(b[:n]), "peerPublicEndpoint:")
			if found {
				peerPublicEndpoint = data
				continue
			}

			data, found = strings.CutPrefix(string(b[:n]), "peerPrivateEndpoint:") 
			if found {
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

//  {frameID: 12345, [L, R, _, _]} -> "12345=LR__"
func peerPacketToBytes(p PeerPacket) []byte {
	s := fmt.Sprintf("%05d=%1c%1c%1c%1c",
		p.frameID,
		p.content[0],
		p.content[1],
		p.content[2],
		p.content[3])
	assert(len(s), 10, "len(s)", "10")
	return []byte(s)
}

// "12345=LR__" -> {frameID: 12345, [L, R, _, _]}
func bytesToPeerPacket(b []byte) PeerPacket {
	assert(len(b), 10, "len(b)", "10")
	s := string(b)
	split := strings.Split(s, "=")

	frameID, err := strconv.Atoi(split[0]); F(err, "bytesToPeerPacket:ID")

	packet := PeerPacket{
		frameID: uint16(frameID),
		content: [4]byte{
			split[1][0],
			split[1][1],
			split[1][2],
			split[1][3],
		},
	}

	return packet
}

func makePeerPacket(frameID uint16, content []signal) PeerPacket {

	pP := PeerPacket{
		frameID: frameID,
		content: [4]byte{'_', '_', '_', '_'},
	}

	for i, b := range content {
		pP.content[i] = byte(b)
	}

	return pP
}
