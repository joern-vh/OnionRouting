package models

type DataConstructTunnel struct {
	NetworkVersion			string
	DestinationAddress 		string
	Port 					uint16
	DestinationHostkey 		[]byte
	PublicKey				[]byte
}

type DataConfirmTunnelConstruction struct {
	DestinationHostkey 		[]byte
	PublicKey				[]byte
}
