package services

import (
	"strings"
	"net"
	"math/rand"
)

const symbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func GetIPOutOfAddr(addr string) string {
	if idx := strings.Index(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// getFreePort returns a new free port
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

// used to generate random data
func GenRandomData() []byte{
	length := (rand.Intn(300 - 0) + 0)
	b := make([]byte, length)
	for i := range b {
		b[i] = symbols[rand.Intn(len(symbols))]
	}
	return b
}
