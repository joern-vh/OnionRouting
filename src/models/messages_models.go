package models

type OnionTunnelBuild struct {
	OnionTunnelBuild 		uint16
	NetworkVersion			string
	Port					uint16
	DestinationAddress 		string
	DestinationHostkey 		string
}

type OnionTunnelReady struct {
	OnionTunnelReady 		uint16
	TunnelID 				string
	DestinationHostkey 		string
}

type OnionTunnelIncoming struct {
	OnionTunnelIncoming 	uint16
	TunnelID 				string
}

type OnionTunnelDestroy struct {
	OnionTunnelDestroy 		uint16
	TunnelID 				string
}

type OnionTunnelData struct {
	OnionTunnelData 		uint16
	TunnelID 				string
	Data 					[]byte
}

type OnionError struct {
	OnionError 				uint16
	RequestType 			uint16
	Reserved 				uint16
	TunnelID 				string
}

type OnionCover struct {
	OnionCover 				uint16
	CoverSize 				uint16
	Reserved 				uint16
}