package services

import (
	"bytes"
	"encoding/binary"
	"models"
	"time"
	"crypto/rsa"
	"log"
)

/*
	Function to create Construct Tunnel Messages. Type: 567.
	Encrypted with ephemeral key between the sender and the destination
	ToDo: Encrypt tunnelID and switch place of port
 */
func CreateConstructTunnelMessage(constructTunnel models.ConstructTunnel, sessionKey []byte, destinationPublicKey *rsa.PublicKey) ([]byte) {
	// Message Type
	messageType := uint16(567)
	message := encryptMessageType(messageType, destinationPublicKey)

	// Convert onion port to Byte Array
	portBuf := new(bytes.Buffer)
	binary.Write(portBuf, binary.BigEndian, constructTunnel.OnionPort)
	message = append(message, portBuf.Bytes()...)

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, constructTunnel.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append size of destination hostkey
	destinationHostkeyLengthBuf := new(bytes.Buffer)
	binary.Write(destinationHostkeyLengthBuf, binary.BigEndian, uint16(len(constructTunnel.DestinationHostkey)))
	message = append(message, destinationHostkeyLengthBuf.Bytes()...)

	// Append destinationHostkey to Message Array
	message = append(message, constructTunnel.DestinationHostkey...)

	// Append size of origin hostkey
	originHostkeyLengthBuf := new(bytes.Buffer)
	binary.Write(originHostkeyLengthBuf, binary.BigEndian, uint16(len(constructTunnel.OriginHostkey)))
	//message = append(message, originHostkeyLengthBuf.Bytes()...)
	messageToEncrypt := originHostkeyLengthBuf.Bytes()

	// Append originHostkey to Message Array
	messageToEncrypt = append(messageToEncrypt, constructTunnel.OriginHostkey...)

	// Append public key to Message Array
	messageToEncrypt = append(messageToEncrypt, constructTunnel.PublicKey...)

	EncryptedMessage, err := EncryptData(sessionKey, messageToEncrypt)
	if err != nil {
		log.Println("CreateConstructTunnel: Failed to Encrypt message", err.Error())
	}

	message = append(message, EncryptedMessage...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

/*
	Function to create Construct Tunnel Messages. Type: 567.
	Encrypted with ephemeral key between the sender and the destination
	ToDo: Encrypt tunnelID
 */
func CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction models.ConfirmTunnelConstruction, sessionKey []byte, destinationPublicKey *rsa.PublicKey) ([]byte) {
	// Message Type
	messageType := uint16(568)
	message := encryptMessageType(messageType, destinationPublicKey)

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, confirmTunnelConstruction.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append size of hostkey
	destinationHostkeyLengthBuf := new(bytes.Buffer)
	binary.Write(destinationHostkeyLengthBuf, binary.BigEndian, uint16(len(confirmTunnelConstruction.DestinationHostkey)))
	message = append(message, destinationHostkeyLengthBuf.Bytes()...)

	// Append hostkey of oneself to Message Array
	message = append(message, confirmTunnelConstruction.DestinationHostkey...)

	// Convert port to Byte Array
	portBuf := new(bytes.Buffer)
	binary.Write(portBuf, binary.BigEndian, confirmTunnelConstruction.Port)
	messageToEncrypt := portBuf.Bytes()

	log.Println("UDP Port: ", confirmTunnelConstruction.Port)

	// Append PublicKey to Message Array
	messageToEncrypt = append(messageToEncrypt, confirmTunnelConstruction.PublicKey...)

	// Encrypt Message
	EncryptedMessage, err := EncryptData(sessionKey, messageToEncrypt)
	if err != nil {
		log.Println("Create Confirm Tunnel Instruction: Failed to Encrypt message", err.Error())
	}

	// Append Encrypted Message
	message = append(message, EncryptedMessage...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

func CreateTunnelInstruction(tunnelInstruction models.TunnelInstruction/*, sessionKey []byte*/, destinationPublicKey *rsa.PublicKey) ([]byte) {
	// Message Type
	messageType := uint16(569)
	message := encryptMessageType(messageType, destinationPublicKey)

	// Convert tunnelID to Byte Arrays
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, tunnelInstruction.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append size of hostkey
	destinationHostkeyLengthBuf := new(bytes.Buffer)
	binary.Write(destinationHostkeyLengthBuf, binary.BigEndian, uint16(len(tunnelInstruction.OriginHostkey)))
	message = append(message, destinationHostkeyLengthBuf.Bytes()...)

	// Append Origin Hostkey
	message = append(message, tunnelInstruction.OriginHostkey...)

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

func CreateConfirmTunnelInstruction(confirmTunnelInstruction models.ConfirmTunnelInstruction/*, sessionKey []byte*/, destinationPublicKey *rsa.PublicKey) ([]byte) {
	// Message Type
	messageType := uint16(570)
	message := encryptMessageType(messageType, destinationPublicKey)

	// Convert tunnelID to Byte Arrays
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, confirmTunnelInstruction.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// ToDo: Encrypt Data.
	message = append(message, confirmTunnelInstruction.Data...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}

// Create exchange key message and encrypt tunnelID and public key with RSA.
func CreateExchangeKey(exchangeKey models.ExchangeKey, publicKey *rsa.PublicKey, destinationPublicKey *rsa.PublicKey) ([]byte) {
	// Message Type
	messageType := uint16(571)
	message := encryptMessageType(messageType, destinationPublicKey)

	// Convert tcp port to Byte Array
	tcpPortBuf := new(bytes.Buffer)
	binary.Write(tcpPortBuf, binary.BigEndian, exchangeKey.TCPPort)
	message = append(message, tcpPortBuf.Bytes()...)

	// Convert Status to Byte Array
	statusBuf := new(bytes.Buffer)
	binary.Write(statusBuf, binary.BigEndian, exchangeKey.Status)
	message = append(message, statusBuf.Bytes()...)

	// TUNNEL ID
	// Convert tunnelID and encrypt with RSA to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	binary.Write(tunnelIDBuf, binary.BigEndian, exchangeKey.TunnelID)
	// Encrypt tunnelIDBuf with RSA
	encryptedTunnelID, err := EncryptKeyExchange(publicKey, tunnelIDBuf.Bytes())
	if err != nil {
		log.Println("Create Exchange Key: Failed to encrypt TunnelID")
	}
	// Append size of encrypted tunnel id
	encryptedTunnelIDLengthBuf := new(bytes.Buffer)
	binary.Write(encryptedTunnelIDLengthBuf, binary.BigEndian, uint16(len(encryptedTunnelID)))
	message = append(message, encryptedTunnelIDLengthBuf.Bytes()...)
	// Append encrypted tunnel id
	message = append(message, encryptedTunnelID...)

	// Append size of Destination Hostkey
	destinationHostkeyLengthBuf := new(bytes.Buffer)
	binary.Write(destinationHostkeyLengthBuf, binary.BigEndian, uint16(len(exchangeKey.DestinationHostkey)))
	message = append(message, destinationHostkeyLengthBuf.Bytes()...)

	// Append Destination Hostkey
	message = append(message, exchangeKey.DestinationHostkey...)

	// Append Public Key (with RSA encrypted)
	encryptedPublicKey, err := EncryptKeyExchange(publicKey, exchangeKey.PublicKey)
	message = append(message, encryptedPublicKey...)

	// Append Delimiter
	//message = append(message, []byte("\r\n")...)

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

func encryptMessagePartsWithRSA(data []byte, destinationPublicKey *rsa.PublicKey) ([]byte) {
	// Encrypt data
	encryptedMessageType, err := EncryptKeyExchange(destinationPublicKey, data)
	if err != nil {
		log.Println("ConstructTunnelMessage: Message Type Encryption failed")
	}

	// Append size of encrypted tunnel id
	encryptedMessageTypeBuf := new(bytes.Buffer)
	binary.Write(encryptedMessageTypeBuf, binary.BigEndian, uint16(len(encryptedMessageType)))
	message := encryptedMessageTypeBuf.Bytes()

	// Append Encrypted Message Type
	message = append(message, encryptedMessageType...)

	return message
}

func encryptMessageType(messageType uint16, destinationPublicKey *rsa.PublicKey) ([]byte) {
	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, messageType)


	return encryptMessagePartsWithRSA(messageTypeBuf.Bytes(), destinationPublicKey)
}