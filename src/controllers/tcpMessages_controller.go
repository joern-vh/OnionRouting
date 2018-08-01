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
	"container/list"
	"bytes"
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

					 if networkVersion == 0 {
						 networkVersionString = "IPv4"
						 ipAdd = net.IP(data[4:8]).String()
						 port = binary.BigEndian.Uint16(data[8:10])
						 destinationHostkey = messageChannel.Message[10:]
					 } else if networkVersion == 1 {
						 networkVersionString = "IPv6"
						 ipAdd = net.IP(messageChannel.Message[4:20]).String()
						 port = binary.BigEndian.Uint16(data[20:22])
						 destinationHostkey = messageChannel.Message[22:]
					 }

						 // now, create new TCP RightWriter for the right side
						 newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd, int(port))
						 if err != nil {
							 return errors.New("Error creating tcp writer, error: " + err.Error())
						 }
						 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter = newTCPWriter
						 constructMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, DestinationAddress: ipAdd, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort:uint16(myPeer.PeerObject.P2P_Port), TunnelID: tunnelID}
						 message := services.CreateConstructTunnelMessage(constructMessage)

					 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
					 log.Println("Send Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + " , Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))
					 break


				case 571:
				 		handleExchangeKey(messageChannel, data[2:], myPeer)
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
			data := messageChannel.Message[8:]
			log.Println("Received Confirm Tunnel Instruction for Tunnel: ", tunnelID)

			command := binary.BigEndian.Uint16(data[0:2])
			log.Println("We got a Conformation for:")
			switch command {
			case 568:
				handleConfirmTunnelInnstructionConstruction(tunnelID, data[2:], myPeer)
				break
			}
			break

		// EXCHANGE KEY
		case 571:
			handleExchangeKey(messageChannel, nil, myPeer)
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
	constructTunnelMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, TunnelID: newTunnelID, DestinationAddress: destinationAddress, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort: uint16(myPeer.PeerObject.P2P_Port)}
	message := services.CreateConstructTunnelMessage(constructTunnelMessage)

	newTCPWriter, err := myPeer.CreateTCPWriter(destinationAddress, int(onionPort))
	if err != nil {
		log.Println("Error creating tcp writer, error: " + err.Error())
	}

	// Initialize list for TunnelHostOrder with new TunnelID as Key for the Hashmap
	myPeer.PeerObject.TunnelHostOrder[newTunnelID] = new(list.List)
	// then, generate the hashed version of the destinationHostKey and add it as first value to the list
	myPeer.PeerObject.TunnelHostOrder[newTunnelID].PushBack(services.GenerateIdentityOfKey(myPeer.PeerObject.PublicKey))

	// now, add the new TCP Connection to the peer under TCPConnections, indentified by the newTunnelid
	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID] = &models.TCPConnection{constructTunnelMessage.TunnelID, nil, newTCPWriter, &models.OnionTunnelBuild{DestinationHostkey: destinationHostkey, Port: onionPort, DestinationAddress: destinationAddress, NetworkVersion: networkVersionString}}

	// at last, send the constructTunnelMessage
	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].RightWriter.TCPWriter.Write(message)
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
	//var destinationAddress string
	var destinationHostkey []byte

	networkVersion := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	onionPort := binary.BigEndian.Uint16(messageChannel.Message[6:8])
	tcpPort := binary.BigEndian.Uint16(messageChannel.Message[8:10])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[10:14])

	if networkVersion == 0 {
		//destinationAddress = net.IP(messageChannel.Message[14:18]).String()
		destinationHostkey = messageChannel.Message[18:]
	} else if networkVersion == 1 {
		//destinationAddress = net.IP(messageChannel.Message[14:30]).String()
		destinationHostkey = messageChannel.Message[30:]
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
	newUDPConnection, err := services.CreateInitialUDPConnectionLeft(ipAdd, int(onionPort), tunnelID)
	if err != nil {
		return nil, errors.New("handleConstructTunnel: " + err.Error())
	}

	// Now, create the crypto object and add it with the tunnel id to our peer
	// Create Pre Master Key for session.
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[strconv.Itoa(int(tunnelID))] = &models.CryptoObject{TunnelId:tunnelID, PrivateKey:privateKey, PublicKey:publicKey, SessionKey:nil, Group:group}

	// If everything worked out, send confirmTunnelConstruction back
	confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey)}
	message := services.CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction)

	myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
	log.Println("Sent Confirm Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationPort))

	return newUDPConnection, nil
}

func handleConfirmTunnelConstruction(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Println("Messagetype: Confirm Tunnel construction")

	onionPort := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[6:10])
	destinationHostkey := messageChannel.Message[10:]

	log.Println("Sender: " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))

	if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter != nil {
		// Now, create a udp righter to the right side (where we forwarded our tunnel construction to
		newRightUDPWriter, err := services.CreateUDPWriter(services.GetIPOutOfAddr(messageChannel.Host), int(onionPort))
		if err != nil {
			log.Println("Error creating UDP Writer right >> " +  err.Error())
		}
		myPeer.PeerObject.UDPConnections[tunnelID].RightWriter = newRightUDPWriter

		// Forward to left a confirmTunnelInstruction
		dataMessage := models.DataConfirmTunnelConstruction{DestinationHostkey: destinationHostkey}
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





		// Now, create a cryptoobject and add it to the hashmap of the tcpConnection
		// First, generate identifier
		destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)
		newIdentifier := strconv.Itoa(int(tunnelID)) + destinationHostkeyString

		// Generate Pre Master Key for session.
		privateKey, publicKey, group := services.GeneratePreMasterKey()
		myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId:tunnelID, PublicKey:publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}



		// Convert messageType to Byte array
		/*messageTypeBuf := new(bytes.Buffer)
		binary.Write(messageTypeBuf, binary.BigEndian, uint16(567))
		data := messageTypeBuf.Bytes()
		*/
		//log.Println(myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder)
		//ip := net.IP(data[2:6]).String()
		//log.Println(ip)
		// geht the element wit the right ip and set its value to confirm: true
		/*for i := range myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder {
			if myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder[i].IpAddress == ip {
				myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder[i].Confirmed = true
			}
		}*/

		//ipAddr := net.ParseIP("192.168.0.15")
		//log.Println(myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder)

		/*data = append(data, ipAddr.To4()...)

		portBuf := new(bytes.Buffer)
		binary.Write(portBuf, binary.BigEndian, uint16(4200))
		data = append(data, portBuf.Bytes()...)*/


		// KEY EXCHANGE TESTING

		// Encrypt Key with RSA
		encryptedKey, err := services.EncryptKeyExchange(newPublicKey, myPeer.PeerObject.CryptoSessionMap[newIdentifier].PublicKey)

		if err != nil {
			log.Println("Error encrypting Key", err.Error())
		}

		keyExchange := models.ExchangeKey{PublicKey: encryptedKey, TunnelID: tunnelID, DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey)}
		keyExchangeMessage := services.CreateExchangeKey(keyExchange)

		myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(keyExchangeMessage)

		// Now, create the UDP Writer Right for the UDP to the next right hop and assign it
		newUDPConnection, err := services.CreateInitialUDPConnectionRight(services.GetIPOutOfAddr(messageChannel.Host), int(onionPort), tunnelID)
		if err != nil {
			log.Println("handleConfirmTunnelConstruction while creating udp connection: " + err.Error())
		}
		myPeer.AppendNewUDPConnection(newUDPConnection)
		// OLD TESTING

		log.Println("TESTING: SEND TUNNEL INSTRUCTION")
		dataMessage := models.DataConstructTunnel{NetworkVersion: "IPv4", DestinationAddress: "192.168.0.10", Port: 4500, DestinationHostkey: destinationHostkey}
		data := services.CreateDataConstructTunnel(dataMessage)

		// Now, just for tests, send a forward to a new peer
		tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateTunnelInstruction(tunnelInstructionMessage)

		myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
		log.Println("Sent Tunnel Instruction to " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))
	}
}

func handleConfirmTunnelInnstructionConstruction(tunnelId uint32, destinationHostKey []byte, myPeer *services.Peer) {
	log.Println("Messagetype: Confirm Tunnel Instruction Construction")
	log.Println("Got a conformation for tunnel: " + strconv.Itoa(int(tunnelId)))

	// Now, create hashed version of the destination hostkey and add it to the TunnelHostOrder
	newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostKey)
	if err != nil {
		log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey")
	}
	hashedVersion := services.GenerateIdentityOfKey(newPublicKey)
	myPeer.PeerObject.TunnelHostOrder[tunnelId].PushBack(hashedVersion)





	// Now, create a cryptoobject and add it to the hashmap of the tcpConnection
	// First, generate identifier
	destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)
	newIdentifier := strconv.Itoa(int(tunnelId)) + destinationHostkeyString
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId:tunnelId, PublicKey:publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}

	//log.Println("Map ",myPeer.PeerObject.CryptoSessionMap)
	/*s := fmt.Sprintf("%s", hashedVersion)
	identi := strconv.Itoa(int(tunnelId)) + s

	log.Println("Map ",myPeer.PeerObject.CryptoSessionMap[identi])

	log.Println("Hashed version: ", hashedVersion)*/
	// check, if we received the destinationhostkey of our final destination >> if so, send  the OnionTunnelReady to the CM/UI Module
	log.Println("ReceivedDestinatioHostKey")
	log.Println(destinationHostKey)
	log.Println("Presaved one")
	log.Println(myPeer.PeerObject.TCPConnections[tunnelId].FinalDestination.DestinationHostkey)
	if bytes.Equal(destinationHostKey, myPeer.PeerObject.TCPConnections[tunnelId].FinalDestination.DestinationHostkey) {
		log.Println("Yes, we've connected to our final destination")
		// TODO: Send OnionTunnelReady to CM/UI module.
		// onionTunnelReady := models.OnionTunnelReady{TunnelID: tunnelId, DestinationHostkey: myPeer.PeerObject.TCPConnections[tunnelId].FinalDestinationHostkey}
		// onionTunnelReadyMessage := services.CreateOnionTunnelReady(onionTunnelReady)
	}
}

func handleExchangeKey(messageChannel services.TCPMessageChannel, data []byte, myPeer *services.Peer) {
	log.Println("EXCHANGE KEY")

	if data != nil {
		messageChannel.Message = data
	}

	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])

	sizeDestinationHostkey := binary.BigEndian.Uint16(messageChannel.Message[8:10])
	endOfDestinationKey := 10 + sizeDestinationHostkey

	destinationHostkey := messageChannel.Message[10:endOfDestinationKey]
	newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
	if err != nil {
		log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey, ", err.Error())
	}

	hashedIdentity := services.GenerateIdentityOfKey(newPublicKey)

	pubKey := messageChannel.Message[endOfDestinationKey:]


	// Compute Ephemeral Key
	// First, generate identifier
	destinationHostkeyString := fmt.Sprintf("%s", hashedIdentity)

	var identifier string;

	if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter == nil {
		identifier = strconv.Itoa(int(tunnelID)) + destinationHostkeyString

		// ToDo: Change state of Connection to established.
		log.Println("Test")
		dataMessage := models.DataConstructTunnel{NetworkVersion: "IPv4", DestinationAddress: "192.168.0.10", Port: 4500, DestinationHostkey: destinationHostkey}
		data := services.CreateDataConstructTunnel(dataMessage)

		// Now, just for tests, send a forward to a new peer
		tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateTunnelInstruction(tunnelInstructionMessage)

		myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
	} else {
		identifier = strconv.Itoa(int(tunnelID))
	}


	cryptoObject := myPeer.PeerObject.CryptoSessionMap[identifier]

	if cryptoObject.SessionKey == nil {
		sessionKey := services.ComputeEphemeralKey(cryptoObject.Group, pubKey, cryptoObject.PrivateKey, myPeer.PeerObject.PrivateKey)
		cryptoObject.SessionKey = sessionKey
		log.Println("Created session key.")

		// Send Exchange Key if Left Writer exists
		if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter != nil {
			encryptedKey, err := services.EncryptKeyExchange(newPublicKey, cryptoObject.PublicKey)

			if err != nil {
				log.Println("Error encrypting Key", err.Error())
			}

			keyExchange := models.ExchangeKey{PublicKey: encryptedKey, TunnelID: tunnelID, DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey)}
			keyExchangeMessage := services.CreateExchangeKey(keyExchange)

			myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(keyExchangeMessage)
		}
	} else if cryptoObject.SessionKey != nil && myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter != nil {
		instructionData := services.CreateDataExchangeKey(messageChannel.Message)

		instruction := models.TunnelInstruction{TunnelID: tunnelID, Data: instructionData}
		instructionMessage := services.CreateTunnelInstruction(instruction)

		myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(instructionMessage)
	}
}
