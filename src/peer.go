package main

import (
	"bufio"
	"log"
	"net"
	"errors"
	"os"
)

const destinationAddress = "192.168.0.15:3000"
//const destinationAddress = "127.0.0.1:3000"

// Open returns a new tcp connection inside a bufio read writer
func open() (*bufio.ReadWriter, error) {

	log.Println("Dial " + destinationAddress)
	// First, resolve TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", destinationAddress)
	if err != nil {
		return nil, errors.New("ResolveTCPAddr failed: " + err.Error())
	}

	// Use tcp here for all "structure like" requests, traffic later over UDP
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, errors.New("Dial failed: " + err.Error())
	}

	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

func main() {
	conn, err := open()
	if err != nil {
		log.Println("Failed opening a connection" + destinationAddress)
		log.Println("Error: ", err)
		os.Exit(0)
	}

	n, err := conn.WriteString("It's not a bug, it's a feature !!!!!!! \n")
	if err != nil {
		log.Println("Failed sending to other peer")
		log.Println("Error: ", err)
	}

	err = conn.Flush()
	log.Println("N: ", n)
}