package main

import (
	"services"
	"log"
	"os"
	"time"
	"models"
)

func main() {

	// Get all start-flags
	configFileFlag, sendMessage, sendMessageHost, sendMessagePort, err := services.ParseStartFlags()
	if err != nil {
		log.Println("Error parsing all start-flags: " + err.Error())
		os.Exit(1)
	}

	// Create new config file based on read in path
	config, err := services.NewConfigObject(configFileFlag)
	if err != nil {
		log.Print("Couldn't parse config file, please check it")
		os.Exit(1)
	}

	// Start creating a peer which can just start listening
	newPeer, err := services.CreateNewPeer(config)
	if err != nil {
		log.Println("Couldn't create new Peer, error:", err)
	}

	// This need to run concurrent, so that we can listen and write at the same time >> Use channel for synchronisation, use gofunc to run concurrent
	communicationChannel := make(chan error)
	go func() {
		// Now start TCP listening
		if err := newPeer.StartTCPListening(); err != nil {
			log.Println("Problem listening for new TCP messages: ")
			log.Println("Stopped peer due to error")
			communicationChannel <- err
			return
		}
	}()

	go func() {
		// Now start UDP listening
		if err := newPeer.StartUDPListening(); err != nil {
			log.Println("Problem listening for new UDP messages: ")
			log.Println("Stopped peer due to error")
			communicationChannel <- err
			return
		}
	}()

	// Sleep for two seconds to give StartTCPListening() time to create the TCPListener
	time.Sleep(time.Second * 2)

	// For testing, check if we need to send a message and if yes, build and send it
	if sendMessage == true {
		log.Println("Create message")
		buildMessage := models.OnionTunnelBuild{OnionTunnelBuild: uint16(560), NetworkVersion: "IPv4", Port: uint16(4200), DestinationAddress: "127.0.0.1", DestinationHostkey: "KEY"}
		onionTunnelBuild := services.CreateOnionTunnelBuild(buildMessage)
		log.Printf("Message: %x\n", onionTunnelBuild)
		newPeer.SendMessage(sendMessageHost, sendMessagePort, onionTunnelBuild)
	}

	// Wait here for errors from the channel
	if err := <- communicationChannel; err != nil {
		log.Println("Yes, error")
		log.Println(err.Error())
		os.Exit(1)
	}
}