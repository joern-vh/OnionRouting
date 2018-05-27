package main

import (
	"log"
	"net"
	"bufio"
	"fmt"
)

func main() {
	log.Println("Started listening")

	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":3000")
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
		log.Println("New message")
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		// output message received
		fmt.Print("Message Received:", string(message))
	}
}
