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
			log.Println("\n New message from " + msg.Host)
			handleTCPMessage(msg, myPeer)
		}
	}()

}

func handleTCPMessage(messageChannel services.TCPMessageChannel, myPeer *services.Peer) error {
	messageType := binary.BigEndian.Uint16(messageChannel.Message[2:4])
	log.Println("Messagetype: ", messageType)

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
			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
			log.Println("Forwarding on Tunnel id", tunnelID)
			log.Println(myPeer.PeerObject.TCPConnections[tunnelID])
			 // First, check if there is a right peer set for this tunnel id, if not decrypt data and exeute, if yes, simply forward it
			 // simply forward
			 if myPeer.PeerObject.TCPConnections[tunnelID].RightWriter != nil {
			 	log.Println("Right Writer exists, simply forward data ")
				 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(messageChannel.Message)
			 } else {
			 	// no right writer exists, handle data now to send the command
			 	// first, determine the command out of data and execute the function
			 	data := messageChannel.Message[8:]
			 	command := binary.BigEndian.Uint16(data[0:2])
			 	log.Println("Command to forward: ", command)
			 	log.Println(data)

				 switch command {
				 // Construct tunnel
				 case 567:
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
						log.Println(ipAdd)

					 //tcpPort :=
					 // now, create new TCP RightWriter for the right side
					 newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd, int(port))
					 if err != nil {
						 return errors.New("Error creating tcp writer, error: " + err.Error())
					 }
					 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter = newTCPWriter
					 constructMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, DestinationAddress: ipAdd, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort:uint16(myPeer.PeerObject.P2P_Port), TunnelID: tunnelID}
					 message := services.CreateConstructTunnelMessage(constructMessage)

					 log.Println("Yes, i was here")
					 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
					 break


				 default:
					 return errors.New("tcpMessagesController: Message Type not Found")
				 }
			 }

			break

		// CONFIRM TUNNEL INSTRUCTION
		case 570:
			log.Println("CONFIRM TUNNEL INSTRUCTION")
			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
			data := messageChannel.Message[8:]
			log.Println("TunnelID: ", tunnelID)
			log.Println("Data: ", data)

			command := binary.BigEndian.Uint16(data[0:2])
			log.Println("Command: ", command)
			switch command {
			case 568:
				log.Println("Got a confirmation for a tunnel construction")
				handleConfirmTunnelInnstructionConstruction(tunnelID, data[2:], myPeer)
				break
			}
			break

		// EXCHANGE KEY
		case 571:
			log.Println("EXCHANGE KEY")

			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])

			sizeDestinationHostkey := binary.BigEndian.Uint16(messageChannel.Message[8:10])
			endOfDestinationKey := 10 + sizeDestinationHostkey

			//destinationHostkey := messageChannel.Message[10:endOfDestinationKey]
			/*newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
			if err != nil {
				log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey, ", err.Error())
			}*/

			//hashedIdentity := services.GenerateIdentityOfKey(newPublicKey)

			pubKey := messageChannel.Message[endOfDestinationKey:]


			// Compute Ephemeral Key
			// First, generate identifier
			//destinationHostkeyString := fmt.Sprintf("%s", hashedIdentity)
			identifier := strconv.Itoa(int(tunnelID))
			cryptoObject := myPeer.PeerObject.CryptoSessionMap[identifier]

			sessionKey := services.ComputeEphemeralKey(cryptoObject.Group, pubKey, cryptoObject.PrivateKey)

			if cryptoObject.SessionKey == nil {
				cryptoObject.SessionKey = sessionKey
				log.Println("Created session key.")
			}

			log.Println(cryptoObject)

			/*dataMessage := models.DataConstructTunnel{NetworkVersion: "IPv4", DestinationAddress: "192.168.0.15", Port: 4200, DestinationHostkey: destinationHostkey}
			data := services.CreateDataConstructTunnel(dataMessage)

			// Now, just for tests, send a forward to a new peer
			tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelID, Data: data}
			message := services.CreateTunnelInstruction(tunnelInstructionMessage)

			myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)*/

			break

		default:
			return errors.New("tcpMessagesController: Message Type not Found")
	}

	return nil
}

func handleOnionTunnelBuild(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	var networkVersionString string
	var destinationAddress string
	var destinationHostkey []byte

	networkVersion := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	onionPort := binary.BigEndian.Uint16(messageChannel.Message[6:8])

	fmt.Println("SIZE OF ONION TUNNEL BUILD: ", len(messageChannel.Message))

	if networkVersion == 0 {
		networkVersionString = "IPv4"
		destinationAddress = net.IP(messageChannel.Message[8:12]).String()
		destinationHostkey = messageChannel.Message[12:]
	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
		destinationAddress = net.IP(messageChannel.Message[8:24]).String()
		destinationHostkey = messageChannel.Message[24:]
	}

	log.Println(destinationHostkey)

	//Construct Tunnel Message
	newTunnelID := services.CreateTunnelID()
	log.Println("NewTunnelID: ", newTunnelID)
	log.Println("IP: ", destinationAddress)
	constructTunnelMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, TunnelID: newTunnelID, DestinationAddress: destinationAddress, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort: uint16(myPeer.PeerObject.P2P_Port)}
	message := services.CreateConstructTunnelMessage(constructTunnelMessage)

	newTCPWriter, err := myPeer.CreateTCPWriter(destinationAddress, 3000)
	if err != nil {
		log.Println("Error creating tcp writer, error: " + err.Error())
	}

	// Generate hash of the final destination hoskey
	newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
	if err != nil {
		log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey, ", err.Error())
	}
	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID] = &models.TCPConnection{constructTunnelMessage.TunnelID, nil, newTCPWriter, list.New(), services.GenerateIdentityOfKey(newPublicKey)}
	log.Println(myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].FinalDestinationHostkey)
	// Now just add the right connection to the map with status pending
	//myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].ConnectionOrder = append(myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].ConnectionOrder, models.ConnnectionOrderObject{TunnelId:constructTunnelMessage.TunnelID, IpAddress:destinationAddress, IpPort:4200, Confirmed:false})
	n, _ := myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].RightWriter.TCPWriter.Write(message)

	log.Println("Size: ", n)

	log.Printf("Network Version: %s\n", networkVersionString)
	log.Printf("Onion Port: %d\n", onionPort)
	log.Printf("Destination Address: %s\n", destinationAddress)
	log.Printf("Destination Hostkey: %x\n", destinationHostkey)
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
	var networkVersionString string
	//var destinationAddress string
	var destinationHostkey []byte

	networkVersion := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	onionPort := binary.BigEndian.Uint16(messageChannel.Message[6:8])
	tcpPort := binary.BigEndian.Uint16(messageChannel.Message[8:10])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[10:14])

	if networkVersion == 0 {
		networkVersionString = "IPv4"
		//destinationAddress = net.IP(messageChannel.Message[14:18]).String()
		destinationHostkey = messageChannel.Message[18:]
	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
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
	newUDPConnection, err := services.CreateInitialUDPConnection(ipAdd, int(onionPort), tunnelID, networkVersionString)
	if err != nil {
		return nil, errors.New("handleConstructTunnel: " + err.Error())
	}

	// Now, create the crypto object and add it with the tunnel id to our peer
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[strconv.Itoa(int(tunnelID))] = &models.CryptoObject{TunnelId:tunnelID, PrivateKey:privateKey, PublicKey:publicKey, SessionKey:nil, Group:group}
	log.Println(myPeer.PeerObject.CryptoSessionMap[strconv.Itoa(int(tunnelID))])
	// If everything worked out, send confirmTunnelConstruction back
	confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey)}
	message := services.CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction)

	myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)

	return newUDPConnection, nil
}

func handleConfirmTunnelConstruction(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Println("CONFIRM RECEIVED")

	onionPort := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[6:10])
	destinationHostkey := messageChannel.Message[10:]

	if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter != nil {
		// Forward to left a confirmTunnelInstruction
		dataMessage := models.DataConfirmTunnelConstruction{DestinationHostkey: destinationHostkey}
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

		// Iterate through list and print its contents.
		for e := myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder.Front(); e != nil; e = e.Next() {
			fmt.Println("List value ", e.Value)
		}

		log.Println("Map ",myPeer.PeerObject.CryptoSessionMap)

		// Now, create a cryptoobject and add it to the hashmap of the tcpConnection
		// First, generate identifier
		destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)
		newIdentifier := strconv.Itoa(int(tunnelID)) + destinationHostkeyString
		privateKey, publicKey, group := services.GeneratePreMasterKey()
		myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId:tunnelID, PublicKey:publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}

		log.Println("Hashed version: ", hashedVersion)
		// We received a confirmation from out final destination, send ready message
		if bytes.Equal(hashedVersion, myPeer.PeerObject.TCPConnections[tunnelID].FinalDestinationHostkey) {
			log.Println("Yes, we've connected to our final destination")

			onionTunnelReady := models.OnionTunnelReady{TunnelID: tunnelID, DestinationHostkey: myPeer.PeerObject.TCPConnections[tunnelID].FinalDestinationHostkey}
			onionTunnelReadyMessage := services.CreateOnionTunnelReady(onionTunnelReady)

			// TODO: Send OnionTunnelReady to CM/UI module.
			log.Println(onionTunnelReadyMessage)
		}

		// Convert messageType to Byte array
		/*messageTypeBuf := new(bytes.Buffer)
		binary.Write(messageTypeBuf, binary.BigEndian, uint16(567))
		data := messageTypeBuf.Bytes()
		*/
		log.Println(myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder)
		//ip := net.IP(data[2:6]).String()
		//log.Println(ip)
		// geht the element wit the right ip and set its value to confirm: true
		/*for i := range myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder {
			if myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder[i].IpAddress == ip {
				myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder[i].Confirmed = true
			}
		}*/

		//ipAddr := net.ParseIP("192.168.0.15")
		log.Println(myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder)

		/*data = append(data, ipAddr.To4()...)

		portBuf := new(bytes.Buffer)
		binary.Write(portBuf, binary.BigEndian, uint16(4200))
		data = append(data, portBuf.Bytes()...)*/


		// KEY EXCHANGE TESTING

		log.Println("TESTING: ", services.GenerateIdentityOfKey(myPeer.PeerObject.PublicKey))
		keyExchange := models.ExchangeKey{PublicKey: myPeer.PeerObject.CryptoSessionMap[newIdentifier].PublicKey, TunnelID: tunnelID, DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey)}
		keyExchangeMessage := services.CreateExchangeKey(keyExchange)

		myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(keyExchangeMessage)


		// OLD TESTING

		/*dataMessage := models.DataConstructTunnel{NetworkVersion: "IPv4", DestinationAddress: "192.168.0.15", Port: 4200, DestinationHostkey: destinationHostkey}
		data := services.CreateDataConstructTunnel(dataMessage)

		// Now, just for tests, send a forward to a new peer
		tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateTunnelInstruction(tunnelInstructionMessage)

		myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)*/


		log.Printf("Onion Port: %d\n", onionPort)
		log.Printf("Tunnel ID: %d\n", tunnelID)
	}
}

func handleConfirmTunnelInnstructionConstruction(tunnelId uint32, destinationHostKey []byte, myPeer *services.Peer) {
	// State now: just add hostkey
	// Add hostkey to the list of available host, but first, convert it
	newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostKey)
	if err != nil {
		log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey")
	}
	myPeer.PeerObject.TCPConnections[tunnelId].ConnectionOrder.PushBack(services.GenerateIdentityOfKey(newPublicKey))

	// Iterate through list and print its contents.
	for e := myPeer.PeerObject.TCPConnections[tunnelId].ConnectionOrder.Front(); e != nil; e = e.Next() {
		fmt.Println("List value ", e.Value)
	}

	hashedVersion := services.GenerateIdentityOfKey(newPublicKey)

	// Now, create a cryptoobject and add it to the hashmap of the tcpConnection
	// First, generate identifier
	destinationHostkeyString := fmt.Sprintf("%s", hashedVersion)
	newIdentifier := strconv.Itoa(int(tunnelId)) + destinationHostkeyString
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[newIdentifier] = &models.CryptoObject{TunnelId:tunnelId, PublicKey:publicKey, PrivateKey:privateKey,SessionKey:nil, Group:group}

	//log.Println("Map ",myPeer.PeerObject.CryptoSessionMap)
	s := fmt.Sprintf("%s", hashedVersion)
	identi := strconv.Itoa(int(tunnelId)) + s

	log.Println("Map ",myPeer.PeerObject.CryptoSessionMap[identi])

	log.Println("Hashed version: ", hashedVersion)
	// We received a confirmation from out final destination, send ready message
	if bytes.Equal(hashedVersion, myPeer.PeerObject.TCPConnections[tunnelId].FinalDestinationHostkey) {
		log.Println("Yes, we've connected to our final destination")

		onionTunnelReady := models.OnionTunnelReady{TunnelID: tunnelId, DestinationHostkey: myPeer.PeerObject.TCPConnections[tunnelId].FinalDestinationHostkey}
		onionTunnelReadyMessage := services.CreateOnionTunnelReady(onionTunnelReady)

		// TODO: Send OnionTunnelReady to CM/UI module.
		log.Println(onionTunnelReadyMessage)
	}
}
