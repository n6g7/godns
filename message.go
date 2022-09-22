package main

import (
	"errors"
	"math/rand"
	"net"
)

type Question struct {
	name  []string
	qtype QTYPE
	class QCLASS
}

type ResourceRecord struct {
	name  []string
	atype QTYPE
	class QCLASS
	ttl   uint32
	rdata []byte
}

type DNSMessage struct {
	id     uint16
	qr     QR
	opcode OPCODE
	aa     bool
	tc     bool
	rd     bool
	ra     bool
	rcode  RCODE

	qdcount uint16

	question   Question
	answers    []ResourceRecord
	authority  []ResourceRecord
	additional []ResourceRecord
}

type QR uint

const (
	Query    QR = 0
	Response    = 1
)

type OPCODE uint

const (
	QUERY  OPCODE = 0
	IQUERY        = 1
	STATUS        = 2
)

type RCODE uint

const (
	NoError        RCODE = 0
	FormatError          = 1
	ServerFailure        = 2
	NameError            = 3
	NotImplemented       = 4
	Refused              = 5
)

type QTYPE uint16

const (
	A     QTYPE = 1
	NS          = 2
	CNAME       = 5
	SOA         = 6
	PTR         = 12
	MX          = 15
	TXT         = 16
)

type QCLASS uint16

const (
	IN QCLASS = 1
	CS        = 2
	CH        = 3
	HS        = 4
)

func (m *DNSMessage) IsQuery() bool {
	return m.qr == Query
}

func (m *DNSMessage) IsResponse() bool {
	return m.qr == Response
}

func (m *DNSMessage) Query(label []string) {
	m.id = uint16(rand.Uint32())
	m.qr = Query
	m.opcode = QUERY
	m.rd = true
	m.rcode = NoError
	m.qdcount = 1
	m.question = Question{
		name:  label,
		qtype: A,
		class: IN,
	}
}

func (m *DNSMessage) Response() (*DNSMessage, error) {
	if !m.IsQuery() {
		return nil, errors.New("message is not a query")
	}

	var response DNSMessage
	response.id = m.id
	response.qr = Response
	response.opcode = m.opcode
	response.aa = false
	response.tc = false
	response.rd = m.rd
	response.ra = false // we don't support recursive resolving for now
	response.rcode = NoError
	response.qdcount = m.qdcount

	response.question = m.question

	return &response, nil
}

func (m *DNSMessage) Dump() []byte {
	return dump(*m)
}

func (m *DNSMessage) AddAnswer(name []string, qtype QTYPE, rdata []byte, ttl uint32) {
	var rr ResourceRecord
	rr.name = name
	rr.atype = qtype
	rr.class = IN
	rr.ttl = ttl
	rr.rdata = rdata

	m.answers = append(m.answers, rr)
}

func (m *DNSMessage) AddAAnswer(name []string, ip net.IP, ttl uint32) {
	m.AddAnswer(name, A, ip.To4(), ttl)
}

func (m *DNSMessage) AddCNAMEAnswer(name []string, labels []string, ttl uint32) {
	m.AddAnswer(name, CNAME, dumpName(labels), ttl)
}
