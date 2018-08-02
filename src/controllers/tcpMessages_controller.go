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
	messageTypeSize := binary.BigEndian.Uint16(messageChannel.Message[2:4])

	encryptedMessageType, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, messageChannel.Message[4:4+messageTypeSize])

	var messageType uint16
	if err == nil {
		messageType = binary.BigEndian.Uint16(encryptedMessageType)
		messageTypeArray := append(messageChannel.Message[0:2], encryptedMessageType...)
		log.Println(len(messageTypeArray))
		messageChannel.Message = append(messageTypeArray, messageChannel.Message[4+messageTypeSize:]...)
	}

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
				messageChannel.Message = append(messageChannel.Message, []byte("\r\n")...)
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

						 log.Println(networkVersionString)

						 // now, create new TCP RightWriter for the right side
						 newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd, int(port))
						 log.Println(newTCPWriter)
						 if err != nil {
							log.Println(err.Error())
							 return errors.New("Error creating tcp writer, error: " + err.Error())
						 }
						 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter = newTCPWriter
						 constructMessage := models.ConstructTunnel{DestinationHostkey: destinationHostkey, OriginHostkey: myPeer.PeerObject.TCPConnections[tunnelID].OriginHostkey, PublicKey: pubKey, TunnelID: tunnelID, OnionPort: uint16(myPeer.PeerObject.UDPPort)}

						 senderHostkeyPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)

						 message := services.CreateConstructTunnelMessage(constructMessage, nil, senderHostkeyPublicKey)

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
				messageChannel.Message = append(messageChannel.Message, []byte("\r\n")...)

				// Encrypt Message
				myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(messageChannel.Message)
			} else {
				// Decrypt message

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

		// EXCHANGE KEY
		case 571:
			handleExchangeKey(messageChannel, myPeer)
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


	onionPort := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[6:10])
	//log.Println("TEEEEESSSTTTTTTTTTTTTTT")
	//log.Println(myPeer.PeerObject.TCPConnections[tunnelID])
	//log.Println(myPeer.PeerObject.TCPConnections[tunnelID].LeftHostkey)

	// Decrypt Message

	// Sender Hostkey
	lengthSenderHostkey := binary.BigEndian.Uint16(messageChannel.Message[10:12])
	endSenderHostkey := 12 + lengthSenderHostkey
	senderHostkey := messageChannel.Message[12:endSenderHostkey]

	// Origin hostkey
	lengthOriginHostkey := binary.BigEndian.Uint16(messageChannel.Message[endSenderHostkey:endSenderHostkey+2])
	endOriginHostkey := endSenderHostkey + 2 + lengthOriginHostkey
	originHostkey := messageChannel.Message[endSenderHostkey+2:endOriginHostkey]

	// Pre-master secret
	pubKey := messageChannel.Message[endOriginHostkey:]

	// UDP Connection
	// First, get ip address of sender
	ipAdd := services.GetIPOutOfAddr(messageChannel.Host)
	//  Now, create new UDP Connection with this "sender" as left side
	newUDPConnection, err := services.CreateInitialUDPConnectionLeft(ipAdd, int(onionPort), tunnelID)
	if err != nil {
		return nil, errors.New("handleConstructTunnel: " + err.Error())
	}

	// Save origin Pubkey
	// ToDo: Watch for errors
	myPeer.PeerObject.TCPConnections[tunnelID].LeftHostkey = senderHostkey
	myPeer.PeerObject.TCPConnections[tunnelID].OriginHostkey = originHostkey
	//log.Println("TCP Connection: ", myPeer.PeerObject.TCPConnections[tunnelID])

	// Get Identifier for crypto object
	senderHostkeyPublicKey, err := x509.ParsePKCS1PublicKey(myPeer.PeerObject.TCPConnections[tunnelID].LeftHostkey)
	if err != nil {
		log.Println("Handle Construct Tunnel: Couldn't convert []byte sender hostkey to rsa Publickey")
	}
	hashedIdentity := services.GenerateIdentityOfKey(senderHostkeyPublicKey)
	senderHostkeyString := fmt.Sprintf("%s", hashedIdentity)
	// Get identifier for crypto object
	identifier := strconv.Itoa(int(tunnelID)) + senderHostkeyString

	// If Crypto Object exists
	if myPeer.PeerObject.CryptoSessionMap[identifier] != nil {
		log.Println("Crypto Object exists.")

		decryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)
		if err != nil {
			log.Println("Handle Construct Tunnel: Failed Decrypting Pre-master key")
		}
		encryptedPubKey, err := services.EncryptKeyExchange(senderHostkeyPublicKey, decryptedPubKey)

		// Reply to construct tunnel
		confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), PublicKey: encryptedPubKey}
		message := services.CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction, senderHostkeyPublicKey)

		// Sent confirm tunnel construction
		myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
		log.Println("Sent Confirm Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationPort))
	} else {
		// Create Pre Master Key for session with origin peer.
		// Compute ephemeral Key
		privateKey, publicKey, group := services.GeneratePreMasterKey()
		myPeer.PeerObject.CryptoSessionMap[identifier] = &models.CryptoObject{TunnelId:tunnelID, PrivateKey:privateKey, PublicKey:publicKey, SessionKey: nil, Group: group}

		log.Println(x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey))
		decryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

		err = saveEphemeralKey(decryptedPubKey, x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), tunnelID, myPeer)
		if err != nil {
			log.Fatal("Handle Construct Tunnel: Error while saving ephemeral key")
		}

		// Reply to construct tunnel
		originPubKey, err := x509.ParsePKCS1PublicKey(originHostkey)
		if err != nil {
			log.Fatal("Handle Construct Tunnel: Error while parsing Origin Public Key")
		}

		// Encrypt Pre-master secret with Origin Public Key
		encryptedPublicKey, err := services.EncryptKeyExchange(originPubKey, publicKey)

		// If everything worked out, send confirmTunnelConstruction back
		confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), PublicKey: encryptedPublicKey}
		message := services.CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction, senderHostkeyPublicKey)

		// Sent confirm tunnel construction
		myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
		log.Println("Sent Confirm Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationPort))

	}
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

		// TODO: I THINK ONLY HERE FORWARDING TO CHANNEL
		//services.CommunicationChannelTCPConfirm <- services.ConfirmMessageChannel{TunnelId:tunnelID}
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

	// TODO: perhaps yes here too????
	services.CommunicationChannelTCPConfirm <- services.ConfirmMessageChannel{TunnelId:tunnelId}
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
		log.Println("Test")
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
		log.Println(sessionKey)
	}

	return nil
}

func handleExchangeKey(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Println("Messagetype: Exchange Key")

	status := binary.BigEndian.Uint16(messageChannel.Message[6:8])
	sizeTunnelID := binary.BigEndian.Uint16(messageChannel.Message[8:10])
	endTunnelID := 10 + sizeTunnelID
	tunnelIDBytes, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, messageChannel.Message[10:endTunnelID])
	if err != nil {
		log.Println("Exchange Key: Failed encrypting Tunnel ID.")
	}

	tunnelID := binary.BigEndian.Uint32(tunnelIDBytes)

	// Parse sender hostkey and generate identity
	sizeSenderHostkey := binary.BigEndian.Uint16(messageChannel.Message[endTunnelID:endTunnelID+2])
	endSenderHostkey := endTunnelID + 2 + sizeSenderHostkey

	senderHostkey := messageChannel.Message[endTunnelID+2:endSenderHostkey]
	senderHostkeyPublicKey, err := x509.ParsePKCS1PublicKey(senderHostkey)
	if err != nil {
		log.Println("Handle Exchange Key: Couldn't convert []byte sender hostkey to rsa Publickey")
	}
	hashedIdentity := services.GenerateIdentityOfKey(senderHostkeyPublicKey)

	// Compute Ephemeral Key
	// First, generate identifier
	senderHostkeyString := fmt.Sprintf("%s", hashedIdentity)

	// Get identifier for crypto object
	identifier := strconv.Itoa(int(tunnelID)) + senderHostkeyString

	// Get transferred public key
	senderPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, messageChannel.Message[endSenderHostkey:])
	if err != nil {
		log.Println("Exchange Key: Failed encrypting Tunnel ID.")
	}

	// Start key exchange
	if status == 0 {
		log.Println("HERE")
		privateKey, publicKey, group := services.GeneratePreMasterKey()
		sessionKey := services.ComputeEphemeralKey(group, senderPubKey, privateKey)

		myPeer.PeerObject.CryptoSessionMap[identifier] = &models.CryptoObject{TunnelId: tunnelID, Group: group, PrivateKey: privateKey, PublicKey: publicKey, SessionKey: sessionKey}
		// Create response key exchange message
		exchangeKey := models.ExchangeKey{TunnelID: tunnelID, Status: uint16(1), DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), PublicKey: publicKey}
		exchangeKeyMessage := services.CreateExchangeKey(exchangeKey, senderHostkeyPublicKey, senderHostkeyPublicKey)
		// Append Delimiter
		exchangeKeyMessage = append(exchangeKeyMessage, []byte("\r\n")...)


		// First, get ip address of sender
		ipAdd := services.GetIPOutOfAddr(messageChannel.Host)
		tcpPort := binary.BigEndian.Uint16(messageChannel.Message[4:6])

		newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd, int(tcpPort))
		if err != nil {
			log.Println("ERROR: ", err.Error())
		}

		// Append the new TCPWriter as LeftTCPWriter to the TCP Connection
		myPeer.CreateInitialTCPConnection(tunnelID, senderHostkey, newTCPWriter)
		//log.Println("TCP Connection: !!!!! ", myPeer.PeerObject.TCPConnections[tunnelID])

		// Send exchange key reply
		myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(exchangeKeyMessage)
	} else if status == 1 {
		log.Println("Handle Exchange Key: Received key exchange response.")
		cryptoObject := myPeer.PeerObject.CryptoSessionMap[identifier]

		sessionKey := services.ComputeEphemeralKey(cryptoObject.Group, senderPubKey, cryptoObject.PrivateKey)
		cryptoObject.SessionKey = sessionKey
		log.Println("Created session key.")

		// Save sender hostkey
		myPeer.PeerObject.TCPConnections[tunnelID].RightHostkey = senderHostkey

		//log.Println("Right Hostkey")
		//log.Println(myPeer.PeerObject.TCPConnections[tunnelID])

		// If origin peer.
		if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter == nil {
			log.Println("Origin Peer")
			encryptedPubKey, err := services.EncryptKeyExchange(senderHostkeyPublicKey, cryptoObject.PublicKey)
			if err != nil {
				log.Println("Handle Exchange Key: Error while encrypting Pub Key for Construct Tunnel")
			}
			constructTunnel := models.ConstructTunnel{OnionPort: uint16(myPeer.PeerObject.UDPPort), TunnelID: tunnelID, OriginHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), PublicKey: encryptedPubKey}
			constructTunnelMessage := services.CreateConstructTunnelMessage(constructTunnel, cryptoObject.SessionKey, senderHostkeyPublicKey)

			// Send exchange key reply
			myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(constructTunnelMessage)
		}
	}
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