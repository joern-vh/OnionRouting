package services

import (
	"models"
	"os"
	"errors"
	"github.com/Thomasdezeeuw/ini"
	"strconv"
)

// NewConfigObject creates a new config struct based on an config.ini file, passed as parameter
func NewConfigObject(pathToFile string) (*models.Config, error) {
	// Open configuration file
	file, err := os.Open(pathToFile)
	if err != nil {
		return &models.Config{0, "0"}, errors.New("NewConfigObject: Couldn't open file, is the path right?")
	}
	defer file.Close()

	// Parse the actual configuration
	config, err := ini.Parse(file)
	if err != nil {
		return &models.Config{0, "0"}, errors.New("NewConfigObject: Couldn't parse the config file")
	}

	// Need to be done like this to handle erros
	configP2PPort, err := strconv.Atoi(config["onion"]["p2p_port"])
	if err != nil {
		return &models.Config{0, "0"}, errors.New("NewConfigObject: Couldn't parse P2PPort")
	}

	return &models.Config{configP2PPort, config["onion"]["p2p_hostname"]}, nil
}