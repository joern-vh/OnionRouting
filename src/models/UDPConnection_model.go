package models

import "net"

type UDPConnection struct {
	Port				int					`json:"port"`
	UDPConn				*net.UDPConn		`json:"udp_listener"`
	TunnelId			string				`json:"tunnelId"`
	NetworkVersion		string				`json:"networkVersion"`
	DestinationAddress	string				`json:"destinationAddress"`
	DestinationHostKey	[]byte				`json:"destinationHostKey"`
}