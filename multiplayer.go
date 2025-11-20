package main

import (
	"fmt"
	"net"
	"log"
	"time"
	"strings"
	"strconv"
)


type PeerPacket struct {
	frameID uint16
	content [4]signal
}

var pingTimes map[uint16]time.Time

var rttBuffer func(time.Duration) (int64, []time.Duration)
var avgRTTuSec int64
var RTTs []time.Duration

var frameDiffGraph func(int)


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
	inputFromPeerCh <-PeerPacket{}
	
	fmt.Printf(" >> Peer found: [%s]\n", peerPubIP)
	fmt.Printf(" >> Listening...\n\n")
	fmt.Println("--- Launching SnakeCycles ---")
	fmt.Println("---     <esc> to quit     ---")

	// After Server Connect
	// Punch hole
	premote, err := net.ResolveUDPAddr("udp4", peerPubIP)
	if err != nil {
		fmt.Printf("(punch)address parse failed: %v\n", err)
		rdvConn.Close()
		return
	}


	// Inbound loop thread
	pingTimes = make(map[uint16]time.Time)
	rttBuffer = makeAverageDurationBuffer(RTT_BUFFER_LEN)

	go listenToPort(rdvConn, inboundInputs, inboundReplies, outboundPackets)
	go sendPings(outboundPackets)


	// Outbound loop
	for {
		// From main thread...
		pP := <-outboundPackets

		input := peerPacketToBytes(pP)

		// ...out to peer.
		n, err := rdvConn.WriteToUDP([]byte(input), premote)
		if err != nil { fmt.Printf("(main)sending [%d bytes] failed: %v\n", n, err) }
	}

}


func listenToPort(conn *net.UDPConn, inboundInputs, inboundReplies, outboundPackets chan PeerPacket) error {

	defer conn.Close()

	b := make([]byte, 64)

	for {
		n, _, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(listener)read error: %v\n", err) }

		assert(n, 10, "n", "expectedPacketLen")

		peerPacket := bytesToPeerPacket(b[:n])


		switch peerPacket.content[0] {
		case iHit:
			inboundReplies <-peerPacket
		case iCrit:
			inboundReplies <-peerPacket
		case iMiss:
			inboundReplies <-peerPacket

		case iPong:
			avgRTTuSec, RTTs = rttBuffer(processPong(peerPacket))

		case iPing:
			outboundPackets <-PeerPacket{
				peerPacket.frameID,
				[4]signal{iPong},
			}

		default:
			avgFrameDiff, frameDiffs = FrameDiffBuffer(int(SIM_FRAME - peerPacket.frameID))
			inboundInputs <-peerPacket
		}

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

func bytesToPeerPacket(b []byte) PeerPacket {
	assert(len(b), 10, "len(b)", "10")
	s := string(b)
	split := strings.Split(s, "=")

	frameID, err := strconv.Atoi(split[0]); F(err, "bytesToPeerPacket:ID")

	packet := PeerPacket{
		frameID: uint16(frameID),
		content: [4]signal{
			signal(split[1][0]),
			signal(split[1][1]),
			signal(split[1][2]),
			signal(split[1][3]),
		},
	}

	return packet
}

func makePeerPacket(frameID uint16, content [4]signal) PeerPacket {

	pP := PeerPacket{
		frameID: frameID,
		content: content,
	}

	return pP
}

func sendCurrentFrameInputs(input signal) {
	if input != iNone {
		callsBox(fmt.Sprintf("send(%03X, %c)       \n", SIM_FRAME, input))
	}

	if online {
		select {
		case packetsToPeerCh <-makePeerPacket(SIM_FRAME, [4]signal{input}):
			return
		default:	
			return
		}
	}  

}


func sendPings(outboundPackets chan PeerPacket) {
	pingTicker := time.NewTicker(time.Millisecond * 100)

	for {
		<-pingTicker.C
		currentFrame := SIM_FRAME
		outboundPackets <-PeerPacket{currentFrame, [4]signal{iPing}}
		pingTimes[currentFrame] = time.Now()
	}

}


func processPong(pP PeerPacket) time.Duration {
	sentTime, _ := pingTimes[pP.frameID]
	
	return time.Since(sentTime)
}


