package services

import (
	"models"
	"os"
	"errors"
	"github.com/Thomasdezeeuw/ini"
	"strconv"
	"flag"
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

// ParseStartFlags check whether one of the available flags is used
func ParseStartFlags() (string, bool, string, int, error){
	// Define and parse command line flags
	configFileFlag := flag.String("c", "config.ini", "Please write the path to your config.ini file ")
	sendMessage := flag.Bool("message", false, "Use parameter m to show that we need parameter h & p")
	sendMessageHost := flag.String("h", "127.0.0.1", "Please enter the ip address of your destination host")
	sendMessagePort := flag.String("p", "127.0.0.1", "Please enter the ip address of your destination host")
	flag.Parse()

	// Check if path was empty
	if *configFileFlag == "" {
		flag.PrintDefaults()
		return "", false, "", 0, errors.New("ParseStartFlags: No config.ini path was provided")
	}

	// Check if we need to send a message
	if *sendMessage == true {
		// Check that port and host aren't empty
		if *sendMessageHost == "" || *sendMessagePort == "" {
			flag.PrintDefaults()
			return "", false, "", 0, errors.New("ParseStartFlags: Please provide an ip and a port when sending a message")
		}

		// Convert sendMessagePort to int
		sendMessagePortInt, err := strconv.Atoi(*sendMessagePort)
		if err != nil {
			flag.PrintDefaults()
			return "", false, "", 0, errors.New("ParseStartFlags: Please provide a valid port as number(int)")
		}

		return *configFileFlag, *sendMessage, *sendMessageHost, sendMessagePortInt, nil
	}

	return *configFileFlag, *sendMessage, "", 0, nil
}