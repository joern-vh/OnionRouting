package models

type UDPConnection struct {
	TunnelId			string				`json:"tunnelId"`
	NetworkVersion		string				`json:"networkVersion"`
	LeftPort			int 				`json:"left_port"`
	RightPort			int 				`json:"right_port"`
	//DestinationAddress	string				`json:"destinationAddress"`
	//DestinationHostKey	[]byte				`json:"destinationHostKey"`
}