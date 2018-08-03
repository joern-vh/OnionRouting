# Group 61 Project Page
## Requirements
* RSA Keypair (4096 bit) in PEM format (`openssl genpkey -algorithm RSA -out keypair.pem -pkeyopt rsa_keygen_bits:4096`)
* Packages
    * `go get github.com/monnand/dhkx`
    * `go get github.com/Thomasdezeeuw/ini`

## Instructions
* run `go run src peer.go -C PATH/TO/CONFIG.INI`

## Status
The master branch contains a runnable version of our implementation.
The encryption branch contains our implementation including encryption. Unfortunately, we had some issues with some hostkeys.

Detailed explanation of our project's status will be stated in our final report.