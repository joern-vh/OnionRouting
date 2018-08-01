package models

import (
	"net"
	"crypto/rsa"
	"github.com/monnand/dhkx"
	"container/list"
)

// TODO: Discuss wether to define it here or to define in it in the service packe >> Downside here is that calling with Peer as caller isn't possible
// Peer is the standard object for a running peer that is accepting connections
type Peer struct {
	TCPListener			*net.TCPListener				`json:"tcp_listener"`
	UDPListener			*net.UDPConn 					`json:"udp_listener"`
	UDPPort				int 							`json:"udp_port"`
	P2P_Port			int								`json:"p2p_port"`			// This is the Port for the TCP port
	P2P_Hostname		string							`json:"p2p_hostname"`		// This is the ip address of the peer
	PrivateKey			*rsa.PrivateKey					`json:"private_key"`
	PublicKey			*rsa.PublicKey 					`json:"public_key"`
	UDPConnections		map[uint32]*UDPConnection 		`json:"udp_connections"`
	TCPConnections		map[uint32]*TCPConnection 		`json:"tcp_writers"`
	CryptoSessionMap	map[string]*CryptoObject 		`json:"crypto_session_map"`
	TunnelHostOrder		map[uint32]*list.List			`json:"tunnel_host_order"`			// Save all hashed hostkey of a tunnel connection ordered in a list
}

// Identify by id in hashmap
type TCPConnection struct {
	TunnelId				uint32 							`json:"tunnel_id"`
	LeftWriter				*TCPWriter 						`json:"left_writer"`
	RightWriter				*TCPWriter 						`json:"right_writer"`
	FinalDestination		*OnionTunnelBuild				`json:"final_destination"`
	OriginHostkey			[]byte							`json:"origin_hostkey"`
}

type TCPWriter struct {
	DestinationIP		string					`json:"destination_ip"`
	DestinationPort		int						`json:"destination_port"`
	TCPWriter			net.Conn 				`json:"tcp_writer"`
}

type CryptoObject struct {
	TunnelId		uint32 							`json:"tunnel_id"`
	PrivateKey		*dhkx.DHKey 						`json:"tunnel_id"`
	PublicKey		[]byte 							`json:"tunnel_id"`
	SessionKey		[]byte 							`json:"tunnel_id"`
	Group			*dhkx.DHGroup 					`json:"tunnel_id"`
}