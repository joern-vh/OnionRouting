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
)

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
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":" + strconv.Itoa(FreePortForListening))
	if err != nil {
		log.Fatal("Problem creating listernier, please check your input")
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal("createTcpListener: Problem creating net.TCPListener: " + err.Error())
	}
	// accept new connections on TCP
	go func() {
		log.Println("StartTCPListening: Started listening")
		for {
			conn1, err := listener.Accept()
			if err != nil {
				log.Println("Couldn't accept new TCP Connection, not my problem!")
			} else {
				// Pass each message into the right channel
				log.Println("Conn: ", conn1)
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