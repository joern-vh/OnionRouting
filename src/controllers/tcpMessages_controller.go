package controllers

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"services"

	"models"
	"fmt"
	"bytes"
	"container/list"
	"crypto/x509"
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

				 switch command {
				 // Construct tunnel
				 case 567:
					 ipAdd := net.IP(data[2:6]).String()
					 //tcpPort :=
					 // now, create new TCP RightWriter for the right side
					 newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd, 4200)
					 if err != nil {
						 return errors.New("Error creating tcp writer, error: " + err.Error())
					 }
					 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter = newTCPWriter

					 constructMessage := models.ConstructTunnel{NetworkVersion: "IPv4", DestinationHostkey: []byte("KEY"), DestinationAddress: ipAdd, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort:uint16(myPeer.PeerObject.P2P_Port), TunnelID: tunnelID}
					 message := services.CreateConstructTunnelMessage(constructMessage)

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
			log.Println(messageChannel.Message[6:9])

			sizeDestinationHostkey := binary.BigEndian.Uint16(messageChannel.Message[4:6])

			destinationhostkey := messageChannel.Message[6:6+sizeDestinationHostkey]

			pubKey := messageChannel.Message[6+sizeDestinationHostkey:]

			log.Println("Destination Hostkey: ", string(destinationhostkey))
			log.Println("PubKey: ", pubKey)

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
	constructTunnelMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, TunnelID: newTunnelID, DestinationAddress: destinationAddress, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort: uint16(myPeer.PeerObject.P2P_Port)}
	message := services.CreateConstructTunnelMessage(constructTunnelMessage)

	newTCPWriter, err := myPeer.CreateTCPWriter(destinationAddress, 3000)
	if err != nil {
		log.Println("Error creating tcp writer, error: " + err.Error())
	}

	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID] = &models.TCPConnection{constructTunnelMessage.TunnelID, nil, newTCPWriter, list.New()}
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
	myPeer.CreateInitialTCPConnection(tunnelID, newTCPWriter)

	//  Now, create new UDP Connection with this "sender" as left side
	newUDPConnection, err := services.CreateInitialUDPConnection(ipAdd, int(onionPort), tunnelID, networkVersionString)
	if err != nil {
		return nil, errors.New("handleConstructTunnel: " + err.Error())
	}

	// If everything worked out, send confirmTunnelConstruction back
	confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: destinationHostkey}
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
		data := []byte("TEST")
		confirmTunnelInstruction := models.ConfirmTunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateConfirmTunnelInstruction(confirmTunnelInstruction)

		myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)
	} else {

		// Add hostkey to the list of available host, but first, convert it
		newPublicKey, err := x509.ParsePKCS1PublicKey(destinationHostkey)
		if err != nil {
			log.Println("Couldn't convert []byte destinationHostKey to rsa Publickey, ", err.Error())
		}
		myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder.PushFront(services.GenerateIdentityOfKey(newPublicKey))

		// Iterate through list and print its contents.
		for e := myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder.Front(); e != nil; e = e.Next() {
			fmt.Println("List value ", e.Value)
		}

		// Convert messageType to Byte array
		messageTypeBuf := new(bytes.Buffer)
		binary.Write(messageTypeBuf, binary.BigEndian, uint16(567))
		data := messageTypeBuf.Bytes()

		log.Println(myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder)
		ip := net.IP(data[2:6]).String()
		log.Println(ip)
		// geht the element wit the right ip and set its value to confirm: true
		/*for i := range myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder {
			if myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder[i].IpAddress == ip {
				myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder[i].Confirmed = true
			}
		}*/

		ipAddr := net.ParseIP("192.168.0.15")
		log.Println(myPeer.PeerObject.TCPConnections[tunnelID].ConnectionOrder)

		data = append(data, ipAddr.To4()...)

		portBuf := new(bytes.Buffer)
		binary.Write(portBuf, binary.BigEndian, uint16(4200))
		data = append(data, portBuf.Bytes()...)


		// Now, just for tests, send a forward to a new peer
		tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelID, Data: data}
		message := services.CreateTunnelInstruction(tunnelInstructionMessage)

		myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)


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
}
