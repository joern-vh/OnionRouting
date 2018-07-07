package main

import (
	"services"
	"log"
	"crypto/sha256"
	"crypto/x509"
)

func main() {
	/*config, err := services.NewConfigObject()
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

	/*decryptedMessage, err := services.DecryptData(config.PrivateKey, encrytedMessage)

	log.Println("Decrypted Message: ", string(decryptedMessage))*/


	//log.Println("Key: ", priv)

	priv, pub, g := services.GeneratePreMasterKey()

	priv2, pub2, g2 := services.GeneratePreMasterKey()

	k := services.ComputeEphemeralKey(g, pub2, priv)
	k2 := services.ComputeEphemeralKey(g2, pub, priv2)

	log.Println("Key1: ", k)
	log.Println("Key2: ", k2)

	data := []byte("Hello World! This is an awesome Day and I have no clue what else to write...")

	encryptedData,_ := services.EncryptData(k, data)

	log.Println("Encrypted Data: ", encryptedData)

	encryptedData2,_ := services.EncryptData(k, encryptedData)

	log.Println("Encrypted Data 2: ", encryptedData2)

	decryptedData,_ := services.DecryptData(k2, encryptedData2)

	log.Println("Decrypted Data: ", string(decryptedData))

	decryptedData2,_ := services.DecryptData(k2, encryptedData)

	log.Println("Decrypted Data 2: ", string(decryptedData2))

	_, pubRSA, _ := services.ParseKeys("testkey.pem")


	log.Printf("Identity: %d\n", len(sha256.Sum256(x509.MarshalPKCS1PublicKey(pubRSA))))

}
