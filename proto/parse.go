package proto

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type LABEL_TYPE uint16

const (
	NORMAL   LABEL_TYPE = 0
	EXTENDED            = 1
	POINTER             = 3
)

type EXTENDED_LABEL_TYPE uint16

const (
	RESERVED EXTENDED_LABEL_TYPE = 63
)

func parseName(request []byte, startpos int) ([]string, int, error) {
	i := startpos
	labels := []string{}
	var nextpos int

	for request[i] != 0 {
		label_type := LABEL_TYPE(request[i] >> 6)
		extra := request[i] & 63

		switch label_type {
		case NORMAL:
			length := int(extra)
			label := request[i+1 : i+length+1]
			labels = append(labels, string(label))
			i += length + 1
		case EXTENDED:
			extendedLabelType := EXTENDED_LABEL_TYPE(extra)
			switch extendedLabelType {
			default:
				return nil, 0, errors.New("We don't support extended label types (yet).")
			}
		case POINTER:
			offset := binary.BigEndian.Uint16(request[i:i+2]) - POINTER<<(6+8)
			if nextpos == 0 {
				nextpos = i + 2
			}
			i = int(offset)
		}
	}

	if nextpos == 0 {
		nextpos = i + 1
	}
	return labels, nextpos, nil
}

func parseResourceRecord(request []byte, startpos int) (*ResourceRecord, int, error) {
	var rr ResourceRecord
	var err error
	i := startpos
	rr.Name, i, err = parseName(request, i)
	if err != nil {
		return nil, 0, fmt.Errorf("Couldn't parse resource record name: %w", err)
	}
	rr.Type = qtype(binary.BigEndian.Uint16(request[i : i+2]))

	switch rr.Type {
	case OPT:
		rr.udpPayloadSize = binary.BigEndian.Uint16(request[i+2 : i+4])
		rr.extRCODE = uint8(request[i+4])
		rr.version = uint8(request[i+5])
		rr.d0 = request[i+6]>>7&1 == 1
		rr.z = binary.BigEndian.Uint16(request[i+6:i+8]) & (2 ^ 16 - 1)
	default:
		rr.Class = class(binary.BigEndian.Uint16(request[i+2 : i+4]))
		rr.Ttl = binary.BigEndian.Uint32(request[i+4 : i+8])
	}

	rdlength := binary.BigEndian.Uint16(request[i+8 : i+10])
	rr.Rdata = request[i+10 : i+10+int(rdlength)]

	return &rr, i + 10 + int(rdlength), nil
}

func parse(request []byte) (*DNSMessage, error) {
	var res DNSMessage
	var err error

	// Headers
	headers := request[0:12]
	res.id = binary.BigEndian.Uint16(headers[0:2])
	flags := headers[2:4]
	res.qr = qr(flags[0] >> 7 & 1)
	res.opcode = opcode(flags[0] >> 3 & 15)
	res.aa = flags[0]>>2&1 == 1
	res.tc = flags[0]>>1&1 == 1
	res.rd = flags[0]>>0&1 == 1
	res.ra = flags[1]>>7&1 == 1
	// z (bits 9-11 of the flags bytes) is ignored
	res.rcode = rcode(flags[1] >> 0 & 15)

	res.qdcount = binary.BigEndian.Uint16(headers[4:6])
	ancount := binary.BigEndian.Uint16(headers[6:8])
	nscount := binary.BigEndian.Uint16(headers[8:10])
	arcount := binary.BigEndian.Uint16(headers[10:12])

	// Questions
	var i int = 12
	res.Question.Name, i, err = parseName(request, i)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse question name: %w", err)
	}
	res.Question.Type = qtype(binary.BigEndian.Uint16(request[i : i+2]))
	res.Question.Class = class(binary.BigEndian.Uint16(request[i+2 : i+4]))
	i += 4

	// Answers
	var answer *ResourceRecord
	for k := 0; k < int(ancount); k++ {
		answer, i, err = parseResourceRecord(request, i)
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse answer RR: %w", err)
		}
		res.Answers = append(res.Answers, answer)
	}

	// Authority
	var authority *ResourceRecord
	for k := 0; k < int(nscount); k++ {
		authority, i, err = parseResourceRecord(request, i)
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse authority RR: %w", err)
		}
		res.Authority = append(res.Authority, authority)
	}

	// Additional
	var additional *ResourceRecord
	for k := 0; k < int(arcount); k++ {
		additional, i, err = parseResourceRecord(request, i)
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse additional RR: %w", err)
		}
		res.Additional = append(res.Additional, additional)
	}

	// Update RCODE with OPT extended value
	if optrr := res.GetOPTPseudoRR(); optrr != nil {
		res.rcode = rcode(uint16(res.rcode) + uint16(optrr.extRCODE<<4))
	}

	return &res, nil
}
