package services

import (
	"models"
	"net"
	"strconv"
	"errors"
	"log"
)

// Just used to wrap the actual Peer from models.Peer here to use it as caller
type UDPConnection struct {
	UCPConnObject *models.UDPConnection
}

// When Creating Initial, always set writer for left side >> Write back to origin
func CreateInitialUDPConnection(leftHost string, leftPort int, tunndelId uint32, networkVersion string) (*models.UDPConnection, error) {
	// first, create leftWriter
	leftWriter, err := createUDPWriter(leftHost, leftPort)
	if err != nil {
		return nil, errors.New("CreateInitialUDPConnection: Problem creating new UDPWriter, " + err.Error())
	}

	return &models.UDPConnection{tunndelId, networkVersion, leftHost, leftPort, "", 0, leftWriter, nil}, nil
}

// Creates a new writer for a given ip and port
func createUDPWriter(destinationAddress string, destinationPort int) (*net.Conn, error){
	newConn, err := net.Dial("udp", destinationAddress + ":" + strconv.Itoa(destinationPort))
	if err != nil {
		return nil, errors.New("createUDPWriter: Error while creating new wirter, error: " + err.Error())
	}

	log.Println("createUDPWriter: Created new writer for " + destinationAddress)

	return &newConn, nil
}

