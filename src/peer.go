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

	// Create new config file based on read in path
	config, err := services.NewConfigObject(*possibleFlags)
	if err != nil {
		log.Print("Couldn't parse config file, please check it")
		os.Exit(1)
	}

	// Start creating a peer which can just start listening
	newPeer, err := services.CreateNewPeer(config)
	if err != nil {
		log.Println("Couldn't create new Peer, error:", err)
	}

	// Now start listening
	if err := newPeer.StartTCPListening(); err != nil {
		log.Println("Problem listening for new messages: ", err.Error())
		log.Println("Stopped peer due to error")
		return
	}

	// TODO: Write new parameter option to send something
	/*
	buildMessage := models.OnionTunnelBuild{OnionTunnelBuild: uint16(560), NetworkVersion: "IPv4", Port: uint16(4200), DestinationAddress: "", DestinationHostkey: "KEY"}
	onionTunnelBuild := services.CreateOnionTunnelBuild(buildMessage)
	fmt.Printf("Message: %x\n", onionTunnelBuild)
	sendMessage("192.168.0.10", 3000, onionTunnelBuild)
	 */
}