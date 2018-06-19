package services

import (
	"log"
	"net"
	"errors"
	"bufio"
	"strconv"
	"models"
)

// Just used to wrap the actual Peer from models.Peer here to use it as caller
type Peer struct {
	PeerObject *models.Peer
}

// GLobal channel for communication errors from TCP
var CommunicationChannelTCPError chan error
var CommunicationChannelTCPMessages chan []byte


// createPeer crates a new Peer object that is just listening, now writing necessary at the moment
func CreateNewPeer(config *models.Config) (*Peer, error) {
	log.Println("CreatePeer: Start creating a new peer on port", config.P2P_Port)

	// Create new TCPListener for peer. Needs to be done like that due to 2=1 missmatch of arguments
	newTCPListener, err := createTCPListener(config.P2P_Port)
	if err != nil {
		log.Println("CreatePeer: Problem creating TCP listener, error: ", err)
		return &Peer{&models.Peer{nil ,0, "", nil, nil, nil}}, err
	}

	// Create new peer
	newPeer := &Peer{&models.Peer{newTCPListener, config.P2P_Port, config.P2P_Hostname, config.PrivateKey, config.PublicKey, make(map[string] *models.UDPConnection)}}

	return newPeer, nil
}

// createTCPListener creates the *net.Listener for one peer for TCP
func createTCPListener(port int)  (*net.TCPListener, error){
	log.Println("createTCPListener: Create a new listener for TCP")

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

// StartTCPListening lets the peer listen for new TCP-messages on its P2P_Port
func (peer *Peer) StartTCPListening() {
	// Initialize global communicationChannelTCP
	CommunicationChannelTCPError = make(chan error)
	CommunicationChannelTCPMessages = make(chan []byte)

	go func() {
		log.Println("StartTCPListening: Started listenting")
		for {
			conn, err := peer.PeerObject.TCPListener.Accept()
			if err != nil {
			//return errors.New("StartTCPListening: Couldn't start accepting new connctions: " + err.Error())
			continue
			}
			log.Println("StartTCPListening: New message")

			// Reader.Read() saves content received in newMessage
			newMessage := make([]byte, 64)
			amountByteRead, err := bufio.NewReader(conn).Read(newMessage)
			if err != nil {
				CommunicationChannelTCPError <- err
			}
			// output message received
			log.Println("StartTCPListening: Message Received with ", amountByteRead , " bytes: ")
			log.Printf("%x\n", newMessage)

			CommunicationChannelTCPMessages <- newMessage
			//controllers.HandleTCPMessage(newMessage, peer)
		}
	}()
}

// SendMessage gets address, port and message(type byte) to send it to one peer
func (peer *Peer) SendMessage(destinationAddress string, destinationPort int, message []byte) (error) {
	conn, err := net.Dial("tcp", destinationAddress + ":" + strconv.Itoa(destinationPort))
	if err != nil {
		return errors.New("SendMessage: Error while dialing to destination, error: " + err.Error())
	}
	defer conn.Close()

	m, err := conn.Write(message)
	if err != nil {
		return errors.New("SendMessage: Error while writing message to destination, error: " + err.Error())
	}

	log.Printf("SendMessage: Send message of size: %d\n", m)
	return nil
}

func (myPeer *Peer) AppendUDPConnection(connection *UDPConnection) {
	// TODO: Check if imports are right >> Not sure about mdoels, choosen the right one in actual connection model?
	myPeer.PeerObject.UDPConnections[connection.UCPConnObject.TunnelId] = connection.UCPConnObject //= append(myPeer.PeerObject.UDPConnections, connection)
	return
}