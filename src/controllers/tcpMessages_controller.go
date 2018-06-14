package controllers

import (
	"encoding/binary"
	"errors"
	"log"
	"fmt"
	"net"
)

func HandleTCPMessage(message []byte) error {
	//var messageTypeByte []byte = message[3:5]
	messageType := binary.BigEndian.Uint16(message[2:4])
	fmt.Printf("%d\n", messageType)

	switch messageType {
		// ONION TUNNEL BUILD
		case 560:
			networkVersion := binary.BigEndian.Uint16(message[4:6])
			onionPort := binary.BigEndian.Uint16(message[6:8])
			var networkVersionString string
			var destinationAddress string
			var destinationHostkey []byte
			if networkVersion == 0 {
				networkVersionString = "IPv4"
				destinationAddress = net.IP(message[8:12]).String()
				destinationHostkey = message[12:]
			} else if networkVersion == 1 {
				networkVersionString = "IPv6"
				destinationAddress = net.IP(message[8:24]).String()
				destinationHostkey = message[24:]
			}

			// ToDo: Implement functionality.

			fmt.Printf("Network Version: %s\n", networkVersionString)
			fmt.Printf("Onion Port: %d\n", onionPort)
			fmt.Printf("Destination Address: %s\n", destinationAddress)
			fmt.Printf("Destination Hostkey: %s\n", destinationHostkey)
			break

		// ONION TUNNEL READY
		case 561:
			tunnelID := string(message[4:8])
			destinationHostkey := message[8:]
			fmt.Printf("Tunnel ID: %s\n", tunnelID)
			fmt.Printf("Destination Hostkey: %s\n", destinationHostkey)
			break

		// ONION TUNNEL INCOMING
		case 562:
			tunnelID := string(message[4:8])
			fmt.Printf("Tunnel ID: %s\n", tunnelID)
			break

		// ONION TUNNEL DESTROY
		case 563:
			log.Println("ONION TUNNEL DESTROY received")
			tunnelID := string(message[4:8])
			fmt.Printf("Tunnel ID: %s\n", tunnelID)
			break
		default:
			return errors.New("tcpMessagesController: Message Type not Found")
	}

	return nil
}