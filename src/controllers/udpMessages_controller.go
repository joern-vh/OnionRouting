package controllers

import (
	"services"
	"log"
	"encoding/binary"
	"errors"
)

func StartUDPController(myPeer *services.Peer) {
	log.Println("StartUDPController: Started UDP Controller")

	go func() {
		for msg := range services.CommunicationChannelUDPMessages {
			handleUDPMessage(msg, myPeer)
		}
	}()
}

func handleUDPMessage(message []byte, myPeer *services.Peer) {
	log.Println("Messagetype: handleUDPMessage")
	tunnelID := binary.BigEndian.Uint32(message[4:8])

	// now, check if there is a right udp writer for this connection
	if (myPeer.PeerObject.UDPConnections[tunnelID].RightWriter != nil) {
		// if so, forward data
		_, err := myPeer.PeerObject.UDPConnections[tunnelID].RightWriter.Write(message)
		if err != nil {
			services.CummunicationChannelError <- services.ChannelError{TunnelId: tunnelID, Error: errors.New("handleUDPMessage, Error while writing: " + err.Error())}
			return
		}
	} else {
		// reached final destination
		log.Println("UDP: Final destiantion reached, message: ")
		log.Println(message[8:])
	}
}