package main

import "encoding/binary"

func dumpName(labels []string) []byte {
	var res []byte

	for i := range labels {
		label := labels[i]
		res = append(res, byte(len(label)))
		res = append(res, []byte(label)...)
	}
	res = append(res, 0)

	return res
}

func dumpRR(rr ResourceRecord) []byte {
	var res []byte

	res = append(res, dumpName(rr.name)...)
	res = binary.BigEndian.AppendUint16(res, uint16(rr.atype))

	switch rr.atype {
	case OPT:
		res = binary.BigEndian.AppendUint16(res, uint16(rr.udpPayloadSize))
		res = append(res, rr.extRCODE)
		res = append(res, rr.version)
		res = binary.BigEndian.AppendUint16(res, bool2uint16(rr.d0)<<15+rr.z)
	default:
		res = binary.BigEndian.AppendUint16(res, uint16(rr.class))
		res = binary.BigEndian.AppendUint32(res, rr.ttl)
	}
	res = binary.BigEndian.AppendUint16(res, uint16(len(rr.rdata)))
	res = append(res, rr.rdata...)

	return res
}

func bool2uint16(b bool) uint16 {
	if b {
		return 1
	} else {
		return 0
	}
}

func dump(message DNSMessage) []byte {
	var res []byte

	res = binary.BigEndian.AppendUint16(res, message.id)

	var flags uint16
	flags += uint16(message.qr) & 1 << 15
	flags += uint16(message.opcode) & 15 << 11
	flags += bool2uint16(message.aa) & 1 << 10
	flags += bool2uint16(message.tc) & 1 << 9
	flags += bool2uint16(message.rd) & 1 << 8
	flags += bool2uint16(message.ra) & 1 << 7
	flags += uint16(message.rcode) & 7 << 0
	res = binary.BigEndian.AppendUint16(res, flags)

	res = binary.BigEndian.AppendUint16(res, message.qdcount)
	res = binary.BigEndian.AppendUint16(res, uint16(len(message.answers)))
	res = binary.BigEndian.AppendUint16(res, uint16(len(message.authority)))
	res = binary.BigEndian.AppendUint16(res, uint16(len(message.additional)))

	res = append(res, dumpName(message.question.name)...)
	res = binary.BigEndian.AppendUint16(res, uint16(message.question.qtype))
	res = binary.BigEndian.AppendUint16(res, uint16(message.question.class))

	for i := range message.answers {
		res = append(res, dumpRR(message.answers[i])...)
	}

	for i := range message.authority {
		res = append(res, dumpRR(message.authority[i])...)
	}

	for i := range message.additional {
		res = append(res, dumpRR(message.additional[i])...)
	}

	return res
}
