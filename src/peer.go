package main

import (
	"log"
	"services"
)

func main() {
	log.Println("Start creating peer")

	// Start creating a peer which can just start listening
	newPeer, err := services.CreateNewPeer(3000, "127.0.0.1")
	if err != nil {
		log.Println("Couldn't create new Peer, error:", err)
	}

	// Now start listening
	if err := newPeer.StartTCPListening(); err != nil {
		log.Println("Problem listening for new messages: ", err.Error())
		log.Println("Stopped peer due to error")
		return
	}
}