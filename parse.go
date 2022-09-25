package main

import (
	"encoding/binary"
)

func parseName(request []byte, startpos int) ([]string, int) {
	i := startpos
	labels := []string{}
	var nextpos int

	for request[i] != 0 {
		// Pointer
		if request[i]>>6 == 3 {
			offset := binary.BigEndian.Uint16(request[i:i+2]) - 3<<(6+8)
			if nextpos == 0 {
				nextpos = i + 2
			}
			i = int(offset)
			continue
		}

		l := int(request[i])
		label := request[i+1 : i+l+1]
		labels = append(labels, string(label))
		i += l + 1
	}

	if nextpos == 0 {
		nextpos = i + 1
	}
	return labels, nextpos
}

func parseResourceRecord(request []byte, startpos int) (ResourceRecord, int) {
	var rr ResourceRecord
	i := startpos
	rr.name, i = parseName(request, i)
	rr.atype = QTYPE(binary.BigEndian.Uint16(request[i : i+2]))

	switch rr.atype {
	case OPT:
		rr.udpPayloadSize = binary.BigEndian.Uint16(request[i+2 : i+4])
		rr.extRCODE = uint8(request[i+4])
		rr.version = uint8(request[i+5])
		rr.d0 = request[i+6]>>7&1 == 1
		rr.z = binary.BigEndian.Uint16(request[i+6:i+8]) & (2 ^ 16 - 1)
	default:
		rr.class = QCLASS(binary.BigEndian.Uint16(request[i+2 : i+4]))
		rr.ttl = binary.BigEndian.Uint32(request[i+4 : i+8])
	}

	rdlength := binary.BigEndian.Uint16(request[i+8 : i+10])
	rr.rdata = request[i+10 : i+10+int(rdlength)]

	return rr, i + 10 + int(rdlength)
}

func parse(request []byte) DNSMessage {
	var res DNSMessage

	// Headers
	headers := request[0:12]
	res.id = binary.BigEndian.Uint16(headers[0:2])
	flags := headers[2:4]
	res.qr = QR(flags[0] >> 7 & 1)
	res.opcode = OPCODE(flags[0] >> 3 & 15)
	res.aa = flags[0]>>2&1 == 1
	res.tc = flags[0]>>1&1 == 1
	res.rd = flags[0]>>0&1 == 1
	res.ra = flags[1]>>7&1 == 1
	// z (bits 9-11 of the flags bytes) is ignored
	res.rcode = RCODE(flags[1] >> 0 & 15)

	res.qdcount = binary.BigEndian.Uint16(headers[4:6])
	ancount := binary.BigEndian.Uint16(headers[6:8])
	nscount := binary.BigEndian.Uint16(headers[8:10])
	arcount := binary.BigEndian.Uint16(headers[10:12])

	// Questions
	var i int = 12
	res.question.name, i = parseName(request, i)
	res.question.qtype = QTYPE(binary.BigEndian.Uint16(request[i : i+2]))
	res.question.class = QCLASS(binary.BigEndian.Uint16(request[i+2 : i+4]))
	i += 4

	// Answers
	var answer ResourceRecord
	for k := 0; k < int(ancount); k++ {
		answer, i = parseResourceRecord(request, i)
		res.answers = append(res.answers, answer)
	}

	// Authority
	var authority ResourceRecord
	for k := 0; k < int(nscount); k++ {
		authority, i = parseResourceRecord(request, i)
		res.authority = append(res.authority, authority)
	}

	// Additional
	var additional ResourceRecord
	for k := 0; k < int(arcount); k++ {
		additional, i = parseResourceRecord(request, i)
		res.additional = append(res.additional, additional)
	}

	// Update RCODE with OPT extended value
	if optrr := res.GetOPTPseudoRR(); optrr != nil {
		res.rcode = RCODE(uint16(res.rcode) + uint16(optrr.extRCODE<<4))
	}

	return res
}
