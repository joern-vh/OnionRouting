#  Student Project Onion Routing
## Requirements
* RSA Keypair (4096 bit) in PEM format (`openssl genpkey -algorithm RSA -out keypair.pem -pkeyopt rsa_keygen_bits:4096`)
* Packages
    * `go get github.com/monnand/dhkx`
    * `go get github.com/Thomasdezeeuw/ini`

## Instructions
* run `go run src peer.go -C PATH/TO/CONFIG.INI`

## Status
The master branch contains a runnable version of our OnionRouting implementation, however encryption is not working.

The encryption branch contains the implementation including encryption (still there are some issues with the hostkeys).

Detailed explanation of our project's status are stated in the final report.

## Contributors
* Jan-Cedric Anslinger ([CedricJAnslinger](https://github.com/CedricJAnslinger))
* Jörn von Henning ([joern-vh](https://github.com/joern-vh))
