package main

import (
	"services"
	"log"
	"os"
)

func main() {
	config, err := services.NewConfigObject()
	if err != nil {
		log.Println("Error creating config object: " + err.Error())
		os.Exit(1)
	}

	// Open configuration file
	//priv, pub, err := services.ParseKeys("testkey.pem")

	if err != nil {
		log.Println("Failed parsing sencond keys.")
	}

	// Encryption Part

	message := []byte("Hello World!")

	log.Println("Original Message: ", string(message))

	encrytedMessage, err := services.EncryptData(config.PublicKey, message)
	if err != nil {
		log.Println("Error while encrypting.")
	}

	log.Println(len(encrytedMessage))

	/*encrytedMessage2, err := services.EncryptData(pub, encrytedMessage[0:230])
	if err != nil {
		log.Println("Error while encrypting.")
	}

	log.Println("Encrypted Message: ", encrytedMessage2)

	decryptedMessage2, err := services.DecryptData(priv, encrytedMessage2)

	log.Println("Decrypted Message: ", string(decryptedMessage2))*/

	decryptedMessage, err := services.DecryptData(config.PrivateKey, encrytedMessage)

	log.Println("Decrypted Message: ", string(decryptedMessage))


	//log.Println("Key: ", priv)
}
