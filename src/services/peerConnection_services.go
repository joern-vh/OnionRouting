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

// Used to transmit the message and the ip in the CommunicationChannelTCPMessages
type TCPMessageChannel struct {
	Message		[]byte
	Host		string		// Attention, hast port
}

// GLobal channel for communication errors and messages from TCP
var CommunicationChannelTCPErrors chan error
var CommunicationChannelTCPMessages chan TCPMessageChannel

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
	newUDPListener, err := createUDPListener()
	if err != nil {
		log.Println("CreatePeer: Problem creating UDP listener, error: ", err)
		return &Peer{&models.Peer{nil , nil, 0, "", nil, nil, nil}}, err
	}

	// Create new peer
	newPeer := &Peer{&models.Peer{newTCPListener, newUDPListener,  config.P2P_Port, config.P2P_Hostname, config.PrivateKey, config.PublicKey, make(map[uint32] *models.UDPConnection)}}

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
	CommunicationChannelTCPMessages = make(chan TCPMessageChannel)

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
			log.Println("StartTCPListening: Received new message with ", amountBytesRead, " bytes from ", conn.RemoteAddr().String())
			// Pass newMessage into TCPMessageChannel
			CommunicationChannelTCPMessages <- TCPMessageChannel{newMessage, conn.RemoteAddr().String()}
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

func createUDPListener() (*net.UDPConn, error) {
	log.Println("createUDPListener: Create a new listener for UDP")

	// First, create new port
	port, err := getFreePort()
	if err != nil {
		return nil, errors.New("createUDPConn: Couldn't create new port, " + err.Error())
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", ":" + strconv.Itoa(port))
	if err != nil {
		return nil, errors.New("createUDPListener: Problem resolving UDP Address: " + err.Error())
	}

	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, errors.New("createTcpListener: Problem creating net.TCPListener: " + err.Error())
	}

	return listener, nil
}

// StartUDPListening lets the peer listen for new UDP-messages
func (peer *Peer) StartUDPListening() {

	// Initialize global communicationChannelUDP
	CommunicationChannelUDPErrors = make(chan error)
	CommunicationChannelUDPMessages = make(chan []byte)


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

			buf2 := make([]byte, 1024)
			copy(buf2, buf)
		}
	}()

	log.Println("StartUDPListening: Finished listening")
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

func (peer *Peer) AppendNewUDPConnection(myUDPConnectio  *models.UDPConnection) {
	peer.PeerObject.UDPConnections[myUDPConnectio.TunnelId] = myUDPConnectio
}