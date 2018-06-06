package services

import (
	"log"
	"net"
	"errors"
	"bufio"
	"strconv"
	"fmt"
	"models"
)

// Just used to wrap the actual Peer from models.Peer here to use it as caller
type Peer struct {
	PeerObject *models.Peer
}
// createPeer crates a new Peer object that is just listening, now writing necessary at the moment
func CreateNewPeer(config *models.Config) (*Peer, error) {
	log.Println("CreatePeer: Start creating a new peer on port", config.P2P_Port)

	// Create new TCPListener for peer. Needs to be done like that due to 2=1 missmatch of arguments
	newTCPListener, err := createTCPListener(config.P2P_Port)
	if err != nil {
		log.Println("CreatePeer: Problem creating TCP listener, error: ", err)
		return &Peer{&models.Peer{nil, 0, ""}}, err
	}

	// Create new peer
	newPeer := &Peer{&models.Peer{newTCPListener, config.P2P_Port, config.P2P_Hostname}}

	return newPeer, nil
}

// createTCPListener creates the *net.Listener for one peer for TCP
func createTCPListener(port int)  (*net.TCPListener, error){
	log.Println("createListener: Create a new listener for TCP")

	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":" + strconv.Itoa(port))
	if err != nil {
		return nil, errors.New("createTcpListener: Problem resolving TCP Address: " + err.Error())
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, errors.New("createTcpListener: Problem creating net.TCPListener: " + err.Error())
	}

	return listener, nil
}

// startTCPListening lets the peer listen for new TCP-messages on its P2P_Port
func (peer *Peer) StartTCPListening() error {
	log.Println("StartTCPListening: Started listenting")
	for {
		conn, err := peer.PeerObject.Listener.Accept()
		if err != nil {
			//return errors.New("StartTCPListening: Couldn't start accepting new connctions: " + err.Error())
			continue
		}
		log.Println("StartTCPListening: New message")

		// Reader.Read() saves content received in newMessage
		newMessage := make([]byte, 64)
		amountByteRead, err := bufio.NewReader(conn).Read(newMessage)
		if err != nil {
			return errors.New("StartTCPListening: Couldn't read message received: " + err.Error())
		}
		// output message received
		log.Println("StartTCPListening: Message Received with ", amountByteRead , " bytes: ")
		fmt.Printf("%x\n", newMessage)
	}
}
