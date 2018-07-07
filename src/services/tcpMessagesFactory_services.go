package services

import (
	"bytes"
	"encoding/binary"
	"models"
	"net"
	"log"
)

func CreateOnionTunnelBuild(onionTunnelBuild models.OnionTunnelBuild) ([]byte)  {
	// Message Type
	messageType := uint16(560)

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
	ip := net.ParseIP(onionTunnelBuild.DestinationAddress)
	if onionTunnelBuild.NetworkVersion == "IPv4"{
		binary.Write(networkVersionBuf, binary.BigEndian, uint16(0))
		ip = ip.To4()
	} else if onionTunnelBuild.NetworkVersion == "IPv6"{
		binary.Write(networkVersionBuf, binary.BigEndian, uint16(1))
		ip.To16()
	}
	message = append(message, networkVersionBuf.Bytes()...)

	// Convert port to Byte Array
	portBuf := new(bytes.Buffer)
	binary.Write(portBuf, binary.BigEndian, onionTunnelBuild.Port)
	message = append(message, portBuf.Bytes()...)

	// Convert destinationAddress to Byte Array
	//ip := net.ParseIP(onionTunnelBuild.DestinationAddress)
	log.Printf("IP: %x\n", []byte(ip))
	message = append(message, ip...)

	// Convert destinationHostkey to Byte Array
	message = append(message, onionTunnelBuild.DestinationHostkey...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateOnionTunnelReady(onionTunnelReady models.OnionTunnelReady) ([]byte) {
	// Message Type
	messageType := uint16(561)

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, onionTunnelReady.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Convert destinationHostkey to Byte Array
	message = append(message, onionTunnelReady.DestinationHostkey...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateOnionTunnelIncoming(onionTunnelIncoming models.OnionTunnelIncoming) ([]byte){
	// Message Type
	messageType := uint16(562)

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, onionTunnelIncoming.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateOnionTunnelDestroy(onionTunnelDestroy models.OnionTunnelDestroy) ([]byte){
	// Message Type
	messageType := uint16(563)

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, onionTunnelDestroy.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateOnionTunnelError(onionError models.OnionError) ([]byte) {
	// Message Type
	messageType := uint16(565)

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)
	message := messageTypeBuf.Bytes()

	// Convert Request Type to Byte Array
	requestTypeBuf := new(bytes.Buffer)
	binary.Write(requestTypeBuf, binary.BigEndian, onionError.RequestType)
	message =  append(message, messageTypeBuf.Bytes()...)

	// Convert Reserved to Byte Array (16 bits of 0)
	reservedBuf := new(bytes.Buffer)
	binary.Write(reservedBuf, binary.BigEndian, uint16(0))
	message =  append(message, reservedBuf.Bytes()...)

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, onionError.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}