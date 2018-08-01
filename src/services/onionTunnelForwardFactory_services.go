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

	// Convert port to Byte Array
	portBuf := new(bytes.Buffer)
	binary.Write(portBuf, binary.BigEndian, dataConstructTunnel.Port)
	message = append(message, portBuf.Bytes()...)

	// Append size of Destination Hostkey
	destinationHostkeyLengthBuf := new(bytes.Buffer)
	binary.Write(destinationHostkeyLengthBuf, binary.BigEndian, uint16(len(dataConstructTunnel.DestinationHostkey)))
	message = append(message, destinationHostkeyLengthBuf.Bytes()...)

	// Append destinationHostkey to Message Array
	message = append(message, dataConstructTunnel.DestinationHostkey...)

	// Append Public Key
	message = append(message, dataConstructTunnel.PublicKey...)

	return message
}

func CreateDataConfirmTunnelConstruction(dataConfirmTunnelConstruction models.DataConfirmTunnelConstruction) ([]byte) {
	// Message Type
	command := uint16(568)

	// Convert command to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, command)
	message := messageTypeBuf.Bytes()

	// Append size of Destination Hostkey
	destinationHostkeyLengthBuf := new(bytes.Buffer)
	binary.Write(destinationHostkeyLengthBuf, binary.BigEndian, uint16(len(dataConfirmTunnelConstruction.DestinationHostkey)))
	message = append(message, destinationHostkeyLengthBuf.Bytes()...)

	// Append destinationHostkey to Message Array
	message = append(message, dataConfirmTunnelConstruction.DestinationHostkey...)

	// Append Public Key
	message = append(message, dataConfirmTunnelConstruction.PublicKey...)

	return message
}

func CreateDataExchangeKey(exchangeKey models.ExchangeKey) ([]byte) {
	// Message Type
	command := uint16(571)

	// Convert command to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, command)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, exchangeKey.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append size of Destination Hostkey
	destinationHostkeyLengthBuf := new(bytes.Buffer)
	binary.Write(destinationHostkeyLengthBuf, binary.BigEndian, uint16(len(exchangeKey.DestinationHostkey)))
	message = append(message, destinationHostkeyLengthBuf.Bytes()...)

	// Append Destination Hostkey
	message = append(message, exchangeKey.DestinationHostkey...)

	// Append Public Key
	message = append(message, exchangeKey.PublicKey...)

	return message
}