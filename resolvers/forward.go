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

func (fr *ForwardResolver) Resolve(q proto.Question) ([]proto.ResourceRecord, error) {
	addr, ok := netip.AddrFromSlice(fr.ip)
	if !ok {
		return nil, fmt.Errorf("Invalid forward resolver address: %s", fr.ip)
	}

	addrport := netip.AddrPortFrom(addr, fr.port)
	udpaddr := net.UDPAddrFromAddrPort(addrport)
	conn, _ := net.DialUDP("udp4", nil, udpaddr)
	defer conn.Close()

	log.Printf("querying %s", conn.RemoteAddr().String())

	var query proto.DNSMessage
	query.Query(q)

	log.Printf("%s <- %v", udpaddr, query)
	conn.Write(query.Dump())

	buffer := make([]byte, 1024)
	n, caddr, _ := conn.ReadFromUDP(buffer)
	response, err := proto.ParseMessage(buffer[0:n])
	if err != nil {
		return nil, err
	}
	log.Printf("%s -> %v", caddr, response)

	return (*response).Answers, nil
}
