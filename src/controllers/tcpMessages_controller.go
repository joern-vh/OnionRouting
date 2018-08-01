package controllers

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"services"

	"models"
	"fmt"
	"container/list"
	"crypto/x509"
	"bytes"
	"strconv"
)

func StartTCPController(myPeer *services.Peer) {
	log.Println("StartTCPController: Started TCP Controller")
	go func() {
		for msg := range services.CommunicationChannelTCPMessages {
			log.Println("\n\n")
			log.Println("New message from " + msg.Host)
			handleTCPMessage(msg, myPeer)
		}
	}()
}

func handleTCPMessage(messageChannel services.TCPMessageChannel, myPeer *services.Peer) error {
	messageType := binary.BigEndian.Uint16(messageChannel.Message[2:4])
	log.Println("Messagetype:", messageType)

	switch messageType {
		// ONION TUNNEL BUILD
		case 560:
			handleOnionTunnelBuild(messageChannel, myPeer)
			break

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
							 if err != nil {
								 return errors.New("Error creating tcp writer, error: " + err.Error())
							 }
							 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter = newTCPWriter

							 //constructMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, DestinationAddress: ipAdd, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort:uint16(myPeer.PeerObject.P2P_Port), TunnelID: tunnelID, PublicKey: pubKey}
							 constructMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, OriginHostkey:destinationHostkey, PublicKey: pubKey, TunnelID: tunnelID, DestinationAddress: ipAdd, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort: uint16(myPeer.PeerObject.P2P_Port)}
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
			return errors.New("tcpMessagesController: Message Type not Found")
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

	//Construct Tunnel Message
	newTunnelID := services.CreateTunnelID()
	log.Println("NewTunnelID: ", newTunnelID)
	log.Println("IP-Address of destination: ", destinationAddress)
	log.Print("Onion Port of destination: ", onionPort)

	destinationPubKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
	hashedVersion := services.GenerateIdentityOfKey(destinationPubKey)
	destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)

	// Only for First Hop
	newIdentifier := strconv.Itoa(int(newTunnelID)) + destinationHostkeyString

	// Generate Pre Master Key for session.
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId: newTunnelID, PublicKey: publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}

	//encryptedPubKey, err := services.EncryptKeyExchange(destinationPubKey, publicKey)

	// Build Construct Tunnel Message
	constructTunnelMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, OriginHostkey:destinationHostkey, PublicKey: publicKey, TunnelID: newTunnelID, DestinationAddress: destinationAddress, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort: uint16(myPeer.PeerObject.P2P_Port)}
	message := services.CreateConstructTunnelMessage(constructTunnelMessage)

	newTCPWriter, err := myPeer.CreateTCPWriter(destinationAddress, int(onionPort))		// TODO: Make dynamic
	if err != nil {
		log.Println("Error creating tcp writer, error: " + err.Error())
	}

	// Generate hash of the final destination hoskey
	newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
	if err != nil {
		log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey, ", err.Error())
	}
	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID] = &models.TCPConnection{constructTunnelMessage.TunnelID, nil, newTCPWriter, list.New(), services.GenerateIdentityOfKey(newPublicKey)}
	// Now just add the right connection to the map with status pending
	//myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].ConnectionOrder = append(myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].ConnectionOrder, models.ConnnectionOrderObject{TunnelId:constructTunnelMessage.TunnelID, IpAddress:destinationAddress, IpPort:4200, Confirmed:false})
	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].RightWriter.TCPWriter.Write(message)
}

func handleOnionTunnelDestroy(messageChannel services.TCPMessageChannel) {
	log.Println("ONION TUNNEL DESTROY received")
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
	log.Printf("Tunnel ID: %s\n", tunnelID)
}


func handleOnionTunnelData(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
	data := messageChannel.Message[8:]

	log.Printf("Tunnel ID: %s\n", tunnelID)
	log.Printf("Tunnel ID: %x\n", data)
}


func handleConstructTunnel(messageChannel services.TCPMessageChannel, myPeer *services.Peer) (*models.UDPConnection, error) {
	log.Print("Messagetype: Handle Construct Tunnel")
	var networkVersionString string
	//var destinationAddress string
	var destinationHostkey []byte
	//var originHostkey []byte
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
		networkVersionString = "IPv4"
		//destinationAddress = net.IP(messageChannel.Message[14:18]).String()
		lengthDestinationHostkey = binary.BigEndian.Uint16(messageChannel.Message[18:20])
		endDestinationHostkey = 20 + lengthDestinationHostkey
		destinationHostkey = messageChannel.Message[20:endDestinationHostkey]
		lengthOriginHostkey = binary.BigEndian.Uint16(messageChannel.Message[endDestinationHostkey:endDestinationHostkey+2])
		endOriginHostkey = endDestinationHostkey + 2 + lengthOriginHostkey
		//originHostkey = messageChannel.Message[endDestinationHostkey+2:endOriginHostkey]

		pubKey = messageChannel.Message[endOriginHostkey:]

	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
		//destinationAddress = net.IP(messageChannel.Message[14:30]).String()
		lengthDestinationHostkey = binary.BigEndian.Uint16(messageChannel.Message[30:32])
		endDestinationHostkey = 32 + lengthDestinationHostkey
		destinationHostkey = messageChannel.Message[32:endDestinationHostkey]

		lengthOriginHostkey = binary.BigEndian.Uint16(messageChannel.Message[endDestinationHostkey:endDestinationHostkey+2])
		endOriginHostkey = endDestinationHostkey + 2 + lengthOriginHostkey
		//originHostkey = messageChannel.Message[endDestinationHostkey + 2:endOriginHostkey]

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
	myPeer.CreateInitialTCPConnection(tunnelID, destinationHostkey ,newTCPWriter)

	//  Now, create new UDP Connection with this "sender" as left side
	newUDPConnection, err := services.CreateInitialUDPConnection(ipAdd, int(onionPort), tunnelID, networkVersionString)
	if err != nil {
		return nil, errors.New("handleConstructTunnel: " + err.Error())
	}

	// Now, create the crypto object and add it with the tunnel id to our peer
	// Create Pre Master Key for session.
	// Compute ephemeral Key
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[strconv.Itoa(int(tunnelID))] = &models.CryptoObject{TunnelId:tunnelID, PrivateKey:privateKey, PublicKey:publicKey, SessionKey:nil, Group:group}
	//encryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)
	err = saveEphemeralKey(pubKey, x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), tunnelID, myPeer)
	if err != nil {
		log.Fatal("Handle Construct Tunnel: Error while saving ephemeral key")

		// ToDo: ONION TUNNEL ERROR
	}

	//originPubKey, err := x509.ParsePKCS1PublicKey(originHostkey)
	if err != nil {
		log.Fatal("Handle Construct Tunnel: Error while parsing Origin Public Key")
	}

	// Encrypt Pre-master secret with Origin Public Key
	//encryptedPublicKey, err := services.EncryptKeyExchange(originPubKey, publicKey)

	// If everything worked out, send confirmTunnelConstruction back
	
	confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: destinationHostkey, PublicKey: publicKey}
	message := services.CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction)

	// Encrypt Message with Public Key of destination. RSA
	//destinationPubKey, err := x509.ParsePKCS1PublicKey(originHostkey)
	//encryptedMessage, err := services.EncryptKeyExchange(destinationPubKey, message)

	// Sent confirm tunnel construction
	myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
	log.Println("Sent Confirm Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationPort))

	return newUDPConnection, nil
}

func handleConfirmTunnelConstruction(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Println("Messagetype: Confirm Tunnel construction")

	//onionPort := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[6:10])
	lengthDestinationHostkey := binary.BigEndian.Uint16(messageChannel.Message[10:12])
	endDestinationHostkey := 12 + lengthDestinationHostkey
	destinationHostkey := messageChannel.Message[12:endDestinationHostkey]
	pubKey := messageChannel.Message[endDestinationHostkey:]

	log.Println("Sender: " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))

	// Check whether left writer exists. If yes, create Confirm Tunnel Instruction with ConfirmTunnelConstruction data.
	if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter != nil {
		// Forward to left a confirmTunnelInstruction

		dataMessage := models.DataConfirmTunnelConstruction{DestinationHostkey: destinationHostkey, PublicKey: pubKey}
		data := services.CreateDataConfirmTunnelConstruction(dataMessage)
		confirmTunnelInstruction := models.ConfirmTunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateConfirmTunnelInstruction(confirmTunnelInstruction)

		myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
	} else {
		// Add hostkey to the list of available host, but first, convert it
		newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
		hashedVersion := services.GenerateIdentityOfKey(newPublicKey)
		if err != nil {
			log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey, ", err.Error())
		}
		myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder.PushFront(hashedVersion)

		// Now, create a cryptoobject and add it to the hashmap of the tcpConnection
		// First, generate identifier
		//destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)

		// ToDo: Remove!
		//newIdentifier := strconv.Itoa(int(tunnelID)) + destinationHostkeyString

		// Generate Pre Master Key for session.
		//privateKey, publicKey, group := services.GeneratePreMasterKey()
		//myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId:tunnelID, PublicKey:publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}



		//encryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

		err = saveEphemeralKey(pubKey, destinationHostkey, tunnelID, myPeer)
		if err != nil{
			log.Fatal("Handle Confirm Tunnel Construction: Error while saving ephemeral key")

			// ToDo: ONION TUNNEL ERROR
		}

		// We received a confirmation from out final destination, send ready message
		if bytes.Equal(hashedVersion, myPeer.PeerObject.TCPConnections[tunnelID].FinalDestinationHostkey) {
			log.Println("Yes, we'are connected to our final destination")

			//onionTunnelReady := models.OnionTunnelReady{TunnelID: tunnelID, DestinationHostkey: myPeer.PeerObject.TCPConnections[tunnelID].FinalDestinationHostkey}
			//onionTunnelReadyMessage := services.CreateOnionTunnelReady(onionTunnelReady)

			// TODO: Send OnionTunnelReady to CM/UI module.
			//log.Println(onionTunnelReadyMessage)
		}

		// KEY EXCHANGE TESTING

		// Encrypt Key with RSA
		/*encryptedKey, err := services.EncryptKeyExchange(newPublicKey, myPeer.PeerObject.CryptoSessionMap[newIdentifier].PublicKey)

		if err != nil {
			log.Println("Error encrypting Key", err.Error())
		}

		keyExchange := models.ExchangeKey{PublicKey: encryptedKey, TunnelID: tunnelID, DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey)}
		keyExchangeMessage := services.CreateExchangeKey(keyExchange)

		myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(keyExchangeMessage)*/







		// TESTING

		_, pub, _ := services.ParseKeys("keypair.pem")

		hashedVersion = services.GenerateIdentityOfKey(pub)

		destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)

		// Only for First Hop
		newIdentifier := strconv.Itoa(int(tunnelID)) + destinationHostkeyString

		// Generate Pre Master Key for session.
		privateKey, publicKey, group := services.GeneratePreMasterKey()
		myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId: tunnelID, PublicKey: publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}

		//privateKey, publicKey, group := services.GeneratePreMasterKey()
		//myPeer.PeerObject.CryptoSessionMap[strconv.Itoa(int(tunnelID))] = &models.CryptoObject{TunnelId:tunnelID, PrivateKey:privateKey, PublicKey:publicKey, SessionKey:nil, Group:group}
		//encryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

		log.Println("TESTING: SEND TUNNEL INSTRUCTION")
		dataMessage := models.DataConstructTunnel{NetworkVersion: "IPv4", DestinationAddress: "192.168.0.15", Port: 4500, DestinationHostkey: x509.MarshalPKCS1PublicKey(pub), PublicKey: publicKey}
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
	// State now: just add hostkey
	// Add hostkey to the list of available host, but first, convert it

	lenghtDestinationKey := binary.BigEndian.Uint16(data[0:2])
	endOfDestinationHostkey := 2 + lenghtDestinationKey
	destinationHostkey := data[2:endOfDestinationHostkey]
	pubKey := data[endOfDestinationHostkey:]

	newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
	if err != nil {
		log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey")
	}
	myPeer.PeerObject.TCPConnections[tunnelId].ConnectionOrder.PushBack(services.GenerateIdentityOfKey(newPublicKey))

	// Iterate through list and print its contents.
	/*for e := myPeer.PeerObject.TCPConnections[tunnelId].ConnectionOrder.Front(); e != nil; e = e.Next() {
		fmt.Println("List value ", e.Value)
	}*/

	hashedVersion := services.GenerateIdentityOfKey(newPublicKey)

	// Now, create a cryptoobject and add it to the hashmap of the tcpConnection
	// First, generate identifier
	//destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)
	//newIdentifier := strconv.Itoa(int(tunnelId)) + destinationHostkeyString
	//privateKey, publicKey, group := services.GeneratePreMasterKey()
	//myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId:tunnelId, PublicKey:publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}

	//encryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

	err = saveEphemeralKey(pubKey, destinationHostkey, tunnelId, myPeer)
	if err != nil {
		log.Fatal("Handle Confirm Tunnel Instruction Construction: Error while saving Ephemeral Key")

		// ToDo: ONION TUNNEL ERROR
	}

	// We received a confirmation from out final destination, send ready message
	if bytes.Equal(hashedVersion, myPeer.PeerObject.TCPConnections[tunnelId].FinalDestinationHostkey) {
		log.Println("Yes, we've connected to our final destination")

		//onionTunnelReady := models.OnionTunnelReady{TunnelID: tunnelId, DestinationHostkey: myPeer.PeerObject.TCPConnections[tunnelId].FinalDestinationHostkey}
		//onionTunnelReadyMessage := services.CreateOnionTunnelReady(onionTunnelReady)

		// TODO: Send OnionTunnelReady to CM/UI module.
		//log.Println(onionTunnelReadyMessage)
	}
}

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

	//encryptedPublicKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, PublicKey)

	// Set new session Key if not set.
	if cryptoObject.SessionKey == nil {
		sessionKey := services.ComputeEphemeralKey(cryptoObject.Group, PublicKey, cryptoObject.PrivateKey)
		cryptoObject.SessionKey = sessionKey
		log.Println("Created session key.")
	}

	return nil
}
