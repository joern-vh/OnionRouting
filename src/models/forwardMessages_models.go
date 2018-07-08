package models

type DataConstructTunnel struct {
	TunnelID				uint32
	NetworkVersion			string
	DestinationAddress 		string
	DestinationHostkey 		[]byte
}
