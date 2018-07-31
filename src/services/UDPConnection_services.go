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

// When Creating Initial (despite first element in tunnel), always set writer for left side >> Write back to origin
func CreateInitialUDPConnectionLeft(leftHost string, leftPort int, tunndelId uint32) (*models.UDPConnection, error) {
	// first, create leftWriter
	leftWriter, err := CreateUDPWriter(leftHost, leftPort)
	if err != nil {
		return nil, errors.New("CreateInitialUDPConnection: Problem creating new UDPWriter, " + err.Error())
	}

	return &models.UDPConnection{tunndelId, leftHost, leftPort, "", 0, leftWriter, nil}, nil
}

func CreateInitialUDPConnectionRight(rightHost string, rightPort int, tunndelId uint32) (*models.UDPConnection, error) {
	// first, create rightWriter
	rightWriter, err := CreateUDPWriter(rightHost, rightPort)
	if err != nil {
		return nil, errors.New("CreateInitialUDPConnection: Problem creating new UDPWriter, " + err.Error())
	}

	return &models.UDPConnection{tunndelId, rightHost, rightPort, "", 0, nil, rightWriter}, nil
}

// Creates a new writer for a given ip and port
func CreateUDPWriter(destinationAddress string, destinationPort int) (net.Conn, error){
	newConn, err := net.Dial("udp", destinationAddress + ":" + strconv.Itoa(destinationPort))
	if err != nil {
		return nil, errors.New("createUDPWriter: Error while creating new wirter, error: " + err.Error())
	}

	log.Println("createUDPWriter: Created new writer to " + destinationAddress + ", Port: " + strconv.Itoa(destinationPort))

	return newConn, nil
}

