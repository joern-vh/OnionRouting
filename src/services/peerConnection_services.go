package services

import (
	"log"
	"net"
	"errors"
	"bufio"
	"strconv"
	"models"
	"bytes"
	"container/list"
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

type ConfirmMessageChannel struct {
	TunnelId	uint32
}

// GLobal channel for communication errors and messages from TCP and special confirm messages for the peerController
var CommunicationChannelTCPErrors chan error
var CommunicationChannelTCPMessages chan TCPMessageChannel
var CommunicationChannelTCPConfirm chan ConfirmMessageChannel

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
		return &Peer{&models.Peer{nil , nil, 0, 0, "", nil, nil, nil, nil, nil, nil}}, err
	}

	// Create new UDPConn to listen for udp messages
	newUDPListener, UDPPort, err := createUDPListener()
	if err != nil {
		log.Println("CreatePeer: Problem creating UDP listener, error: ", err)
		return &Peer{&models.Peer{nil , nil, 0, 0, "", nil, nil, nil,  nil, nil, nil}}, err
	}

	// Create new peer
	newPeer := &Peer{&models.Peer{newTCPListener, newUDPListener,  UDPPort,config.P2P_Port, config.P2P_Hostname, config.PrivateKey, config.PublicKey, make(map[uint32] *models.UDPConnection), make(map[uint32] *models.TCPConnection),  make(map[string] *models.CryptoObject), make(map[uint32] *list.List)}}

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
	CommunicationChannelTCPConfirm = make(chan ConfirmMessageChannel)

	go func() {
		log.Println("StartTCPListening: Started listening")
		for {
			conn, err := peer.PeerObject.TCPListener.Accept()
			if err != nil {
				//return errors.New("StartTCPListening: Couldn't start accepting new connctions: " + err.Error())
				log.Fatal(err)
				//continue
			}

			// Pass each message into the right channel
			go handleMessages(conn)
		}
	}()
}

// Passes each message into the right channel
func handleMessages (conn net.Conn) {
	reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(reader)
	defer conn.Close()

		//message, err := reader.ReadBytes('\r', '\n')
		scanner.Split(ScanCRLF)

		for scanner.Scan() {
			/*if err != nil {
				CommunicationChannelTCPErrors <- err
			}*/

			// Pass newMessage into TCPMessageChannel
			CommunicationChannelTCPMessages <- TCPMessageChannel{scanner.Bytes(), conn.RemoteAddr().String()}
		}

}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func ScanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		// We have a full newline-terminated line.
		return i + 2, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// createUDPListener creates a new UDP Listener
func createUDPListener() (*net.UDPConn, int, error) {
	log.Println("createUDPListener: Create a new listener for UDP")

	// First, create new port
	port, err := getFreePort()
	log.Println("createUDPListener: my port: " + strconv.Itoa(port))
	if err != nil {
		return nil, 0, errors.New("createUDPConn: Couldn't create new port, " + err.Error())
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", ":" + strconv.Itoa(port))
	if err != nil {
		return nil, 0, errors.New("createUDPListener: Problem resolving UDP Address: " + err.Error())
	}

	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, 0, errors.New("createTcpListener: Problem creating net.TCPListener: " + err.Error())
	}

	return listener, port, nil
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
			n,_,err := peer.PeerObject.UDPListener.ReadFromUDP(buf)
			if err != nil {
				if err != nil {
					log.Println("StartUDPListening: error " + err.Error())
				}
				CommunicationChannelUDPErrors <- err
			}

			CommunicationChannelUDPMessages <- buf[0:n]

			buf2 := make([]byte, 1024)
			copy(buf2, buf)
		}
	}()
}

func (peer *Peer) CreateTCPWriter (destinationIP string, tcpPort int ) (*models.TCPWriter, error) {

	conn, err := net.Dial("tcp", destinationIP + ":" + strconv.Itoa(tcpPort))
	if err != nil {
		return nil, errors.New("createTCPWriter: Error while dialing to destination, error: " + err.Error())
	}

	return &models.TCPWriter{destinationIP, tcpPort, conn}, nil
}

// Creates a new TCPConnection for the peer with the left writer already set
func (peer *Peer) CreateInitialTCPConnection(tunnelId uint32, finalDestinationHostkey []byte, leftWriter *models.TCPWriter, originHostkey []byte) {
	peer.PeerObject.TCPConnections[tunnelId] = &models.TCPConnection{tunnelId, leftWriter, nil, nil, originHostkey}
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