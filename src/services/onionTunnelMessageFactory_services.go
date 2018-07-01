package services

import (
	"bytes"
	"encoding/binary"
	"net"
	"log"
	"models"
	"time"
)

/*
	Function to create Construct Tunnel Messages. Type: 567.
 */
func CreateConstructTunnelMessage(constructTunnel models.ConstructTunnel) ([]byte) {
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

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	//newID := CreateTunnelID()
	binary.Write(tunnelIDBuf, binary.BigEndian, constructTunnel.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Convert destinationAddress to Byte Array
	log.Printf("IP: %x\n", []byte(ip))
	message = append(message, ip...)

	// Append destinationHostkey to Message Array
	message = append(message, constructTunnel.DestinationHostkey...)

	log.Println("KEY SIZE: ", len(constructTunnel.DestinationHostkey))

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction models.ConfirmTunnelConstruction) ([]byte) {
	// Message Type
	messageType := uint16(568)

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, confirmTunnelConstruction.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append destinationHostkey to Message Array
	message = append(message, confirmTunnelConstruction.DestinationHostkey...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateTunnelInstruction(tunnelInstruction models.TunnelInstruction) ([]byte) {
	// Message Type
	messageType := uint16(569)

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Arrays
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, tunnelInstruction.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// ToDo: Encrypt Data.
	message = append(message, tunnelInstruction.Data...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

// ToDo: Error Messages.


// TunnelID: 32 bits.
func CreateTunnelID() (uint32){
	currentTime := time.Now().UnixNano()
	currentTimeBuf := new(bytes.Buffer)

	binary.Write(currentTimeBuf, binary.BigEndian, currentTime)
	id := currentTimeBuf.Bytes()[4:8]

	return binary.BigEndian.Uint32(id)
}