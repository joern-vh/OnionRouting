package controllers

import (
	"log"
	"services"
)

// Handles TCP Messages retieved from the CommunicationChannelTCPMessages
func StartTCPController(myPeer *services.Peer) {
	for msg := range services.CommunicationChannelTCPMessages {
		handleTCPMessage(msg, myPeer)
	}
}

func  handleTCPMessage(message []byte, myPeer *services.Peer) error {
	log.Println(message)
	return nil
}