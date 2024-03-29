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
	"os"
	"math/rand"
	"time"
	"errors"
	"github.com/Thomasdezeeuw/ini"
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

var cmModule string

// define list of available host
var AvailableHosts = []*availableHost{
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.2.3", Port:4200, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub1)},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.2.3", Port:4500, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub2)},
	&availableHost{NetworkVersion:"IPv4", DestinationAddress:"192.168.2.3", Port:4800, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub3)},
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
			log.Println(error.Error.Error())
			handleOnionTunnelDestroy(uint32(error.TunnelId), myPeer)
			os.Exit(1)
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
	cmModule = services.GetIPOutOfAddr(messageChannel.Host)
	readHosts()

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
	log.Println(tunnelId)
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

	// Now, start traffic jamming
	newOnionJammingMessage := models.OnionTunnelTrafficJam{TunnelID:tunnelId, Data: services.GenRandomData()}
	messageJamming := services.CreateOnionTunnelTrafficJamTCP(newOnionJammingMessage)
	log.Println(messageJamming)
	mypeer.PeerObject.TCPConnections[tunnelId].RightWriter.TCPWriter.Write(messageJamming)


	// Then, start listening
	// TODO: Find idea to determine the function
	log.Println("Start listening for Confirm messages")
	go func() {
		for msg := range services.CommunicationChannelTCPConfirm {	// TODO: Check for tunnelID!
			log.Println("\n\n")
			log.Println("ConfirmListener: New confirm for tunnel " + strconv.Itoa(int(msg.TunnelId)))
			log.Println("myTunnelId: " + strconv.Itoa(int(tunnelId)))
			log.Println("receivedTunnelId: " + strconv.Itoa(int(msg.TunnelId)))
			// first, check if the tunnelID matchs to the id which initialized the function
			if msg.TunnelId == tunnelId {
				// now, check if the length of TunnelHostOrder[tunnelId] < minAmountHups >> if so, start a new tunnel construction
				if mypeer.PeerObject.TunnelHostOrder[tunnelId].Len() < minAmountHups {
					// ToDo: RPS Module
						// rpsQuery := services.CreateRPSQuery()
					// Send to RPS Module
					// Wait for response
						// rpsPeer := handleRPSPeer()
						// connectToNextHop(rpsPeer, tunnelId, mypeer)

					connectToNextHop(AvailableHosts[mypeer.PeerObject.TunnelHostOrder[tunnelId].Len() - 1], tunnelId, mypeer)
				} else {
					log.Println("We've enough hops!!!!")

					conn, err := net.Dial("tcp", cmModule + ":" + strconv.Itoa(9999))
					if err != nil {
						log.Println("READY MESSAGE")
					}

					ReadyMessage := models.OnionTunnelReady{DestinationHostkey: mypeer.PeerObject.TCPConnections[tunnelId].FinalDestination.DestinationHostkey, TunnelID: tunnelId}
					message := services.CreateOnionTunnelReady(ReadyMessage)

					n, err1 := conn.Write(message)
					if err1 != nil {
						log.Println(err1.Error())
					}

					log.Println(n)

					for e := mypeer.PeerObject.TunnelHostOrder[tunnelId].Front(); e != nil; e = e.Next() {
						fmt.Println(e.Value) // print out the elements
					}

					return
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

func InitiateTrafficJamming(tunnelId uint32, myPeer *services.Peer)  {
	i := 0
	for true {
		// Now, start traffic jamming
		rand.Seed(time.Now().Unix())
		time.Sleep(time.Duration((rand.Intn(2000-500) + 500)) * time.Millisecond)
		newTrafficJamMessage := models.OnionTunnelTrafficJam{TunnelID:tunnelId, Data:services.GenRandomData()}
		message := services.CreateOnionTunnelTrafficJamTCP(newTrafficJamMessage)
		// like that to prevent null pointer exceptions
		if myPeer.PeerObject.TCPConnections[tunnelId] != nil {
			_, err := myPeer.PeerObject.TCPConnections[tunnelId].RightWriter.TCPWriter.Write(message)
			if err != nil {
				services.CummunicationChannelError <- services.ChannelError{TunnelId: tunnelId, Error: errors.New("InitiateTrafficJamming, Error: " + err.Error())}
				return
			}
			i++
		} else {
			return
		}
	}
}

func readHosts() {
	AvailableHosts = []*availableHost{}
	log.Println("Read Hosts")
	// Open hosts file
	file, err := os.Open("tests/hosts.ini")
	if err != nil {
	}
	defer file.Close()

	// Parse the actual configuration
	hosts, err := ini.Parse(file)
	if err != nil {
	}

	// Need to be done like this to handle erros
	number, err := strconv.Atoi(hosts["Hosts"]["Number"])
	if err != nil {
	}

	for i := 1; i <= number; i++ {
		identifier := "Host" + strconv.Itoa(i)
		log.Println(identifier)
		networkVersion := hosts[identifier]["NetworkVersion"]
		destinationAddress := hosts[identifier]["DestinationAddress"]
		port, _ := strconv.Atoi(hosts[identifier]["Port"])
		destinationHostkeyPath := hosts[identifier]["DestinationHostkey"]
		_, pub, _ := services.ParseKeys(destinationHostkeyPath)

		availableHost := availableHost{NetworkVersion: networkVersion, DestinationAddress: destinationAddress, Port: port, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub)}
		AvailableHosts = append(AvailableHosts, &availableHost)
	}

	log.Println(AvailableHosts)
}