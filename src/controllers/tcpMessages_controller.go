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

func handleTCPMessage(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	messageType := binary.BigEndian.Uint16(messageChannel.Message[2:4])
	log.Println("Messagetype:", messageType)

	switch messageType {
		// ONION TUNNEL DESTROY
		case 563:
			handleOnionTunnelDestroy(binary.BigEndian.Uint32(messageChannel.Message[4:8]), myPeer)
			break

		// ONION TUNNEL DATA
		case 564:
			handleOnionTunnelData(messageChannel, myPeer)
			break

		// CONSTRUCT TUNNEL
		case 567:
			newUDPConnection := handleConstructTunnel(messageChannel, myPeer)
			if newUDPConnection == nil {
				services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error:errors.New("Didn't catched an error!!!")}
				break
			} else {
				myPeer.AppendNewUDPConnection(newUDPConnection)
			}
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
				if _, err := myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(messageChannel.Message); err != nil {
					services.CummunicationChannelError <- services.ChannelError{TunnelId:tunnelID, Error:err}
					return
				}
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
							 services.CummunicationChannelError <- services.ChannelError{TunnelId:tunnelID, Error: errors.New("Error creating tcp writer, error: " + err.Error())}
							 return
						 }
						 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter = newTCPWriter
						 constructMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, OriginHostkey: myPeer.PeerObject.TCPConnections[tunnelID].OriginHostkey, PublicKey: pubKey, TunnelID: tunnelID, DestinationAddress: ipAdd, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort: uint16(myPeer.PeerObject.P2P_Port)}
						 message := services.CreateConstructTunnelMessage(constructMessage)

						 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
						 log.Println("Send Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationIP + " , Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.DestinationPort))
						 break
					 default:
						 services.CummunicationChannelError <- services.ChannelError{TunnelId:tunnelID, Error: errors.New("tcpMessagesController: Message Type not Found")}
						 return
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

		case 566:
			log.Println("Messagetype: Onion Tunnel Traffic Jam TCP")
			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
			data := messageChannel.Message[8:]
			log.Println("Fake data send on tunnel ", tunnelID)
			log.Println(data)
			break


	default:
			return
	}

	return
}

func handleOnionTunnelDestroy(tunnelId uint32, myPeer *services.Peer) {
	log.Println("ONION TUNNEL DESTROY")
	log.Println(tunnelId)

	// check if right writer tcp exists >
	if myPeer.PeerObject.TCPConnections[tunnelId] != nil {

		if myPeer.PeerObject.TCPConnections[tunnelId].RightWriter != nil {
			onionTunnelDestroy := models.OnionTunnelDestroy{TunnelID:tunnelId}
			message := services.CreateOnionTunnelDestroy(onionTunnelDestroy)
			//newMessage := append(message, []byte("\r\n")...)
			//log.Println(newMessage)
			myPeer.PeerObject.TCPConnections[tunnelId].RightWriter.TCPWriter.Write(message)
		}
		// now, delete everything
		delete(myPeer.PeerObject.TCPConnections, tunnelId)
		delete(myPeer.PeerObject.UDPConnections, tunnelId)
		// TODO: Delete Cryptpsessionmap
		delete(myPeer.PeerObject.TunnelHostOrder, tunnelId)
	} else {
		log.Println("Done deleting")
	}
	log.Println(myPeer.PeerObject.TCPConnections[tunnelId])

	return
}


func handleOnionTunnelData(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Println("Messagetype: Onion Tunnel Data")
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
	//data := messageChannel.Message[8:]

	log.Printf("Tunnel ID: %d\n", tunnelID)

	_, err := myPeer.PeerObject.UDPConnections[tunnelID].RightWriter.Write(messageChannel.Message)
	if err != nil {
		services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("handleOnionTunnelData, Error: " + err.Error())}
		return
	}
}


func handleConstructTunnel(messageChannel services.TCPMessageChannel, myPeer *services.Peer) *models.UDPConnection {
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
		services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Error creating tcp writer, error: " + err.Error())}
		return nil
	}

	// Append the new TCPWriter as LeftTCPWriter to the TCP Connection
	myPeer.CreateInitialTCPConnection(tunnelID, destinationHostkey ,newTCPWriter, originHostkey)

	//  Now, create new UDP Connection with this "sender" as left side
	newUDPConnection, err := services.CreateInitialUDPConnectionLeft(ipAdd, int(onionPort), tunnelID)
	if err != nil {
		services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("handleConstructTunnelerror: " + err.Error())}
		return nil
	}

	// Now, create the crypto object and add it with the tunnel id to our peer
	// Create Pre Master Key for session.
	// Compute ephemeral Key
	privateKey, publicKey, group := services.GeneratePreMasterKey()
	myPeer.PeerObject.CryptoSessionMap[strconv.Itoa(int(tunnelID))] = &models.CryptoObject{TunnelId:tunnelID, PrivateKey:privateKey, PublicKey:publicKey, SessionKey:nil, Group:group}
	//encryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

	log.Println(x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey))
	decryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)

	err = saveEphemeralKey(decryptedPubKey, x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), tunnelID, myPeer)
	if err != nil {
		services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Handle Construct Tunnel: Error while saving ephemeral key: " + err.Error())}
		return nil
	}

	originPubKey, err := x509.ParsePKCS1PublicKey(originHostkey)
	if err != nil {
		services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Handle Construct Tunnel: Error while parsing Origin Public Key: " + err.Error())}
		return nil
	}

	// Encrypt Pre-master secret with Origin Public Key
	encryptedPublicKey, err := services.EncryptKeyExchange(originPubKey, publicKey)
	if err != nil {
		services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Handle Construct Tunnel: Error while EncryptKeyExchange: " + err.Error())}
		return nil
	}
	// If everything worked out, send confirmTunnelConstruction back
	confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: x509.MarshalPKCS1PublicKey(myPeer.PeerObject.PublicKey), PublicKey: encryptedPublicKey}
	message := services.CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction)

	// Sent confirm tunnel construction
	_, err = myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
	if err != nil {
		services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Handle Construct Tunnel: Error while writing: " + err.Error())}
		return nil
	}
	log.Println("Sent Confirm Tunnel Construction to " + myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationIP + ", Port: " + strconv.Itoa(myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.DestinationPort))

	return newUDPConnection
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


	// Start trafffic jamming
	go func() {
		InitiateTrafficJamming(tunnelID, myPeer)
	}()

	// Check whether left writer exists. If yes, create Confirm Tunnel Instruction with ConfirmTunnelConstruction data.
	if myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter != nil {
		// Now, create a udp righter to the right side (where we forwarded our tunnel construction to
		newRightUDPWriter, err := services.CreateUDPWriter(services.GetIPOutOfAddr(messageChannel.Host), int(onionPort))
		if err != nil {
			services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Error creating UDP Writer right >> " + err.Error())}
			return
		}
		myPeer.PeerObject.UDPConnections[tunnelID].RightWriter = newRightUDPWriter

		dataMessage := models.DataConfirmTunnelConstruction{DestinationHostkey: destinationHostkey, PublicKey: pubKey}
		data := services.CreateDataConfirmTunnelConstruction(dataMessage)
		confirmTunnelInstruction := models.ConfirmTunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateConfirmTunnelInstruction(confirmTunnelInstruction)
		log.Println("Final Destination exists??????")
		log.Println(myPeer.PeerObject.TCPConnections[tunnelID])

		_, err = myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
		if err != nil {
			services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Error while writing " + err.Error())}
			return
		}
	} else {
		log.Println("Final Destination > PEER 0")
		log.Println(myPeer.PeerObject.TCPConnections[tunnelID].FinalDestination)

		// we reached peer 0 > create hashed hostkey of confirmation sender and add it to tunnelHostOrder
		newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
		if err != nil {
			services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Couldn't convert []byte destinationHostKey to rsa Publickey,  " + err.Error())}
			return
		}
		hashedVersion := services.GenerateIdentityOfKey(newPublicKey)
		myPeer.PeerObject.TunnelHostOrder[tunnelID].PushBack(hashedVersion)

		decryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)
		err = saveEphemeralKey(decryptedPubKey, destinationHostkey, tunnelID, myPeer)
		if err != nil{
			services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("Handle Confirm Tunnel Construction: Error while saving ephemeral key " + err.Error())}
			return
		}

		// Now, create the UDP Writer Right for the UDP to the next right hop and assign it
		newUDPConnection, err := services.CreateInitialUDPConnectionRight(services.GetIPOutOfAddr(messageChannel.Host), int(onionPort), tunnelID)
		if err != nil {
			services.CummunicationChannelError <- services.ChannelError{TunnelId:binary.BigEndian.Uint32(messageChannel.Message[4:8]), Error: errors.New("handleConfirmTunnelConstruction while creating udp connection:  " + err.Error())}
			return

		}
		myPeer.AppendNewUDPConnection(newUDPConnection)

		go func() {services.CommunicationChannelTCPConfirm <- services.ConfirmMessageChannel{TunnelId:tunnelID} } ()
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
		services.CummunicationChannelError <- services.ChannelError{TunnelId:tunnelId, Error: errors.New("Couldn't convert []byte destinationHostKey to rsa Publickey " + err.Error())}
		return
	}

	// Now, create hashed version of the destination hostkey and add it to the TunnelHostOrder
	hashedVersion := services.GenerateIdentityOfKey(newPublicKey)
	myPeer.PeerObject.TunnelHostOrder[tunnelId].PushBack(hashedVersion)

	// Decrypt Pre master secret
	decryptedPubKey, err := services.DecryptKeyExchange(myPeer.PeerObject.PrivateKey, pubKey)
	if err != nil {
		services.CummunicationChannelError <- services.ChannelError{TunnelId:tunnelId, Error: errors.New("Couldn't run DecryptKeyExchange(), " + err.Error())}
		return
	}

	// Save Ephemeral Key
	err = saveEphemeralKey(decryptedPubKey, destinationHostkey, tunnelId, myPeer)
	if err != nil {
		services.CummunicationChannelError <- services.ChannelError{TunnelId:tunnelId, Error: errors.New("Handle Confirm Tunnel Instruction Construction: Error while saving Ephemeral Key " + err.Error())}
		return
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