package services

import (
	"bytes"
	"encoding/binary"
	"net"
	"log"
	"models"
)

/*
	Build-2:
		- UDPPort
		- TunnelID (wird selbst erstellt)
		- Destination Address
		- Destination Hostkey

	Ready-2:
		- Port
		- TunnelID
		- (Hostkey)


	destinationListener
	destinationWriter
	returnListener
	returnWriter

	Peer checkt ob für die TunnelID noch ein returnWriter gesetzt ist. Falls ja => weiterleiten, ansonsten behalten.

*/

/*
	Function to create Construct Tunnel Messages. Type: 567.
 */
func CreateConstructTunnelMessage(constructTunnel models.ConstructTunnel) []byte {
	// Message Type
	messageType := uint16(567)

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)
	message := messageTypeBuf.Bytes()

	/*****
		Reserved and networkVersion
	 	Convert networkVersion to Byte Array
		Set to 0 if IPv4. Set to 1 if IPv6
	*****/
	networkVersionBuf := new(bytes.Buffer)
	ip := net.ParseIP(constructTunnel.DestinationAddress)
	if constructTunnel.NetworkVersion == "IPv4"{
		binary.Write(networkVersionBuf, binary.BigEndian, uint16(0))
		ip = ip.To4()
	} else if constructTunnel.NetworkVersion == "IPv6"{
		binary.Write(networkVersionBuf, binary.BigEndian, uint16(1))
		ip.To16()
	}
	message = append(message, networkVersionBuf.Bytes()...)

	// Convert port to Byte Array
	portBuf := new(bytes.Buffer)
	binary.Write(portBuf, binary.BigEndian, constructTunnel.Port)
	message = append(message, portBuf.Bytes()...)

	// Convert destinationAddress to Byte Array
	log.Printf("IP: %x\n", []byte(ip))
	message = append(message, ip...)

	// Append destinationHostkey to Message Array
	message = append(message, constructTunnel.DestinationHostkey...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction models.ConfirmTunnelConstruction) []byte {
	// Message Type
	messageType := uint16(567)

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	message = append(message, []byte(confirmTunnelConstruction.TunnelID)...)

	// Append destinationHostkey to Message Array
	message = append(message, confirmTunnelConstruction.DestinationHostkey...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

// ToDo: Error Messages.