package models

type OnionTunnelBuild struct {
	Size 					uint16
	OnionTunnelBuild 		uint16
	NetworkVersion			string
	Port					uint16
	DestinationAddress 		string
	DestinationHostkey 		string
}

type OnionTunnelReady struct {
	Size 					uint16
	OnionTunnelReady 		uint16
	TunnelID 				string
	DestinationHostkey 		string
}

type OnionTunnelIncoming struct {
	Size 					uint16
	OnionTunnelIncoming 	uint16
	TunnelID 				string
}

type OnionTunnelDestroy struct {
	Size 					uint16
	OnionTunnelDestroy 		uint16
	TunnelID 				string
}

type OnionTunnelData struct {
	Size 					uint16
	OnionTunnelData 		uint16
	TunnelID 				string
	Data 					[]byte
}

type OnionError struct {
	Size 					uint16
	OnionError 				uint16
	RequestType 			uint16
	Reserved 				uint16
	TunnelID 				string
}

type OnionCover struct {
	Size 					uint16
	OnionCover 				uint16
	CoverSize 				uint16
	Reserved 				uint16
}