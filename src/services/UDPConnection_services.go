package services

import (
	"models"
)

// Just used to wrap the actual Peer from models.Peer here to use it as caller
type UDPConnection struct {
	UCPConnObject *models.UDPConnection
}

/*
// createNewUDPConnection creates a new udp connection for a peer
func CreateNewUDPConnection(port int, networkVersion string, destinationAddress string, destinationHostKey []byte) (*UDPConnection, error) {
	newUDPConn, err := createUDPConn(port)
	if err != nil {
		return nil, errors.New("CreateNewUDPConnection: Error creating conn" + err.Error())
	}

	// TODO: Generate tunnel id

	return &UDPConnection{ &models.UDPConnection{port, newUDPConn, "test", networkVersion, destinationAddress, destinationHostKey}}, nil
}*/

