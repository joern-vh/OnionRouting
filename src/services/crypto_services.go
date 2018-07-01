package services

import (
	"io/ioutil"
	"errors"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"github.com/monnand/dhkx"
	"crypto/sha256"
	"crypto/rand"
	"log"
)

// Parse RSA keys (used for DH key exchange) from given path
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

// Generate a pre-master key for DH algorithm
func GeneratePreMasterKey() (*dhkx.DHKey, []byte, *dhkx.DHGroup) {
	g, _ := dhkx.GetGroup(0)

	priv, _ := g.GeneratePrivateKey(nil)
	pub := priv.Bytes()

	// Send Public Key
	return priv, pub, g
}

// Compute an ephemeral key using DH algorithm
func ComputeEphemeralKey(g *dhkx.DHGroup, receivedPublicKey []byte, priv *dhkx.DHKey) ([]byte){
	recvPubKey := dhkx.NewPublicKey(receivedPublicKey)

	// Compute the key
	k, _ := g.ComputeKey(recvPubKey, priv)

	// Get the key in the form of []byte
	key := k.Bytes()

	keyHash := sha256.Sum256(key)

	return keyHash[:]
}

// Encrypt DH key for key exchange
func EncryptKeyExchange(publicKey *rsa.PublicKey, key []byte) ([]byte, error) {
	label := []byte("")
	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, key, []byte(label))
	if err != nil {
		log.Fatalf("Encrypt: %s", err)
		return nil, errors.New("Crypto: New Error occurred while encrypting DH Public key: " + err.Error())
	}

	return encryptedData, nil
}

// Decrypt DH key for key exchange
func DecryptKeyExchange(privateKey *rsa.PrivateKey, key []byte) ([]byte, error) {
	label := []byte("")
	decryptedData, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, key, []byte(label))
	if err != nil {
		log.Fatalf("Decrypt: %s", err)
		return nil, errors.New("Crypto: New Error occurred while decrypting DH Public key: " + err.Error())
	}

	return decryptedData, nil
}

// Encrypt Data with DH key.
func EncryptData(key []byte, data []byte) ([]byte, error) {
	/*block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("Crypto: New Error occurred while encrypting: " + err.Error())
	}

	//stream := cipher.NewCFBEncrypter(block, data)
*/
	return nil, nil
}

// Decrypt Data with DH key.
func DecryptData(key []byte, data []byte) ([]byte, error) {
	return nil, nil
}
