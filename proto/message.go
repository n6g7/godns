package proto

import (
	"errors"
	"math/rand"
	"net"
)

type Question struct {
	Name  []string
	Type  qtype
	Class class
}

type ResourceRecord struct {
	Name  []string
	Type  qtype
	Class class
	Ttl   uint32
	Rdata []byte

	// OPT fields
	udpPayloadSize uint16
	extRCODE       uint8
	version        uint8
	d0             bool
	z              uint16
}

type DNSMessage struct {
	id     uint16
	qr     qr
	opcode opcode
	aa     bool
	tc     bool
	rd     bool
	ra     bool
	rcode  rcode

	qdcount uint16

	Question   Question
	Answers    []*ResourceRecord
	Authority  []*ResourceRecord
	Additional []*ResourceRecord
}

type qr uint8

const (
	QRQuery    qr = 0
	QRResponse    = 1
)

type opcode uint8

const (
	OpcodeQuery  opcode = 0
	OpcodeIQuery        = 1
	OpcodeStatus        = 2
)

type rcode uint16

const (
	RcodeNoError        rcode = 0
	RcodeFormatError          = 1
	RcodeServerFailure        = 2
	RcodeNameError            = 3
	RcodeNotImplemented       = 4
	RcodeRefused              = 5
)

type qtype uint16

const (
	A     qtype = 1
	NS          = 2
	CNAME       = 5
	SOA         = 6
	PTR         = 12
	MX          = 15
	TXT         = 16
	OPT         = 41
)

type class uint16

const (
	IN class = 1
	CS       = 2
	CH       = 3
	HS       = 4
)

func ParseMessage(buffer []byte) (*DNSMessage, error) {
	msg, err := parse(buffer)
	return msg, err
}

func (m *DNSMessage) IsQuery() bool {
	return m.qr == QRQuery
}

func (m *DNSMessage) IsResponse() bool {
	return m.qr == QRResponse
}

func (m *DNSMessage) Query(q Question) {
	m.id = uint16(rand.Uint32())
	m.qr = QRQuery
	m.opcode = OpcodeQuery
	m.rd = true
	m.rcode = RcodeNoError
	m.qdcount = 1
	m.Question = q
}

func (m *DNSMessage) Response() (*DNSMessage, error) {
	if !m.IsQuery() {
		return nil, errors.New("Cannot respond to a response")
	}

	var response DNSMessage
	response.id = m.id
	response.qr = QRResponse
	response.opcode = m.opcode
	response.aa = false
	response.tc = false
	response.rd = m.rd
	response.ra = false // we don't support recursive resolving for now
	response.rcode = RcodeNoError
	response.qdcount = m.qdcount

	response.Question = m.Question

	if opt := m.GetOPTPseudoRR(); opt != nil {
		response.Additional = append(response.Additional, &ResourceRecord{
			Type:           OPT,
			udpPayloadSize: 1024,
		})
	}

	return &response, nil
}

func (m *DNSMessage) Dump() []byte {
	return dump(*m)
}

func (m *DNSMessage) AddAnswer(name []string, type_ qtype, rdata []byte, ttl uint32) {
	var rr ResourceRecord
	rr.Name = name
	rr.Type = type_
	rr.Class = IN
	rr.Ttl = ttl
	rr.Rdata = rdata

	m.Answers = append(m.Answers, &rr)
}

func (m *DNSMessage) AddAAnswer(name []string, ip net.IP, ttl uint32) {
	m.AddAnswer(name, A, ip.To4(), ttl)
}

func (m *DNSMessage) AddCNAMEAnswer(name []string, labels []string, ttl uint32) {
	m.AddAnswer(name, CNAME, DumpName(labels), ttl)
}

func (m *DNSMessage) GetOPTPseudoRR() *ResourceRecord {
	for i := range m.Additional {
		if m.Additional[i].Type == OPT {
			return m.Additional[i]
		}
	}
	return nil
}
