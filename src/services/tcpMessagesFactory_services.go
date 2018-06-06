package services

import (
	"bytes"
	"encoding/binary"
	"models"
)

func CreateOnionTunnelBuild(onionTunnelBuild *models.OnionTunnelBuild) ([]byte)  {
	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, onionTunnelBuild.OnionTunnelBuild)
	message := messageTypeBuf.Bytes()

	/*****
		Reserved and networkVersion
	 	Convert networkVersion to Byte Array
		Set to 0 if IPv4. Set to 1 if IPv6
	*****/
	networkVersionBuf := new(bytes.Buffer)
	if onionTunnelBuild.NetworkVersion == "IPv4"{
		binary.Write(networkVersionBuf, binary.BigEndian, uint16(0))
	} else if onionTunnelBuild.NetworkVersion == "IPv6"{
		binary.Write(networkVersionBuf, binary.BigEndian, uint16(1))
	}
	message = append(message, networkVersionBuf.Bytes()...)

	// Convert port to Byte Array
	portBuf := new(bytes.Buffer)
	binary.Write(portBuf, binary.BigEndian, onionTunnelBuild.Port)
	message = append(message, portBuf.Bytes()...)

	// Convert destinationAddress to Byte Array
	message = append(message, []byte(onionTunnelBuild.DestinationAddress)...)

	// Convert destinationHostkey to Byte Array
	message = append(message, []byte(onionTunnelBuild.DestinationHostkey)...)


	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateOnionTunnelReady(onionTunnelReady *models.OnionTunnelReady) ([]byte) {
	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, onionTunnelReady.OnionTunnelReady)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	message = append(message, []byte(onionTunnelReady.TunnelID)...)

	// Convert destinationHostkey to Byte Array
	message = append(message, []byte(onionTunnelReady.DestinationHostkey)...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateOnionTunnelIncoming(onionTunnelIncoming *models.OnionTunnelIncoming) ([]byte){
	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, onionTunnelIncoming.OnionTunnelIncoming)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	message = append(message, []byte(onionTunnelIncoming.TunnelID)...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateOnionTunnelDestroy(onionTunnelDestroy *models.OnionTunnelDestroy) ([]byte){
	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, onionTunnelDestroy.OnionTunnelDestroy)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	message = append(message, []byte(onionTunnelDestroy.TunnelID)...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}