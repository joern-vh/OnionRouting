package services

import (
	"models"
	"net"
	"strconv"
	"errors"
	"log"
)

// Just used to wrap the actual Peer from models.Peer here to use it as caller
type UDPConnection struct {
	UCPConnObject *models.UDPConnection
}

var CommunicationChannelUDP chan error


// createNewUDPConnection creates a new udp connection for a peer
func CreateNewUDPConnection(port int, networkVersion string, destinationAddress string, destinationHostKey []byte) (*UDPConnection, error) {
	newUDPConn, err := createUDPConn(port)
	if err != nil {
		return nil, errors.New("CreateNewUDPConnection: Error creating conn" + err.Error())
	}

	// TODO: Generate tunnel id

	return &UDPConnection{ &models.UDPConnection{port, newUDPConn, "test", networkVersion, destinationAddress, destinationHostKey}}, nil
}

// createUDPConn creates the *net.Conn for one peer for UDP
func createUDPConn(port int)  (*net.UDPConn, error){
	log.Println("createUDPConn: Create a new conn for UDP")

	udpAddr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(port))
	if err != nil {
		return nil, errors.New("createUDPConn: Problem resolving UDP Address: " + err.Error())
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, errors.New("createUDPConn: Problem creating net.UDPConn: " + err.Error())
	}

	return conn, nil
}

// StartUDPListening lets the peer listen for new UDP-messages
func (udpConn *UDPConnection) StartUDPListening() error {
	log.Println("StartUDPListening: Started listenting")
	buf := make([]byte, 1024)
	for {
		n,addr,err := udpConn.UCPConnObject.UDPConn.ReadFromUDP(buf)
		if err != nil {
			return errors.New("StartUDPListening: Couldn't read message received: " + err.Error())
		}
		log.Println("StartUDPListening: Message Received ", string(buf[0:n]), " from ",addr)

	}
}