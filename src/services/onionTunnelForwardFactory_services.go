package services

import (
	"models"
	"bytes"
	"encoding/binary"
	"net"
)

func CreateDataConstructTunnel(dataConstructTunnel models.DataConstructTunnel) ([]byte) {
	// Message Type
	command := uint16(567)

	// Convert command to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, command)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	//newID := CreateTunnelID()
	binary.Write(tunnelIDBuf, binary.BigEndian, dataConstructTunnel.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	/*****
		Reserved and networkVersion
	 	Convert networkVersion to Byte Array
		Set to 0 if IPv4. Set to 1 if IPv6
	*****/
	networkVersionBuf := new(bytes.Buffer)
	ip := net.ParseIP(dataConstructTunnel.DestinationAddress)
	if dataConstructTunnel.NetworkVersion == "IPv4"{
		binary.Write(networkVersionBuf, binary.BigEndian, uint16(0))
		ip = ip.To4()
	} else if dataConstructTunnel.NetworkVersion == "IPv6"{
		binary.Write(networkVersionBuf, binary.BigEndian, uint16(1))
		ip.To16()
	}
	message = append(message, networkVersionBuf.Bytes()...)

	// Convert destinationAddress to Byte Array
	message = append(message, ip...)

	// Append destinationHostkey to Message Array
	message = append(message, dataConstructTunnel.DestinationHostkey...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}