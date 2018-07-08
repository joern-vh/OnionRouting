package services

/*func CreateDataConstructTunnel(exchangeKey models.ExchangeKey) ([]byte) {
	// Message Type
	command := uint16(567)

	// Convert command to Byte array
	messageTypeBuf := new(bytes.Buffer)
	binary.Write(messageTypeBuf, binary.BigEndian, command)
	message := messageTypeBuf.Bytes()

	// Convert tunnelID to Byte Array
	tunnelIDBuf := new(bytes.Buffer)
	//newID := CreateTunnelID()
	binary.Write(tunnelIDBuf, binary.BigEndian, constructTunnel.TunnelID)
	message = append(message, tunnelIDBuf.Bytes()...)

	// Append Delimiter
	message = append(message, []byte("\r\n")...)

	// Prepend size of message
	sizeBuf := new(bytes.Buffer)
	binary.Write(sizeBuf, binary.BigEndian, uint16(len(message)+2))
	message = append(sizeBuf.Bytes(), message...)

	return message
}*/