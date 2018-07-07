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
)

func StartTCPController(myPeer *services.Peer) {
	log.Println("StartTCPController: Started TCP Controller")
	go func() {
		for msg := range services.CommunicationChannelTCPMessages {
			handleTCPMessage(msg, myPeer)
		}
	}()

}

func handleTCPMessage(messageChannel services.TCPMessageChannel, myPeer *services.Peer) error {
	messageType := binary.BigEndian.Uint16(messageChannel.Message[2:4])

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
			log.Println("Ja, hier war ich!")
			endOfMessage := len(messageChannel.Message) -3
			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:12])

			 // First, check if there is a right peer set for this tunnel id, if not decrypt data and exeute, if yes, simply forward it
			 // simply forward
			 if myPeer.PeerObject.TCPConnections[tunnelID].RightWriter != nil {
			 	log.Println("Right Writer exists, simply forward data ")
				 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(messageChannel.Message)
			 } else {
			 	// no right writer exists, handle data now to send the command
			 	// first, determine the command out of data and execute the function
			 	data := messageChannel.Message[4:endOfMessage]

			 	log.Println(string(data))

			 	/*command := binary.BigEndian.Uint16(data[0:2])

				 switch command {
				 case 567:*/
					// constructMessage := models.ConstructTunnel{NetworkVersion: "IPv4", DestinationHostkey: []byte("KEY"), DestinationAddress: "192.168.0.15", Port: 61637}
					 //message := services.CreateConstructTunnelMessage(constructMessage)
					 // now, create new TCP RightWriter for the right side
					 ipAdd := "192.168.0.10"
					 //newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd)
					 conn, err := net.Dial("tcp", ipAdd + ":4200")
					 if err != nil {
					 	return errors.New("createTCPWriter: Error while dialing to destination, error: " + err.Error())
				 	}

				 	newTCPWriter := &models.TCPWriter{ipAdd, 4200, conn}

					 if err != nil {
						 return errors.New("Error creating tcp writer, error: " + err.Error())
					 }
					 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter = newTCPWriter
					 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write([]byte("Hey bitches !!:D "))
					 //myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
				 	/*break

				 default:
					 return errors.New("tcpMessagesController: Message Type not Found")
				 }*/
			 }

			break

		// CONFIRM TUNNEL INSTRUCTION
		case 570:
			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:12])
			data := messageChannel.Message[4:]

			log.Println("TunnelID: ", tunnelID)
			log.Println("Data: ", data)
			break

		// EXCHANGE KEY
		case 571:
			pubKey := messageChannel.Message[4:]

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
	constructTunnelMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: []byte("KEY"), TunnelID: newTunnelID, DestinationAddress: destinationAddress, OnionPort: uint16(myPeer.PeerObject.UDPPort), TCPPort: uint16(myPeer.PeerObject.P2P_Port)}
	message := services.CreateConstructTunnelMessage(constructTunnelMessage)
	log.Println(message)

	newTCPWriter, err := myPeer.CreateTCPWriter(destinationAddress)
	if err != nil {
		log.Println("Error creating tcp writer, error: " + err.Error())
	}

	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID] = &models.TCPConnection{constructTunnelMessage.TunnelID, nil, newTCPWriter}
	n, _ := myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].RightWriter.TCPWriter.Write(message)

	log.Println("Size: ", n)

	log.Printf("Network Version: %s\n", networkVersionString)
	log.Printf("Onion Port: %d\n", onionPort)
	log.Printf("Destination Address: %s\n", destinationAddress)
	log.Printf("Destination Hostkey: %s\n", destinationHostkey)
}

func handleOnionTunnelDestroy(messageChannel services.TCPMessageChannel) {
	log.Println("ONION TUNNEL DESTROY received")
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
	log.Printf("Tunnel ID: %s\n", tunnelID)
}

func handleConstructTunnel(messageChannel services.TCPMessageChannel, myPeer *services.Peer) (*models.UDPConnection, error) {
	endOfMessage := len(messageChannel.Message) -3

	var networkVersionString string
	var destinationAddress string
	var destinationHostkey []byte

	networkVersion := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	onionPort := binary.BigEndian.Uint16(messageChannel.Message[6:8])
	tcpPort := binary.BigEndian.Uint16(messageChannel.Message[8:10])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[10:14])

	if networkVersion == 0 {
		networkVersionString = "IPv4"
		destinationAddress = net.IP(messageChannel.Message[14:18]).String()
		destinationHostkey = messageChannel.Message[18:endOfMessage]
	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
		destinationAddress = net.IP(messageChannel.Message[14:30]).String()
		destinationHostkey = messageChannel.Message[30:endOfMessage]
	}

	log.Printf("Network Version: %s\n", networkVersionString)
	log.Printf("Onion Port: %d\n", onionPort)
	log.Printf("TCP Port: %d\n", tcpPort)
	log.Printf("Tunnel ID: %d\n", tunnelID)
	log.Printf("Destination Address: %s\n", destinationAddress)
	log.Printf("Destination Hostkey: %s\n", destinationHostkey)

	// First, get ip address of sender
	ipAdd := services.GetIPOutOfAddr(messageChannel.Host)

	// Then, create the TCPWriter left
	newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd)
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
	confirmTunnelConstruction := models.ConfirmTunnelConstruction{TunnelID: tunnelID, Port: uint16(myPeer.PeerObject.UDPPort), DestinationHostkey: []byte("Key")}
	message := services.CreateConfirmTunnelCronstructionMessage(confirmTunnelConstruction)

	myPeer.PeerObject.TCPConnections[tunnelID].LeftWriter.TCPWriter.Write(message)

	return newUDPConnection, nil
}

func handleConfirmTunnelConstruction(messageChannel services.TCPMessageChannel, myPeer *services.Peer) {
	log.Println("CONFIRM RECEIVED")
	endOfMessage := len(messageChannel.Message) - 3

	onionPort := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[6:10])
	destinationHostkey := messageChannel.Message[10:endOfMessage]

	// Convert messageType to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, uint16(567))
	data := messageTypeBuf.Bytes()


	ipAddr := net.ParseIP("192.168.0.15")
	data = append(data, ipAddr.To4()...)

	portBuf := new(bytes.Buffer)
	binary.Write(portBuf, binary.BigEndian, uint16(4200))
	data = append(data, portBuf.Bytes()...)

	log.Println("DATA: ", data)

	// Now, just for tests, send a forward to a new peer
	tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelID, Data: data}
	message := services.CreateTunnelInstruction(tunnelInstructionMessage)

	n,_ := myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)

	log.Println(message)
	log.Println("Size: ", n)

	log.Printf("Onion Port: %d\n", onionPort)
	log.Printf("Tunnel ID: %d\n", tunnelID)
	log.Printf("Destination Hostkey: %s\n", destinationHostkey)
}
