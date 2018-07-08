package models

type DataConstructTunnel struct {
	NetworkVersion			string
	DestinationAddress 		string
	DestinationHostkey 		[]byte
}

type DataConfirmTunnelConstruction struct {
	DestinationHostkey 		[]byte
}
