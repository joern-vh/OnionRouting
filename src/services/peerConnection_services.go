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

// GLobal channel for communication errors and messages from TCP
var CommunicationChannelTCPErrors chan error
var CommunicationChannelTCPMessages chan []byte

// GLobal channel for communication errors and messages from TCP
var CommunicationChannelUDPErrors chan error
var CommunicationChannelUDPMessages chan []byte

// CreatePeer crates a new Peer object that is just listening, now writing necessary at the moment
func CreateNewPeer(config *models.Config) (*Peer, error) {
	log.Println("CreatePeer: Start creating a new peer on port", config.P2P_Port)

	// Create new TCPListener for peer. Needs to be done like that due to 2=1 missmatch of arguments
	newTCPListener, err := createTCPListener(config.P2P_Port)
	if err != nil {
		log.Println("CreatePeer: Problem creating TCP listener, error: ", err)
		return &Peer{&models.Peer{nil , nil, 0, "", nil, nil, nil}}, err
	}

	// Create new UDPConn to listen for udp messages
	newUDPListener, err := createUDPConn()
	if err != nil {
		log.Println("CreatePeer: Problem creating UDP listener, error: ", err)
		return &Peer{&models.Peer{nil , nil, 0, "", nil, nil, nil}}, err
	}

	// Create new peer
	newPeer := &Peer{&models.Peer{newTCPListener, newUDPListener,  config.P2P_Port, config.P2P_Hostname, config.PrivateKey, config.PublicKey, make(map[string] *models.UDPConnection)}}

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
	CommunicationChannelTCPErrors = make(chan error)
	CommunicationChannelTCPMessages = make(chan []byte)

	go func() {
		log.Println("StartTCPListening: Started listening")
		for {
			conn, err := peer.PeerObject.TCPListener.Accept()
			if err != nil {
				//return errors.New("StartTCPListening: Couldn't start accepting new connctions: " + err.Error())
				continue
			}

			// Reader.Read() saves content received in newMessage
			newMessage := make([]byte, 64)
			amountBytesRead, err := bufio.NewReader(conn).Read(newMessage)
			if err != nil {
				CommunicationChannelTCPErrors <- err
			}
			log.Println("StartTCPListening: Received new message with ", amountBytesRead, " bytes")
			// Pass newMessage into TCPMessageChannel
			CommunicationChannelTCPMessages <- newMessage
		}
	}()
}

// createUDPConn creates the *net.Conn for one peer to listen to UDP messages
func createUDPConn()  (*net.UDPConn, error){
	log.Println("createUDPConn: Create a new listener for UDP")
	// First, create new port
	port, err := getFreePort()
	if err != nil {
		return nil, errors.New("createUDPConn: Couldn't create new port, " + err.Error())
	}

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
func (peer *Peer) StartUDPListening() {

	// Initialize global communicationChannelUDP
	CommunicationChannelUDPErrors = make(chan error)
	CommunicationChannelUDPMessages = make(chan []byte)

	CreateTunnelID()

	go func() {
		log.Println("StartUDPListening: Started listening")
		buf := make([]byte, 1024)
		for {
			n,addr,err := peer.PeerObject.UDPListener.ReadFromUDP(buf)
			if err != nil {
				CommunicationChannelUDPErrors <- err
			}
			log.Println("StartUDPListening: Message Received ", string(buf[0:n]), " from ",addr)

			CommunicationChannelUDPMessages <- buf[0:n]
		}
	}()
}

// getFreePort returns a new free port
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
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
