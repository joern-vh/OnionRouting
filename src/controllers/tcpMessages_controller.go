package controllers

import (
	"encoding/binary"
	"errors"
	"log"
	"fmt"
)

func HandleTCPMessage(message []byte) error {
	//var messageTypeByte []byte = message[3:5]
	messageValue := binary.BigEndian.Uint16(message[2:4])
	fmt.Printf("%d\n", messageValue)

	switch messageValue {
		case 560:
			log.Println("ONION TUNNEL BUILD")
			networkVersion := binary.BigEndian.Uint16(message[4:6])
			onionPort := binary.BigEndian.Uint16(message[6:8])
			networkVersionString := ""
			destinationAddress := ""
			if networkVersion == 0 {
				networkVersionString = "IPv4"
				destinationAddress = string(message[8:12])
			} else if networkVersion == 1 {
				networkVersionString = "IPv6"
				destinationAddress = string(message[8:24])
			}

			fmt.Printf("Network Version: %s\n", networkVersionString)
			fmt.Printf("Onion Port: %d\n", onionPort)
			fmt.Printf("Destination Address: %s\n", destinationAddress)
			break
		case 561:
			log.Println("ONION TUNNEL READY")
			break
		case 562:
			log.Println("ONION TUNNEL INCOMING")
			break
		case 563:
			log.Println("ONION TUNNEL DESTROY")
			break
		default:
			return errors.New("tcpMessagesController: Message Type not Found")
	}

	return nil
}