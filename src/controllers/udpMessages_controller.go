package controllers

import (
	"services"
	"log"
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
	log.Println("handleUDPMessage: ", message)
}