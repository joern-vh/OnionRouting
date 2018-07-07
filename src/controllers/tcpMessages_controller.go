package controllers

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"services"

	"models"
)

func StartTCPController(myPeer *services.Peer) {
	log.Println("StartTCPController: Started TCP Controller")
	go func() {
		for msg := range services.CommunicationChannelTCPMessages {
			log.Println("Yes, i got a new messagebitch")
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
			log.Println("handleConstructTunnel")
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
			log.Println("Forwarding!!!")

			tunnelID := binary.BigEndian.Uint32(messageChannel.Message[4:8])
			log.Println("Tunnel id", tunnelID)
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
			 	log.Println(data)

			 	command := binary.BigEndian.Uint16(data[0:2])
			 	log.Println("I got this command: ", command)

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

					 constructMessage := models.ConstructTunnel{NetworkVersion: "IPv4", DestinationHostkey: []byte("KEY"), DestinationAddress: ipAdd, Port: uint16(myPeer.PeerObject.UDPPort)}
					 message := services.CreateConstructTunnelMessage(constructMessage)

					 myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)
					 break


				 default:
					 return errors.New("tcpMessagesController: Message Type not Found")
				 }
			 }

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
	constructTunnelMessage := models.ConstructTunnel{NetworkVersion: networkVersionString, DestinationHostkey: destinationHostkey, TunnelID: newTunnelID, DestinationAddress: destinationAddress, Port: uint16(myPeer.PeerObject.UDPPort)}
	message := services.CreateConstructTunnelMessage(constructTunnelMessage)

	newTCPWriter, err := myPeer.CreateTCPWriter(destinationAddress, 3000)
	if err != nil {
		log.Println("Error creating tcp writer, error: " + err.Error())
	}

	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID] = &models.TCPConnection{constructTunnelMessage.TunnelID, nil, newTCPWriter}
	myPeer.PeerObject.TCPConnections[constructTunnelMessage.TunnelID].RightWriter.TCPWriter.Write(message)

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
	endOfMessage := len(messageChannel.Message)
	log.Println(endOfMessage)

	var networkVersionString string
	var destinationAddress string
	var destinationHostkey []byte

	networkVersion := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	onionPort := binary.BigEndian.Uint16(messageChannel.Message[6:8])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[8:12])

	if networkVersion == 0 {
		networkVersionString = "IPv4"
		destinationAddress = net.IP(messageChannel.Message[12:16]).String()
		destinationHostkey = messageChannel.Message[16:endOfMessage]
		log.Printf("Destination Hostkey: %s\n", destinationHostkey)
	} else if networkVersion == 1 {
		networkVersionString = "IPv6"
		destinationAddress = net.IP(messageChannel.Message[12:28]).String()
		destinationHostkey = messageChannel.Message[28:endOfMessage]
	}

	log.Printf("Network Version: %s\n", networkVersionString)
	log.Printf("Onion Port: %d\n", onionPort)
	log.Printf("Tunnel ID: %d\n", tunnelID)
	log.Printf("Destination Address: %s\n", destinationAddress)
	log.Printf("Destination Hostkey: %s\n", destinationHostkey)

	// First, get ip address of sender
	ipAdd := services.GetIPOutOfAddr(messageChannel.Host)

	// Then, create the TCPWriter left
	newTCPWriter, err := myPeer.CreateTCPWriter(ipAdd, 3000)
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
	endOfMessage := len(messageChannel.Message) -3

	onionPort := binary.BigEndian.Uint16(messageChannel.Message[4:6])
	tunnelID := binary.BigEndian.Uint32(messageChannel.Message[6:10])
	destinationHostkey := messageChannel.Message[10:endOfMessage]

	// Now, just for tests, send a forward to a new peer
	tunnelInstructionMessage := models.TunnelInstruction{TunnelID: tunnelID, Data: []byte("Some Data")}
	message := services.CreateTunnelInstruction(tunnelInstructionMessage)

	log.Println(myPeer.PeerObject.TCPConnections[tunnelID].RightWriter)
	myPeer.PeerObject.TCPConnections[tunnelID].RightWriter.TCPWriter.Write(message)

	log.Printf("Onion Port: %d\n", onionPort)
	log.Printf("Tunnel ID: %d\n", tunnelID)
	log.Printf("Destination Hostkey: %s\n", destinationHostkey)
}
