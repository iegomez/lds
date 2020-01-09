package lds

import (
	"net"
	"errors"
)

// NSClient is a raw UDP client
type NSClient struct {
	Connected bool
	Server string
	Port int
}

func (client *NSClient) send(bytes []byte) error {
	ip := net.ParseIP(client.Server)

	if ip == nil {
		return errors.New("bad network server IP")
	}

	addr := net.UDPAddr{
		IP:   ip,
		Port: client.Port,
	}

	conn, err := net.DialUDP("udp", nil, &addr)
 
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(bytes)
	return err
}
