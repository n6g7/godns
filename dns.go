package main

import (
	"fmt"
	"net"
)

func serv() {
	addr, _ := net.ResolveUDPAddr("udp4", ":53")
	conn, _ := net.ListenUDP("udp4", addr)
	defer conn.Close()
	fmt.Println("listening on", addr)

	buffer := make([]byte, 1024)
	for {
		n, caddr, _ := conn.ReadFromUDP(buffer)
		rawQuery := buffer[0:n]

		query := parse(rawQuery)
		fmt.Println(caddr, "->", query)

		response, _ := query.Response()
		cname := []string{"test", "example", "com"}
		response.AddCNAMEAnswer(query.question.name, cname, 600)
		response.AddAAnswer(cname, net.IPv4(1, 2, 3, 4), 600)

		rawResponse := response.Dump()
		conn.WriteToUDP(rawResponse, caddr)
		fmt.Println(caddr, "<-", response)
	}
}

func client() {
	saddr, _ := net.ResolveUDPAddr("udp4", "1.1.1.1:53")
	conn, _ := net.DialUDP("udp4", nil, saddr)
	defer conn.Close()

	fmt.Printf("Querying %s\n", conn.RemoteAddr().String())

	var query DNSMessage
	query.Query([]string{"example", "com"})
	fmt.Println(saddr, "<-", query)
	conn.Write(query.Dump())

	buffer := make([]byte, 1024)
	n, caddr, _ := conn.ReadFromUDP(buffer)
	response := parse(buffer[0:n])
	fmt.Println(caddr, "->", response)
}

func main() {
	fmt.Println("godns")

	serv()
	// client()
}
