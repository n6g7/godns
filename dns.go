package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/n6g7/godns/interfaces"
	"github.com/n6g7/godns/proto"
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
	var ifaceType, resolverType, resolverIP string
	var serverPort uint

	flag.StringVar(&ifaceType, "interface", "server", "Interface to receive queries from (cli, server)")
	flag.StringVar(&resolverType, "resolver", "forward", "Resolution mode (forward)")
	flag.UintVar(&serverPort, "port", 53, "Port to listen on when using the server interface")
	flag.StringVar(&resolverIP, "ip", "", "Forward resolver IP")
	flag.Usage = usage
	flag.Parse()

	log.Println("godns")

	var iface interfaces.Interface
	var resolver resolvers.Resolver

	switch ifaceType {
	case "server":
		iface = interfaces.NewServerInterface(uint16(serverPort))
	case "cli":
		log.Fatal("CLI interface not implemented")
	default:
		log.Fatalf("Unknown interface: %s", ifaceType)
	}

	switch resolverType {
	case "forward":
		resolver = resolvers.NewForwardResolver(net.ParseIP(resolverIP), 53)
	case "static":
		resolver = resolvers.NewStaticResolver([]proto.ResourceRecord{
			{Name: []string{"test", "example", "com"}, Type: proto.CNAME, Class: proto.IN, Ttl: 600, Rdata: proto.DumpName([]string{"test", "example", "com"})},
			{Name: []string{"test", "example", "com"}, Type: proto.A, Class: proto.IN, Ttl: 600, Rdata: net.IPv4(1, 2, 3, 4).To4()},
		})
	default:
		log.Fatalf("Unknown resolver: %s", resolverType)
	}

	run(iface, resolver)
}

func usage() {
	fmt.Printf("Usage:\n\t%s [options]\n\nOptions:\n", os.Args[0])
	flag.PrintDefaults()
}
