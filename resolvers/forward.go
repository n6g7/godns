package resolvers

import (
	"fmt"
	"log"
	"net"
	"net/netip"

	"github.com/n6g7/godns/proto"
)

type ForwardResolver struct {
	ip   net.IP
	port uint16
}

func NewForwardResolver(ip net.IP, port uint16) *ForwardResolver {
	return &ForwardResolver{
		ip:   ip,
		port: port,
	}
}

func (fr *ForwardResolver) Resolve(q proto.Question) ([]*proto.ResourceRecord, error) {
	addr, ok := netip.AddrFromSlice(fr.ip)
	if !ok {
		return nil, fmt.Errorf("Invalid forward resolver address: %s", fr.ip)
	}

	addrport := netip.AddrPortFrom(addr, fr.port)
	udpaddr := net.UDPAddrFromAddrPort(addrport)
	conn, err := net.DialUDP("udp4", nil, udpaddr)
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to upstream %s: %w", udpaddr, err)
	}
	defer conn.Close()

	log.Printf("querying %s", conn.RemoteAddr().String())

	var query proto.DNSMessage
	query.Query(q)

	log.Printf("%s <- %v", udpaddr, query)
	_, err = conn.Write(query.Dump())
	if err != nil {
		return nil, fmt.Errorf("Couldn't send data to upstream (%s): %w", udpaddr, err)
	}

	buffer := make([]byte, 1024)
	n, caddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("Can't read data from upstream (%s): %w", udpaddr, err)
	}
	response, err := proto.ParseMessage(buffer[0:n])
	if err != nil {
		return nil, fmt.Errorf("Can't parse upstream message: %w", err)
	}
	log.Printf("%s -> %v", caddr, response)

	return (*response).Answers, nil
}
