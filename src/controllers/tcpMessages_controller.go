package controllers

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"services"

	"models"
	"fmt"
	"crypto/x509"
	"strconv"
	"bytes"
)

func handleTCPMessage(messageChannel services.TCPMessageChannel, myPeer *services.Peer) error {
	messageType := binary.BigEndian.Uint16(messageChannel.Message[2:4])
	log.Println("Messagetype:", messageType)

	switch messageType {
		// ONION TUNNEL DESTROY
		case 563:
			handleOnionTunnelDestroy(messageChannel)
			break

		// ONION TUNNEL DATA
		case 564:
			handleOnionTunnelData(messageChannel, myPeer)
			break

		// CONSTRUCT TUNNEL
		case 567:
			newUDPConnection, err := handleConstructTunnel(messageChannel, myPeer)
			if err != nil {
				return err
			}

			myPeer.AppendNewUDPConnection(newUDPConnection)
			break

		// CONFIRM TUNNEL CONSTRUCTION
		case 568:
			handleConfirmTunnelConstruction(messageChannel, myPeer)
			break

		// TUNNEL INSTRUCTION
		case 569:
			log.Println("Messagetype: Tunnel Instruction")
			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
			log.Println("Tunnel Instruction for Tunnel: ", tunnelID)
			 // First, check if there is a right peer set for this tunnel id, if not decrypt data and exeute, if yes, simply forward it
			 // simply forward
			 if myPeer.PeerObject.TCPConnections[tunnelID].RightWriter != nil {
			 	log.Println("Right Writer exists, simply forward data to " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))
				 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(messageChannel.Message)
			 } else {
			 	// no right writer exists, handle data now to send the command
			 	// first, determine the command out of data and execute the function
			 	data := messageChannel.Message[8:]
			 	command := binary.BigEndian.Uint16(data[0:2])
			 	log.Println("Messagetype to execute: ", command)

				 switch command {
					 // Construct tunnel
					 case 567:
						log.Println("Messagetype: Construct Tunnel")
						 networkVersion := binary.BigEndian.Uint16(data[2:4])
						 var networkVersionString string
						 var ipAdd string
						 var port uint16
						 var destinationHostkey []byte
						 var lengthDestinationHostkey uint16
						 var endDestinationHostkey uint16
						 var pubKey []byte

						 if networkVersion == 0 {
							 networkVersionString = "IPv4"
							 ipAdd = net.IP(data[4:8]).String()
							 port = binary.BigEndian.Uint16(data[8:10])
							 lengthDestinationHostkey = binary.BigEndian.Uint16(data[10:12])
							 endDestinationHostkey = 12 +lengthDestinationHostkey
							 destinationHostkey = data[12:endDestinationHostkey]
							 pubKey = data[endDestinationHostkey:]
						 } else if networkVersion == 1 {
							 networkVersionString = "IPv6"
							 ipAdd = net.IP(messageChannel.Message[4:20]).String()
							 port = binary.BigEndian.Uint16(data[20:22])
							 lengthDestinationHostkey = binary.BigEndian.Uint16(data[22:24])
							 endDestinationHostkey = 24 +lengthDestinationHostkey
							 destinationHostkey = data[24:endDestinationHostkey]
							 pubKey = data[endDestinationHostkey:]
						 }


						 // now, create new TCP RightWriter for the right side
						 newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd, int(port))
						 log.Println(newTCPWriter)
						 if err != nil {
							log.Println(err.Error())
							 return errors.New("Error creating tcp writer, error: " + err.Error())
						 }
						 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter = newTCPWriter
						 constructMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, OriginHostkey: myPeer.PeerObject.TCPConnections[tunnelID].OriginHostkey, PublicKey: pubKey, TunnelID: tunnelID, DestinationAddress: ipAdd, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort: uint16(myPeer.PeerObject.P2P_Port)}
						 message := services.CreateConstructTunnelMessage(constructMessage)

						 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
						 log.Println("Send Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + " , Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))
						 break
					 default:
						 return errors.New("tcpMessagesController: Message Type not Found")
				 }
			 }

			break

		// CONFIRM TUNNEL INSTRUCTION
		case 570:
			log.Println("Messagetype: Confirm Tunnel Instruction")
			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])

			if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter != nil {
				log.Println("Right Writer exists, simply forward data to " + myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationPort))
				myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(messageChannel.Message)
			} else {
				data := messageChannel.Message[8:]
				log.Println("Received Confirm Tunnel Instruction for Tunnel: ", tunnelID)

				command := binary.BigEndian.Uint16(data[0:2])
				log.Println("We got a Confirmation for:")
				switch command {
				case 568:
					handleConfirmTunnelInnstructionConstruction(tunnelID, data[2:], myPeer)
					break
				}
			}
			break

		default:
			log.Println("Message not found, ignore")
			return nil
	}

	return nil
}

func handleOnionTunnelDestroy(messageChannel services.TCPMessageChannel) {
	log.Println("ONION TUNNEL DESTROY received")
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
	log.Printf("Tunnel ID: %s\n", tunnelID)
}


func handleOnionTunnelData(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Println("Messagetype: Onion Tunnel Data")
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
	//data := messageChannel.Message[8:]

	log.Printf("Tunnel ID: %d\n", tunnelID)

	myPeer.PeerObject.UDPConnections[tunnelID].RightWriter.Write(messageChannel.Message)
}


func handleConstructTunnel(messageChannel services.TCPMessageChannel, myPeer *services.Peer) (*models.UDPConnection, error) {
	log.Print("Messagetype: Handle Construct Tunnel")
	//var networkVersionString string
	//var destinationAddress string
	var destinationHostkey []byte
	var originHostkey []byte
	var lengthDestinationHostkey uint16
	var endDestinationHostkey uint16
	var lengthOriginHostkey uint16
	var endOriginHostkey uint16
	var pubKey []byte

	networkVersion := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	onionPort := binary.BigEndian.Uint16(messageChannel.Message[6:8])
	tcpPort := binary.BigEndian.Uint16(messageChannel.Message[8:10])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[10:14])

	if networkVersion == 0 {
		//networkVersionString = "IPv4"
		//destinationAddress = net.IP(messageChannel.Message[14:18]).String()
		lengthDestinationHostkey = binary.BigEndian.Uint16(messageChannel.Message[18:20])
		endDestinationHostkey = 20 + lengthDestinationHostkey
		destinationHostkey = messageChannel.Message[20:endDestinationHostkey]
		lengthOriginHostkey = binary.BigEndian.Uint16(messageChannel.Message[endDestinationHostkey:endDestinationHostkey+2])
		endOriginHostkey = endDestinationHostkey + 2 + lengthOriginHostkey
		originHostkey = messageChannel.Message[endDestinationHostkey+2:endOriginHostkey]

		pubKey = messageChannel.Message[endOriginHostkey:]

	} else if networkVersion == 1 {
		//networkVersionString = "IPv6"
		//destinationAddress = net.IP(messageChannel.Message[14:30]).String()
		lengthDestinationHostkey = binary.BigEndian.Uint16(messageChannel.Message[30:32])
		endDestinationHostkey = 32 + lengthDestinationHostkey
		destinationHostkey = messageChannel.Message[32:endDestinationHostkey]

		lengthOriginHostkey = binary.BigEndian.Uint16(messageChannel.Message[endDestinationHostkey:endDestinationHostkey+2])
		endOriginHostkey = endDestinationHostkey + 2 + lengthOriginHostkey
		originHostkey = messageChannel.Message[endDestinationHostkey + 2:endOriginHostkey]

		pubKey = messageChannel.Message[endOriginHostkey:]
	}

	// First, get ip address of sender
	ipAdd := services.GetIPOutOfAddr(messageChannel.Host)

	// Then, create the TCPWriter left
	newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd, int(tcpPort))
	if err != nil {
		return nil, errors.New("Error creating tcp writer, error: " + err.Error())
	}

	// Append the new TCPWriter as LeftTCPWriter to the TCP Connection
	myPeer.CreateInitialTCPConnection(tunnelID, destinationHostkey ,newTCPWriter, originHostkey)

	//  Now, create new UDP Connection with this "sender" as left side
	newUDPConnection, err := services.CreateInitialUDPConnectionLeft(ipAdd, int(onionPort), tunnelID)
	if err != nil {
		return nil, errors.New("handleConstructTunnel: " + err.Error())
	}

	// Now, create the crypto object and add it with the tunnel id to our peer
	// Create Pre Master Key for session.
	// Compute ephemeral Key
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[strconv.Itoa(int(tunnelID))] = &models.CryptoObject{TunnelId:tunnelID, PrivateKey:privateKey, PublicKey:publicKey, SessionKey:nil, Group:group}
	//encryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

	decryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

	err = saveEphemeralKey(decryptedPubKey, x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), tunnelID, myPeer)
	if err != nil {
		log.Fatal("Handle Construct Tunnel: Error while saving ephemeral key")

		// ToDo: ONION TUNNEL ERROR
	}

	originPubKey, err := x509.ParsePKCS1PublicKey(originHostkey)
	if err != nil {
		log.Fatal("Handle Construct Tunnel: Error while parsing Origin Public Key")
	}

	// Encrypt Pre-master secret with Origin Public Key
	encryptedPublicKey, err := services.EncryptKeyExchange(originPubKey, publicKey)

	// If everything worked out, send confirmTunnelConstruction back
	confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), PublicKey: encryptedPublicKey}
	message := services.CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction)

	// Sent confirm tunnel construction
	myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
	log.Println("Sent Confirm Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationPort))

	return newUDPConnection, nil
}

func handleConfirmTunnelConstruction(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Println("Messagetype: Confirm Tunnel construction")

	onionPort := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[6:10])
	lengthDestinationHostkey := binary.BigEndian.Uint16(messageChannel.Message[10:12])
	endDestinationHostkey := 12 + lengthDestinationHostkey
	destinationHostkey := messageChannel.Message[12:endDestinationHostkey]
	pubKey := messageChannel.Message[endDestinationHostkey:]

	log.Println("Sender: " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))

	// Check whether left writer exists. If yes, create Confirm Tunnel Instruction with ConfirmTunnelConstruction data.
	if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter != nil {
		// Now, create a udp righter to the right side (where we forwarded our tunnel construction to
		newRightUDPWriter, err := services.CreateUDPWriter(services.GetIPOutOfAddr(messageChannel.Host), int(onionPort))
		if err != nil {
			log.Println("Error creating UDP Writer right >> " +  err.Error())
		}
		myPeer.PeerObject.UDPConnections[tunnelID].RightWriter = newRightUDPWriter

		dataMessage := models.DataConfirmTunnelConstruction{DestinationHostkey: destinationHostkey, PublicKey: pubKey}
		data := services.CreateDataConfirmTunnelConstruction(dataMessage)
		confirmTunnelInstruction := models.ConfirmTunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateConfirmTunnelInstruction(confirmTunnelInstruction)
		log.Println("Final Destination exists??????")
		log.Println(myPeer.PeerObject.TCPConnections[tunnelID])

		myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
	} else {
		// TODO: I THINK ONLY HERE FORWARDING TO CHANNEL
		log.Println("Final Destination > PEER 0")
		log.Println(myPeer.PeerObject.TCPConnections[tunnelID].FinalDestination)

		// we reached peer 0 > create hashed hostkey of confirmation sender and add it to tunnelHostOrder
		newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
		if err != nil {
			log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey, ", err.Error())
		}
		hashedVersion := services.GenerateIdentityOfKey(newPublicKey)
		myPeer.PeerObject.TunnelHostOrder[tunnelID].PushBack(hashedVersion)

		decryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

		err = saveEphemeralKey(decryptedPubKey, destinationHostkey, tunnelID, myPeer)
		if err != nil{
			log.Fatal("Handle Confirm Tunnel Construction: Error while saving ephemeral key")

			// ToDo: ONION TUNNEL ERROR
		}

		// Now, create the UDP Writer Right for the UDP to the next right hop and assign it
		newUDPConnection, err := services.CreateInitialUDPConnectionRight(services.GetIPOutOfAddr(messageChannel.Host), int(onionPort), tunnelID)
		if err != nil {
			log.Println("handleConfirmTunnelConstruction while creating udp connection: " + err.Error())
		}
		myPeer.AppendNewUDPConnection(newUDPConnection)





		// TESTING
		_, pub, _ := services.ParseKeys("keypair3.pem")

		hashedVersion = services.GenerateIdentityOfKey(pub)

		destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)

		// Only for First Hop
		newIdentifier := strconv.Itoa(int(tunnelID)) + destinationHostkeyString

		// Generate Pre Master Key for session.
		privateKey, publicKey, group := services.GeneratePreMasterKey()
		myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId: tunnelID, PublicKey: publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}

		// Encrypt Public Key with destination key
		encryptedPubKey, err := services.EncryptKeyExchange(pub, publicKey)

		log.Println("TESTING: SEND TUNNEL INSTRUCTION")
		dataMessage := models.DataConstructTunnel{NetworkVersion: "IPv4", DestinationAddress: "192.168.0.15", Port: 4500, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub), PublicKey: encryptedPubKey}
		data := services.CreateDataConstructTunnel(dataMessage)

		// Now, just for tests, send a forward to a new peer
		tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateTunnelInstruction(tunnelInstructionMessage)

		myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
		log.Println("Sent Tunnel Instruction to " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))
	}
}

func handleConfirmTunnelInnstructionConstruction(tunnelId uint32, data []byte, myPeer *services.Peer) {
	log.Println("Messagetype: Confirm Tunnel Instruction Construction")
	log.Println("Got a confirmation for tunnel: " + strconv.Itoa(int(tunnelId)))

	lenghtDestinationKey := binary.BigEndian.Uint16(data[0:2])
	endOfDestinationHostkey := 2 + lenghtDestinationKey
	destinationHostkey := data[2:endOfDestinationHostkey]
	pubKey := data[endOfDestinationHostkey:]

	newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
	if err != nil {
		log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey")
	}

	// Now, create hashed version of the destination hostkey and add it to the TunnelHostOrder
	hashedVersion := services.GenerateIdentityOfKey(newPublicKey)
	myPeer.PeerObject.TunnelHostOrder[tunnelId].PushBack(hashedVersion)

	// Decrypt Pre master secret
	decryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

	// Save Ephemeral Key
	err = saveEphemeralKey(decryptedPubKey, destinationHostkey, tunnelId, myPeer)
	if err != nil {
		log.Fatal("Handle Confirm Tunnel Instruction Construction: Error while saving Ephemeral Key")
	}


	for e := myPeer.PeerObject.TunnelHostOrder[tunnelId].Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value) // print out the elements
	}

	if bytes.Equal(destinationHostkey, myPeer.PeerObject.TCPConnections[tunnelId].FinalDestination.DestinationHostkey) {
		log.Println("Yes, we've connected to our final destination")

		//onionTunnelReady := models.OnionTunnelReady{TunnelID: tunnelId, DestinationHostkey: myPeer.PeerObject.TCPConnections[tunnelId].FinalDestinationHostkey}
		//onionTunnelReadyMessage := services.CreateOnionTunnelReady(onionTunnelReady)

		// TODO: Send OnionTunnelReady to CM/UI module.
		//log.Println(onionTunnelReadyMessage)

		for e := myPeer.PeerObject.TunnelHostOrder[tunnelId].Front(); e != nil; e = e.Next() {
			fmt.Println(e.Value) // print out the elements
		}
	}
}

// Save Ephemeral Key
func saveEphemeralKey(PublicKey []byte, destinationHostkey []byte, tunnelID uint32, myPeer *services.Peer) (error) {
	log.Println("EXCHANGE KEY")

	newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
	if err != nil {
		return errors.New("Couldn't convert []byte destinationHostKey to rsa Publickey")
	}

	hashedIdentity := services.GenerateIdentityOfKey(newPublicKey)

	// Compute Ephemeral Key
	// First, generate identifier
	destinationHostkeyString := fmt.Sprintf("%s", hashedIdentity)

	// Get identifier for crypto object
	var identifier string;
	if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter == nil {
		identifier = strconv.Itoa(int(tunnelID)) + destinationHostkeyString
	} else {
		identifier = strconv.Itoa(int(tunnelID))
	}
	cryptoObject := myPeer.PeerObject.CryptoSessionMap[identifier]

	// Set new session Key if not set.
	if cryptoObject.SessionKey == nil {
		sessionKey := services.ComputeEphemeralKey(cryptoObject.Group, PublicKey, cryptoObject.PrivateKey)
		cryptoObject.SessionKey = sessionKey
		log.Println("Created session key.")
	}

	return nil
}

func handleRPSPeer(messageChannel services.TCPMessageChannel, myPeer *services.Peer) (models.RPSPeer) {
	// Number of Portmaps
	numberPortmaps := binary.BigEndian.Uint16(messageChannel.Message[4:6])

	var portmap uint16
	var onionPort uint16

	// Get Portmaps
	for i := 0; i < int(numberPortmaps); i++ {
		// Get portmap type
		portmap = binary.BigEndian.Uint16(messageChannel.Message[6+i*4:8+i*4])

		// Onion Portmap
		if portmap == 560 {
			onionPort = binary.BigEndian.Uint16(messageChannel.Message[8+i*4:10+i*4])
		}
	}

	// Get IP Address and hostkey
	startIndexAddress := 6 + (numberPortmaps * 4)
	networkVersion := binary.BigEndian.Uint16(messageChannel.Message[6:8])

	var peerHostkey []byte
	var networkVersionString string
	var destinationAddress string

	if networkVersion == 0 {
		networkVersionString = "IPv4"
		destinationAddress = net.IP(messageChannel.Message[startIndexAddress:startIndexAddress+4]).String()
		peerHostkey = messageChannel.Message[startIndexAddress+4:]
	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
		destinationAddress = net.IP(messageChannel.Message[startIndexAddress:startIndexAddress+16]).String()
		peerHostkey = messageChannel.Message[startIndexAddress+16:]
	}

	return models.RPSPeer{NetworkVersion: networkVersionString, OnionPort: onionPort, DestinationAddress: destinationAddress, PeerHostkey: peerHostkey}
}