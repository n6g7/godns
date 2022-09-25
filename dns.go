package main

import (
	"log"
	"net"

	"github.com/n6g7/godns/proto"
	"github.com/n6g7/godns/resolvers"
)

func serv(resolver resolvers.Resolver) {
	addr, _ := net.ResolveUDPAddr("udp4", ":53")
	conn, _ := net.ListenUDP("udp4", addr)
	defer conn.Close()
	log.Printf("listening on %s", addr)

	buffer := make([]byte, 1024)
	for {
		n, caddr, _ := conn.ReadFromUDP(buffer)
		rawQuery := buffer[0:n]

		query, err := proto.ParseMessage(rawQuery)
		if err != nil {
			log.Fatal("deal with this later")
		}
		log.Printf("%s -> %v", caddr, query)

		answers, err := resolver.Resolve(query.Question)

		response, err := query.Response()
		response.Answers = answers

		rawResponse := response.Dump()
		conn.WriteToUDP(rawResponse, caddr)
		log.Printf("%s <- %v", caddr, response)
	}
}

func main() {
	log.Println("godns")

	resolver := resolvers.NewForwardResolver(net.IPv4(1, 1, 1, 1), 53)
	// resolver := resolvers.NewStaticResolver([]proto.ResourceRecord{
	// 	{Name: []string{"test", "example", "com"}, Type: proto.CNAME, Class: proto.IN, Ttl: 600, Rdata: proto.DumpName([]string{"test", "example", "com"})},
	// 	{Name: []string{"test", "example", "com"}, Type: proto.A, Class: proto.IN, Ttl: 600, Rdata: net.IPv4(1, 2, 3, 4).To4()},
	// })

	serv(resolver)
}
