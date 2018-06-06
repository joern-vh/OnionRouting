package main

import (
	"net"
	"bufio"
	"fmt"
	"log"
	"strconv"
	"models"
	"services"
)

func listen(port string) {
	fmt.Printf("Started listening on port %s:\n\n", port)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":" + port)
	if err != nil {
		log.Println(err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Println(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {

		}
		fmt.Printf("New Message: %s\n", message)
	}
}

func sendMessage(destinationAddress string, destinationPort int, message []byte) {
	conn, err := net.Dial("tcp", destinationAddress + ":" + strconv.Itoa(destinationPort))
	if err != nil {
		fmt.Printf("Error while sending Message\n")
	}

	m, err := conn.Write(message)

	if err != nil {
		log.Println("Error while writing: ", err)
	}

	fmt.Printf("Message Size: %d\n", m)

	conn.Close()
}

func main() {
	//listen("3000")
	buildMessage := models.OnionTunnelBuild{OnionTunnelBuild: uint16(560), NetworkVersion: "IPv4", Port: uint16(4200), DestinationAddress: "", DestinationHostkey: "KEY"}
	onionTunnelBuild := services.CreateOnionTunnelBuild(buildMessage)
	fmt.Printf("Message: %x\n", onionTunnelBuild)
	sendMessage("192.168.0.10", 3000, onionTunnelBuild)
}