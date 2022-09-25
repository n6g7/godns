package interfaces

import (
	"fmt"
	"log"
	"net"

	"github.com/n6g7/godns/proto"
)

type ServerInterface struct {
	address string
	conn    *net.UDPConn
	buffer  []byte
}

func NewServerInterface(port uint16) *ServerInterface {
	return &ServerInterface{
		address: fmt.Sprintf(":%d", port),
		buffer:  make([]byte, 1024),
	}
}

func (serv *ServerInterface) Start() error {
	var err error
	udpaddr, err := net.ResolveUDPAddr("udp4", serv.address)
	if err != nil {
		return fmt.Errorf("Cannot resolve server address: %s", serv.address)
	}

	serv.conn, err = net.ListenUDP("udp4", udpaddr)
	if err != nil {
		return fmt.Errorf("Cannot listen on %s, already in use?", udpaddr)
	}

	log.Printf("listening on %s", udpaddr)

	return nil
}

func (serv *ServerInterface) Stop() error {
	defer serv.conn.Close()

	return nil
}

func (serv *ServerInterface) GetQuery() (*QueryContext, error) {
	n, caddr, err := serv.conn.ReadFromUDP(serv.buffer)
	if err != nil {
		return nil, err
	}

	query, err := proto.ParseMessage(serv.buffer[0:n])
	if err != nil {
		return nil, err
	}

	log.Printf("%s -> %v", caddr, query)

	return &QueryContext{
		Query: query,
		caddr: caddr,
	}, nil
}

func (serv *ServerInterface) Respond(context *QueryContext, response *proto.DNSMessage) error {
	_, err := serv.conn.WriteToUDP(response.Dump(), context.caddr)
	if err != nil {
		return err
	}

	log.Printf("%s <- %v", context.caddr, response)

	return nil
}
