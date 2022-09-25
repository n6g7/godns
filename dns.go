package main

import (
	"log"
	"net"

	"github.com/n6g7/godns/interfaces"
	"github.com/n6g7/godns/resolvers"
)

func run(iface interfaces.Interface, resolver resolvers.Resolver) {
	server := interfaces.NewServerInterface(53)
	err := server.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Stop()

	for {
		context, err := server.GetQuery()
		if err != nil {
			log.Fatal("error!")
		}

		answers, err := resolver.Resolve(context.Query.Question)
		if err != nil {
			log.Fatal("No!")
		}
		response, err := context.Query.Response()
		if err != nil {
			log.Fatal("No!")
		}
		response.Answers = answers

		err = server.Respond(context, response)
		if err != nil {
			log.Fatal("No!")
		}
	}
}

func main() {
	log.Println("godns")

	iface := interfaces.NewServerInterface(53)

	resolver := resolvers.NewForwardResolver(net.IPv4(1, 1, 1, 1), 53)
	// resolver := resolvers.NewStaticResolver([]proto.ResourceRecord{
	// 	{Name: []string{"test", "example", "com"}, Type: proto.CNAME, Class: proto.IN, Ttl: 600, Rdata: proto.DumpName([]string{"test", "example", "com"})},
	// 	{Name: []string{"test", "example", "com"}, Type: proto.A, Class: proto.IN, Ttl: 600, Rdata: net.IPv4(1, 2, 3, 4).To4()},
	// })

	run(iface, resolver)
}
