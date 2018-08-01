package main

import (
	"services"
	"log"
	"os"
	"controllers"
	"time"
	"os/signal"
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

	// Now, start Controlling
	controllers.StartPeerController(newPeer)
	controllers.StartUDPController(newPeer)

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	log.Println("Peer status: offline \n\n\n")
	os.Exit(0)

}