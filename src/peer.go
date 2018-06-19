package main

import (
	"services"
	"log"
	"os"
	"controllers"
	"time"
)

func main() {
	// First, create new config object which is used to create this peer later
	config, err := services.NewConfigObject()
	if err != nil {
		log.Println("Error creating config object: " + err.Error())
		os.Exit(1)
	}

	// Now, create peer based on config object
	newPeer, err := services.CreateNewPeer(config)
	if err != nil {
		log.Println("Error creating peer: " + err.Error())
		os.Exit(1)
	}

	// Now, start TCP-Listening and UDP-Listening
	newPeer.StartTCPListening()
	time.Sleep(time.Second * 1)
	newPeer.StartUDPListening()

	// Now, start TCP-Controlling
	controllers.StartTCPController(newPeer)
}