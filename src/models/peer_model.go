package models

import "net"

// TODO: Discuss wether to define it here or to define in it in the service packe >> Downside here is that calling with Peer as caller isn't possible
// Peer is the standard object for a running peer that is accepting connections
type Peer struct {
	TCPListener		*net.TCPListener	`json:"tcp_listener"`
	UDPConn			*net.UDPConn		`json:"udp_listener"`
	P2P_Port		int					`json:"p2p_port"`			// This is the Port for the TCP port
	P2P_Hostname	string				`json:"p2p_hostname"`		// This is the ip address of the peer
}