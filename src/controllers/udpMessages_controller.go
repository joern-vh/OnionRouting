package controllers

import "services"

func StartUDPController(myPeer *services.Peer) {
	for msg := range services.CommunicationChannelUDPMessages {
		handleUDPMessage(msg, myPeer)
	}
}

func handleUDPMessage(message []byte, myPeer *services.Peer) {
	
}