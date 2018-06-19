package models

type OnionTunnelBuild struct {
	NetworkVersion			string
	Port					uint16
	DestinationAddress 		string
	DestinationHostkey 		[]byte
}

type OnionTunnelReady struct {
	TunnelID 				string
	DestinationHostkey 		[]byte
}

type OnionTunnelIncoming struct {
	TunnelID 				string
}

type OnionTunnelDestroy struct {
	TunnelID 				string
}

type OnionTunnelData struct {
	TunnelID 				string
	Data 					[]byte
}

type OnionError struct {
	RequestType 			uint16
	Reserved 				uint16
	TunnelID 				string
}

type OnionCover struct {
	CoverSize 				uint16
	Reserved 				uint16
}

type ConstructTunnel struct {
	NetworkVersion			string
	Port					uint16
	DestinationAddress 		string
	DestinationHostkey 		[]byte
}

type ConfirmTunnelConstruction struct {
	Port					uint16
	TunnelID 				string
	DestinationHostkey 		[]byte
}