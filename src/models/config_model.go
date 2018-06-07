package models

// Model used to parse all necessary information out of an config.ini file TODO: Adapt later to actual config file, now just testing
type Config struct {
	P2P_Port		int			`json:"p2p_port"`
	P2P_Hostname	string		`json:"p2p_hostname"`
}