package controllers
// Use this controller to react to external incoming messsages from other modules

import (
	"services"
	"log"
	"encoding/binary"
	"fmt"
	"strconv"
	"models"
	"crypto/x509"
	"container/list"
	"net"
)

type availableHost struct {
	NetworkVersion		string
	DestinationAddress	string
	Port				int
	DestinationHostkey	[]byte
}

var _, pub1, _ = services.ParseKeys("keypair2.pem")
var _, pub2, _ = services.ParseKeys("keypair3.pem")
var _, pub3, _ = services.ParseKeys("keypair4.pem")

// define list of available host
var AvailableHosts = []*availableHost{
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.0.10", Port:4200, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub1)},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.0.10", Port:4500, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub2)},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.0.10", Port:4800, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub3)},
}

func StartPeerController(myPeer *services.Peer) {
	log.Println("StartPeerController: Started Controller")
	go func() {
		for msg := range services.CommunicationChannelTCPMessages {
			log.Println("\n\n")
			log.Println("StartPeerController: New message from " + msg.Host)
			handleTCPMessage(msg, myPeer)
			handlePeerControllerMessage(msg, myPeer)
		}
	}()
}

func StartErrorHandling(myPeer *services.Peer){
	log.Println("Error Handling, started Controller")
	go func() {
		for error := range services.CummunicationChannelError {
			log.Println("\n\n")
			log.Println("StartErrorHandling, new error for Tunnel " , strconv.Itoa(int(error.TunnelId)))
			// Now, handle error
		}
	}()
}

func handlePeerControllerMessage(messageChannel services.TCPMessageChannel, myPeer *services.Peer) error {
	messageType := binary.BigEndian.Uint16(messageChannel.Message[2:4])
	log.Println("PeerController: Messagetype:", messageType)

	switch messageType {
	// ONION TUNNEL BUILD
	case 560:
		handleOnionTunnelBuild(messageChannel, myPeer)
		break

	default:
		return nil
	}

	return nil
}

func handleOnionTunnelBuild(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Print("Messagetype: Onion tunnel build")
	var networkVersionString string
	var destinationAddress string
	var destinationHostkey []byte

	networkVersion := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	onionPort := binary.BigEndian.Uint16(messageChannel.Message[6:8])

	if networkVersion == 0 {
		networkVersionString = "IPv4"
		destinationAddress = net.IP(messageChannel.Message[8:12]).String()
		destinationHostkey = messageChannel.Message[12:]
	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
		destinationAddress = net.IP(messageChannel.Message[8:24]).String()
		destinationHostkey = messageChannel.Message[24:]
	}

	// Generate new tunnel id
	newTunnelID := services.CreateTunnelID()
	log.Println("NewTunnelID: ", newTunnelID)
	log.Println("IP-Address of destination: ", destinationAddress)
	log.Print("Onion Port of destination: ", onionPort)


	// Initialize list for TunnelHostOrder with new TunnelID as Key for the Hashmap
	myPeer.PeerObject.TunnelHostOrder[newTunnelID] = new(list.List)
	// then, generate the hashed version of the destinationHostKey and add it as first value to the list
	myPeer.PeerObject.TunnelHostOrder[newTunnelID].PushBack(services.GenerateIdentityOfKey(myPeer.PeerObject.PublicKey))
	// now, add the new TCP Connection to the peer under TCPConnections, indentified by the newTunnelid
	myPeer.PeerObject.TCPConnections[newTunnelID] = &models.TCPConnection{newTunnelID, nil, nil, &models.OnionTunnelBuild{DestinationHostkey: destinationHostkey, Port: onionPort, DestinationAddress: destinationAddress, NetworkVersion: networkVersionString}, x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey)}

	// now, initiate Tunnel Construction
	initiateTunnelConstruction(newTunnelID, myPeer, 4)
}

func initiateTunnelConstruction(tunnelId uint32, mypeer *services.Peer, minAmountHups int){
	// This function is used to keep track of the tunnel state

	//first, generate TCP Writer for the first hop and assign it to the conneciton
	newTCPWriter, err := mypeer.CreateTCPWriter(AvailableHosts[0].DestinationAddress, AvailableHosts[0].Port)
	if err != nil {
		log.Println("Error creating tcp writer, error: " + err.Error())
	}
	mypeer.PeerObject.TCPConnections[tunnelId].RightWriter = newTCPWriter

	// Then, generate the first construct tunnel
	// hash dynamicly loaded key to generate new identifier
	destinationPubKey, err := x509.ParsePKCS1PublicKey(AvailableHosts[0].DestinationHostkey)
	if err != nil {
		log.Println("initiateTunnelConstruction: Can't parse destinationHostKey ", err.Error())
	}
	hashedVersion := services.GenerateIdentityOfKey(destinationPubKey)
	destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)

	// Only for First Hop, generate identifier
	newIdentifier := strconv.Itoa(int(tunnelId)) + destinationHostkeyString

	// Generate Pre Master Key for session.
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	mypeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId: tunnelId, PublicKey: publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}
	encryptedPubKey, err := services.EncryptKeyExchange(destinationPubKey, publicKey)

	constructTunnelMessage := models.ConstructTunnel{AvailableHosts[0].NetworkVersion, uint16(mypeer.PeerObject.UDPPort), uint16(mypeer.PeerObject.P2P_Port), uint32(tunnelId), AvailableHosts[0].DestinationAddress, x509.MarshalPKCS1PublicKey(mypeer.PeerObject.PublicKey), x509.MarshalPKCS1PublicKey(mypeer.PeerObject.PublicKey), encryptedPubKey}
	// Build Construct Tunnel Message
	message := services.CreateConstructTunnelMessage(constructTunnelMessage)

	// Then, send the message
	// at last, send the constructTunnelMessage
	mypeer.PeerObject.TCPConnections[tunnelId].RightWriter.TCPWriter.Write(message)

	// Then, start listening
	// TODO: Write function to keep track of confirmations >>> build evntloop
	log.Println("Start listening for Confirm messages")
	go func() {
		for msg := range services.CommunicationChannelTCPConfirm {	// TODO: Check for tunnelID!
			log.Println("\n\n")
			log.Println("ConfirmListener: New confirm for tunnel " + strconv.Itoa(int(msg.TunnelId)))

			// first, check if the tunnelID matchs to the id which initialized the function
			if msg.TunnelId == tunnelId {
				// now, check if the length of TunnelHostOrder[tunnelId] < minAmountHups >> if so, start a new tunnel construction
				if mypeer.PeerObject.TunnelHostOrder[tunnelId].Len() < minAmountHups {
					connectToNextHop(AvailableHosts[mypeer.PeerObject.TunnelHostOrder[tunnelId].Len() - 1], tunnelId, mypeer)
				} else {
					log.Println("We've enough hops!!!!")
					for e := mypeer.PeerObject.TunnelHostOrder[tunnelId].Front(); e != nil; e = e.Next() {
						fmt.Println(e.Value) // print out the elements
					}
					// TODO: connect to final and if else to check if final yes or no
				}
			}
			// no else needed, another instance of the function handles that
		}
	}()
}

func connectToNextHop(nextHop *availableHost, tunnelId uint32, myPeer *services.Peer) {
	hostkeyPub, err := x509.ParsePKCS1PublicKey(nextHop.DestinationHostkey)
	if err != nil {
		log.Println("Error parsing hostkeyPUB " , err.Error())
	}
	hashedVersion := services.GenerateIdentityOfKey(hostkeyPub)
	destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)

	// Only for First Hop
	newIdentifier := strconv.Itoa(int(tunnelId)) + destinationHostkeyString

	// Generate Pre Master Key for session.
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId: tunnelId, PublicKey: publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}

	// Encrypt Public Key with destination key
	log.Println("PEER: ", nextHop.DestinationHostkey)
	encryptedPubKey, err := services.EncryptKeyExchange(hostkeyPub, publicKey)
	if err != nil {
		log.Println("Problem encrypting pubblic key")
	}

	log.Println("TESTING: SEND TUNNEL INSTRUCTION")
	dataMessage := models.DataConstructTunnel{NetworkVersion: nextHop.NetworkVersion, DestinationAddress: nextHop.DestinationAddress, Port: uint16(nextHop.Port), DestinationHostkey: x509.MarshalPKCS1PublicKey(hostkeyPub), PublicKey: encryptedPubKey}
	data := services.CreateDataConstructTunnel(dataMessage)

	// Now, just for tests, send a forward to a new peer
	tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelId, Data: data}
	message := services.CreateTunnelInstruction(tunnelInstructionMessage)

	myPeer.PeerObject.TCPConnections[tunnelId].RightWriter.TCPWriter.Write(message)
	log.Println("Sent Tunnel Instruction to " + myPeer.PeerObject.TCPConnections[tunnelId].RightWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelId].RightWriter.DestinationPort))
}
