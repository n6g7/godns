package client

import (
	"fmt"
	"log"
	"net"
	"net/netip"

	"github.com/n6g7/godns/proto"
)

type Client struct {
	Ip   net.IP
	port uint16
}

func NewClient(ip net.IP) *Client {
	return &Client{
		Ip:   ip,
		port: 53,
	}
}

func (c *Client) Resolve(q proto.Question) (*proto.DNSMessage, error) {
	addr, ok := netip.AddrFromSlice(c.Ip)
	if !ok {
		return nil, fmt.Errorf("Invalid server address: %s", c.Ip)
	}

	addrport := netip.AddrPortFrom(addr, c.port)
	udpaddr := net.UDPAddrFromAddrPort(addrport)
	conn, err := net.DialUDP("udp4", nil, udpaddr)
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to server %s: %w", udpaddr, err)
	}
	defer conn.Close()

	log.Printf("querying %s", conn.RemoteAddr().String())

	var query proto.DNSMessage
	query.Query(q)

	log.Printf("%s <- %v", udpaddr, query)
	_, err = conn.Write(query.Dump())
	if err != nil {
		return nil, fmt.Errorf("Couldn't send data to server (%s): %w", udpaddr, err)
	}

	buffer := make([]byte, 1024)
	n, caddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("Can't read data from server (%s): %w", udpaddr, err)
	}
	response, err := proto.ParseMessage(buffer[0:n])
	if err != nil {
		return nil, fmt.Errorf("Can't parse server message: %w", err)
	}
	log.Printf("%s -> %v", caddr, response)

	return response, nil
}
