package models

import (
	"net"
	"crypto/rsa"
)

// TODO: Discuss wether to define it here or to define in it in the service packe >> Downside here is that calling with Peer as caller isn't possible
// Peer is the standard object for a running peer that is accepting connections
type Peer struct {
	TCPListener		*net.TCPListener				`json:"tcp_listener"`
	UDPListener		*net.UDPConn 					`json:"udp_listener"`
	UDPPort			int 							`json:"udp_port"`
	P2P_Port		int								`json:"p2p_port"`			// This is the Port for the TCP port
	P2P_Hostname	string							`json:"p2p_hostname"`		// This is the ip address of the peer
	PrivateKey		*rsa.PrivateKey					`json:"private_key"`
	PublicKey		*rsa.PublicKey 					`json:"public_key"`
	UDPConnections	map[uint32]*UDPConnection 		`json:"udp_connections"`
	TCPConnections	map[uint32]*TCPConnection 		`json:"tcp_writers"`
}

// Identify by id in hashmap
type TCPConnection struct {
	TunnelId		uint32 			`json:"tunnel_id"`
	LeftWriter		*TCPWriter 		`json:"left_writer"`
	RightWriter		*TCPWriter 		`json:"right_writer"`
}

type TCPWriter struct {
	DestinationIP		string		`json:"destination_ip"`
	DestinationPort		int			`json:"destination_port"`
	TCPWriter			net.Conn 	`json:"tcp_writer"`
}