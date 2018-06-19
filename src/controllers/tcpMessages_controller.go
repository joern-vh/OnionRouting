package controllers

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"services"

)

func StartTCPController(myPeer *services.Peer) {
	for msg := range services.CommunicationChannelTCPMessages {
		handleTCPMessage(msg, myPeer)
	}
}

func  handleTCPMessage(message []byte, myPeer *services.Peer) error {
	messageType := binary.BigEndian.Uint16(message[2:4])

	switch messageType {
		// ONION TUNNEL BUILD
		case 560:
			handleOnionTunnelBuild(message)
			break

		// ONION TUNNEL DESTROY
		case 563:
			handleOnionTunnelDestroy(message)
			break

		// CONSTRUCT TUNNEL
		case 567:
			handleConstructTunnel(message)
			break

		// CONFIRM TUNNEL CONSTRUCTION
		case 568:
			handleConfirmTunnelConstruction(message)
			break

		// ToDo: Handle Error Messages while construction is ongoing.

		default:
			return errors.New("tcpMessagesController: Message Type not Found")
	}

	return nil
}

func handleOnionTunnelBuild(message []byte) {
	var networkVersionString string
	var destinationAddress string
	var destinationHostkey []byte

	networkVersion := binary.BigEndian.Uint16(message[4:6])
	onionPort := binary.BigEndian.Uint16(message[6:8])

	if networkVersion == 0 {
		networkVersionString = "IPv4"
		destinationAddress = net.IP(message[8:12]).String()
		destinationHostkey = message[12:]
	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
		destinationAddress = net.IP(message[8:24]).String()
		destinationHostkey = message[24:]
	}

	// ToDo: Functionality.

	log.Printf("Network Version: %s\n", networkVersionString)
	log.Printf("Onion Port: %d\n", onionPort)
	log.Printf("Destination Address: %s\n", destinationAddress)
	log.Printf("Destination Hostkey: %s\n", destinationHostkey)
}

func handleOnionTunnelDestroy(message []byte) {
	log.Println("ONION TUNNEL DESTROY received")
	tunnelID := string(message[4:8])
	log.Printf("Tunnel ID: %s\n", tunnelID)
}

func handleConstructTunnel(message []byte) {
	var networkVersionString string
	var destinationAddress string
	var destinationHostkey []byte

	networkVersion := binary.BigEndian.Uint16(message[4:6])
	onionPort := binary.BigEndian.Uint16(message[6:8])

	tunnelID := binary.BigEndian.Uint16(message[8:12])

	if networkVersion == 0 {
		networkVersionString = "IPv4"
		destinationAddress = net.IP(message[12:16]).String()
		destinationHostkey = message[16:]
	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
		destinationAddress = net.IP(message[12:28]).String()
		destinationHostkey = message[28:]
	}

	// ToDo: Functionality.

	log.Printf("Network Version: %s\n", networkVersionString)
	log.Printf("Onion Port: %d\n", onionPort)
	log.Printf("Tunnel ID: %d\n", tunnelID)
	log.Printf("Destination Address: %s\n", destinationAddress)
	log.Printf("Destination Hostkey: %s\n", destinationHostkey)
}

func handleConfirmTunnelConstruction(message []byte) {
	onionPort := binary.BigEndian.Uint16(message[4:6])
	tunnelID := string(message[6:10])
	destinationHostkey := message[10:]

	log.Printf("Onion Port: %s\n", onionPort)
	log.Printf("Tunnel ID: %s\n", tunnelID)
	log.Printf("Destination Hostkey: %s\n", destinationHostkey)
}
