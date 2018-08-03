package main

import (
	"services"
	"net"
	"strconv"
	"log"
	"models"
	"os"
	"bufio"
	"crypto/x509"
	"os/signal"
	"bytes"
	"encoding/binary"
)

// Used to transmit the message and the ip in the CommunicationChannelTCPMessages
type TCPMessageChannel struct {
	Message		[]byte
	Host		string		// Attention, hast port
}

type availableHost struct {
	NetworkVersion		string
	DestinationAddress	string
	Port				int
	DestinationHostkeyPath	string
}

var LocalIpAddress = "127.0.0.1"
var FreePortForListening int

/************************************************************************************************************
********** test case one, simply create one tunnel, wait till it's finished and send some UDP data **********
************************************************************************************************************/

// if you want to add new peers, parse the kexpair before so that the peer can be created with the destinationHostKey
var _, pub0, _ = services.ParseKeys("keypair1.pem")
var _, pub1, _ = services.ParseKeys("keypair2.pem")
var _, pub2, _ = services.ParseKeys("keypair3.pem")
var _, pub3, _ = services.ParseKeys("keypair4.pem")

var CommunicationChannelTCPMessages chan TCPMessageChannel

// define list of available host >> If you want to add a new host, simply use the &availableHost Model as shown below >> TODO: Please adapt iPaddress
var AvailableHosts = []*availableHost{
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.2.3", Port:3000, DestinationHostkeyPath: "keypair1.pem"},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.2.3", Port:4200, DestinationHostkeyPath: "keypair2.pem"},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.2.3", Port:4500, DestinationHostkeyPath: "keypair3.pem"},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.2.3", Port:4800, DestinationHostkeyPath: "keypair4.pem"},
}

func main()  {
	writeFile()
	var err error
	FreePortForListening, err = services.GetFreePort()
	// first, create a TCP Listener to receive incoming messages >> for example the OnionTunnelReady, imitating the CM/UI module.
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":" + strconv.Itoa(9999))
	if err != nil {
		log.Fatal("Problem creating listerning, please check your input")
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal("createTcpListener: Problem creating net.TCPListener: " + err.Error())
	}

	CommunicationChannelTCPMessages = make(chan TCPMessageChannel)

	listening()

	// accept new connections on TCP
	go func() {
		log.Println("StartTCPListening: Started listening")
		for {
			conn1, err := listener.Accept()
			if err != nil {
				log.Println("Couldn't accept new TCP Connection, not my problem!")
			} else {
				// Pass each message into the right channel
				go handleMessages(conn1)
			}
		}
	}()

	// second, create and send a onionTunnelBuildMessage
	conn, err := net.Dial("tcp", AvailableHosts[0].DestinationAddress + ":" + strconv.Itoa(AvailableHosts[0].Port))
	if err != nil {
		log.Fatal("Problem dialing to your first host, please verify that everything is set up right")
	}

	var _, pub, _ = services.ParseKeys(AvailableHosts[(len(AvailableHosts) - 1)].DestinationHostkeyPath)

	onionTunnelBuildMessage := models.OnionTunnelBuild{NetworkVersion: AvailableHosts[(len(AvailableHosts) - 1)].NetworkVersion, Port:uint16(AvailableHosts[(len(AvailableHosts) - 1)].Port) , DestinationAddress: AvailableHosts[(len(AvailableHosts) - 1)].DestinationAddress, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub)}
	message := services.CreateOnionTunnelBuild(onionTunnelBuildMessage)

	conn.Write(message)



	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	log.Println("Peer status: offline \n\n\n")
	os.Exit(0)

}


// Passes each message into the right channel
func handleMessages (conn net.Conn) {
	reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(reader)

	defer conn.Close()

	scanner.Split(ScanCRLF)

	for scanner.Scan() {
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

func listening() {
	go func() {
		for msg := range CommunicationChannelTCPMessages {
			log.Println("\n\n")
			log.Println("StartPeerController: New message from " + msg.Host)
			log.Println(msg.Message)
			log.Println("Message Type: ", binary.BigEndian.Uint16(msg.Message[2:4]))
		}
	}()
}

func writeFile() {
	f, err := os.Create("tests/hosts.ini")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	f.WriteString("; Host Testing\n\n")
	f.WriteString("[Hosts]\n")
	f.WriteString("Number = " + strconv.Itoa(len(AvailableHosts)-1) + "\n\n")

	for i := 1; i < len(AvailableHosts); i++ {
		availableHost := AvailableHosts[i]
		f.WriteString("[Host" + strconv.Itoa(i) + "]\n")
		f.WriteString("NetworkVersion = " + availableHost.NetworkVersion + "\n")
		f.WriteString("DestinationAddress = " + availableHost.DestinationAddress + "\n")
		f.WriteString("Port = " + strconv.Itoa(availableHost.Port) + "\n")
		f.WriteString("DestinationHostkey = " + availableHost.DestinationHostkeyPath + "\n")
	}

	f.Sync()

	w := bufio.NewWriter(f)

	w.Flush()
}