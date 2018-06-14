package services

import (
	"log"
	"net"
	"errors"
	"bufio"
	"strconv"
	"fmt"
	"models"
	"controllers"
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
		return &Peer{&models.Peer{nil, nil ,0, "", nil, nil}}, err
	}

	// Create new UDPListener for peer
	newUDPListener, err := createUDPConn(config.P2P_Port)
	if err != nil {
		log.Println("CreatePeer: Problem creating UDP conn, error: ", err)
		return &Peer{&models.Peer{nil, nil ,0, "", nil, nil}}, err
	}

	// Create new peer
	newPeer := &Peer{&models.Peer{newTCPListener, newUDPListener, config.P2P_Port, config.P2P_Hostname, config.PrivateKey, config.PublicKey}}

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

// StartTCPListening lets the peer listen for new TCP-messages on its P2P_Port
func (peer *Peer) StartTCPListening() error {
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
			return errors.New("StartTCPListening: Couldn't read message received: " + err.Error())
		}
		// output message received
		log.Println("StartTCPListening: Message Received with ", amountByteRead , " bytes: ")
		fmt.Printf("%x\n", newMessage)
		controllers.HandleTCPMessage(newMessage)
	}
}

// StartUDPListening lets the peer listen for new UDP-messages
func (peer *Peer) StartUDPListening() error {
	log.Println("StartUDPListening: Started listenting")
	buf := make([]byte, 1024)
	for {
		n,addr,err := peer.PeerObject.UDPConn.ReadFromUDP(buf)
		if err != nil {
			return errors.New("StartUDPListening: Couldn't read message received: " + err.Error())
		}
		log.Println("StartUDPListening: Message Received ", string(buf[0:n]), " from ",addr)

	}
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