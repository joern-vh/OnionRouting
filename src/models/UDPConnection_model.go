package models

import "net"

type UDPConnection struct {
	TunnelId			uint32				`json:"tunnelId"`
	NetworkVersion		string				`json:"networkVersion"`
	LeftHost			string 				`json:"left_host"`
	LeftPort			int 				`json:"left_port"`
	RightHost			string 				`json:"right_host"`
	RightPort			int 				`json:"right_port"`
	LeftWriter			net.Conn 			`json:"left_writer"`
	RightWriter			net.Conn 			`json:"right_writer"`
	//DestinationAddress	string				`json:"destinationAddress"`
	//DestinationHostKey	[]byte				`json:"destinationHostKey"`
}