package main

import (
	"net"
	"bufio"
	"fmt"
	"log"
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

func main() {
	listen("3000")
}