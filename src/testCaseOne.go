package main

import (
	"services"
	"crypto/x509"
	"net"
	"strconv"
	"log"
	"models"
	"errors"
)

type availableHost struct {
	NetworkVersion		string
	DestinationAddress	string
	Port				int
	DestinationHostkey	[]byte
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
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.0.10", Port:3000, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub0)},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.0.10", Port:4200, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub1)},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.0.10", Port:4500, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub2)},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.0.10", Port:4800, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub3)},
}

func main()  {
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

	// second, create and send a onionTunnelBuildMessage
	conn, err := net.Dial("tcp", AvailableHosts[0].DestinationAddress + strconv.Itoa(AvailableHosts[0].Port))
	if err != nil {
		log.Fatal("Problem dialing to your first host, please verify that everything is set up right")
	}

	onionTunnelBuildMessage := models.OnionTunnelBuild{NetworkVersion: AvailableHosts[(len(AvailableHosts) - 1)].NetworkVersion, Port:uint16(AvailableHosts[(len(AvailableHosts) - 1)].Port) , DestinationAddress: AvailableHosts[(len(AvailableHosts) - 1)].DestinationAddress, DestinationHostkey: AvailableHosts[(len(AvailableHosts) - 1)].DestinationHostkey}
	message := services.CreateOnionTunnelBuild(onionTunnelBuildMessage)

	conn.Write(message)

}