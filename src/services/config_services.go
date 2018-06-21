package services

import (
	"models"
	"os"
	"errors"
	"github.com/Thomasdezeeuw/ini"
	"strconv"
	"encoding/pem"
	"crypto/x509"
	"io/ioutil"
	"flag"
	"crypto/rsa"
)

// NewConfigObject creates a new config struct based on an config.ini file, passed as parameter
func NewConfigObject() (*models.Config, error) {
	// First, get the path to the config file
	pathConfigFile, err := parseConfigFlag()
	if err != nil {
		return nil, errors.New("NewConfigObject: " + err.Error())
	}
	// Open configuration file
	file, err := os.Open(pathConfigFile)
	if err != nil {
		return &models.Config{0, "0", nil, nil}, errors.New("NewConfigObject: Couldn't open file, is the path right?")
	}
	defer file.Close()

	// Parse the actual configuration
	config, err := ini.Parse(file)
	if err != nil {
		return &models.Config{0, "0", nil, nil}, errors.New("NewConfigObject: Couldn't parse the config file")
	}

	// Need to be done like this to handle erros
	configP2PPort, err := strconv.Atoi(config["onion"]["p2p_port"])
	if err != nil {
		return &models.Config{0, "0", nil, nil}, errors.New("NewConfigObject: Couldn't parse P2PPort")
	}

	privateKey, publicKey, err := ParseKeys(config["onion"]["hostkey"])
	if err != nil {
		return &models.Config{0, "0", nil, nil}, errors.New("NewConfigObject: Couldn't parse private key, error: " + err.Error())
	}

	return &models.Config{configP2PPort, config["onion"]["p2p_hostname"], privateKey, publicKey}, nil
}

func parseKeys(path string) ([]byte, []byte, error) {
	pemKey, err := ioutil.ReadFile(path) // just pass the file name
	if err != nil {
		return nil, nil, errors.New("parseKeys: Error reading file, err: " + err.Error())
	}

	block, _ := pem.Decode([]byte(pemKey))
	parseResult, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	privateKey := parseResult.(*rsa.PrivateKey)
	publicKey := privateKey.PublicKey
	return privateKey.N.Bytes(), publicKey.N.Bytes(), nil
}

// parseConfigFlag parse the key C to get the path to the config.ini
func parseConfigFlag() (string, error){
	// Define and parse command line flags
	configFileFlag := flag.String("C", "config.ini", "Please write the path to your config.ini file ")
	flag.Parse()

	// Check if path was empty
	if *configFileFlag == "" {
		flag.PrintDefaults()
		return "", errors.New("ParseStartFlags: No config.ini path was provided")
	}

	return *configFileFlag, nil
}