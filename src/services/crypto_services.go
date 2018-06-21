package services

import (
	"io/ioutil"
	"errors"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"crypto/rand"
	"log"
	"crypto/sha256"
)

func ParseKeys(path string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	pemKey, err := ioutil.ReadFile(path) // just pass the file name
	if err != nil {
		return nil, nil, errors.New("parseKeys: Error reading file, err: " + err.Error())
	}

	block, _ := pem.Decode([]byte(pemKey))
	parseResult, _ := x509.ParsePKCS8PrivateKey(block.Bytes)

	privateKey := parseResult.(*rsa.PrivateKey)
	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil
}

func EncryptData(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	label := []byte("")
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, []byte(label))
	if err != nil {
		log.Fatalf("encrypt: %s", err)
		return nil, errors.New("Crypto: New Error occurred while encrypting: " + err.Error())
	}

	return encryptedData, nil
}

func DecryptData(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	label := []byte("")
	decryptedData, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data, []byte(label))
	if err != nil {
		log.Fatalf("decrypt: %s", err)
		return nil, errors.New("Crypto: New Error occurred while decrypting: " + err.Error())
	}

	return decryptedData, nil
}
