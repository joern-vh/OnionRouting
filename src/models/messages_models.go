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
	//NetworkVersion			string
	OnionPort				uint16
	//TCPPort					uint16
	TunnelID				uint32
	//DestinationAddress 		string
	DestinationHostkey 		[]byte
	OriginHostkey			[]byte
	PublicKey				[]byte
}

type ConfirmTunnelConstruction struct {
	Port					uint16
	TunnelID 				uint32
	DestinationHostkey 		[]byte
	PublicKey				[]byte
}

type TunnelInstruction struct {
	TunnelID 				uint32
	Data					[]byte
}

type ConfirmTunnelInstruction struct {
	TunnelID				uint32
	Data					[]byte
}

type ExchangeKey struct {
	NetworkVersion			string
	DestinationAddress 		string
	TCPPort					uint16
	TunnelID				uint32
	Status					uint16
	DestinationHostkey		[]byte
	PublicKey				[]byte
}

type RPSPeer struct {
	NetworkVersion			string
	OnionPort				uint16
	DestinationAddress		string
	PeerHostkey				[]byte
}