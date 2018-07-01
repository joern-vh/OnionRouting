package models

type OnionTunnelBuild struct {
	NetworkVersion			string
	Port					uint16
	DestinationAddress 		string
	DestinationHostkey 		[]byte
}

type OnionTunnelReady struct {
	TunnelID 				uint32
	DestinationHostkey 		[]byte
}

type OnionTunnelIncoming struct {
	TunnelID 				uint32
}

type OnionTunnelDestroy struct {
	TunnelID 				uint32
}

type OnionTunnelData struct {
	TunnelID 				uint32
	Data 					[]byte
}

type OnionError struct {
	RequestType 			uint16
	Reserved 				uint16
	TunnelID 				uint32
}

type OnionCover struct {
	CoverSize 				uint16
	Reserved 				uint16
}

type ConstructTunnel struct {
	NetworkVersion			string
	Port					uint16
	TunnelID				uint32
	DestinationAddress 		string
	DestinationHostkey 		[]byte
}

type ConfirmTunnelConstruction struct {
	Port					uint16
	TunnelID 				uint32
	DestinationHostkey 		[]byte
}

type TunnelInstruction struct {
	Command					uint16
	TunnelID 				uint32
	Data					[]byte
}

type ExchangeKey struct {
	PublicKey				[]byte
}