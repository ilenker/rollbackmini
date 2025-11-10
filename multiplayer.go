package main

import (
	"fmt"
	"net"
	"log"
	"strings"
	"golang.org/x/term"
)

var Debug bool

type TermInfo struct {
	fd int
	Cols int
	Rows int
	oldState *term.State
}


func connect() {
	version := "v0.1"

	fmt.Println("====   Snakecycles P2P    ====")
	fmt.Printf("                           %s   \n", version)
	fmt.Println(" >> connecting to rendezvous")
	fmt.Println("    please waiting          ")

	// Rendezvous server address
	rdvAddr := net.UDPAddr{
		IP: net.ParseIP("34.172.225.134"),
		Port: 55585,
	}

	Debug = true

	localIP := GetOutboundIP()
	laddr, err := net.ResolveUDPAddr("udp4", localIP.String() + ":0")
	if err != nil { fmt.Printf("(rdv)address parse failed: %v\n", err) }

	 // Bound Source Port (Listen)
	rdvConn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		fmt.Printf("(rdv)binding failed: %v\n", err)
		rdvConn.Close()
		return
	}

	 // Wait for peer connection + information from rendezvous 
	peerPubIP, _ := waitForRdvReply(rdvConn, &rdvAddr)
	
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

	rdvConn.WriteToUDP([]byte("punch"), premote)

	go listenToPort(rdvConn)

	fmt.Printf(" >> Listening...\n\n")
	fmt.Println("--- Launching SnakeCycles ---")
	fmt.Println("---     <esc> to quit     ---")


	for {
		n, err := rdvConn.WriteToUDP([]byte(input), premote)
		if err != nil { fmt.Printf("(main)sending [%d bytes] failed: %v\n", n, err) }

	}

}


func listenToPort(conn *net.UDPConn) error {

	defer conn.Close()

	b := make([]byte, 512)

	for {
		n, addr, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(listener)read error: %v\n", err) }

		if Debug {
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
