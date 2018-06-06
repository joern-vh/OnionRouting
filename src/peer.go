package main

import (
	"services"
	"log"
	"flag"
	"os"
)

func main() {
	// Define and parse command line flags
	possibleFlags := flag.String("c", "config.ini", "Pleas write the path to your config.ini file ")
	flag.Parse()

	// Check if path was empty
	if *possibleFlags == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	
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