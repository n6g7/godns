package proto

import "encoding/binary"

func DumpName(labels []string) []byte {
	var res []byte

	for i := range labels {
		label := labels[i]
		res = append(res, byte(len(label)))
		res = append(res, []byte(label)...)
	}
	res = append(res, 0)

	return res
}

func DumpRR(rr *ResourceRecord) []byte {
	var res []byte

	res = append(res, DumpName(rr.Name)...)
	res = binary.BigEndian.AppendUint16(res, uint16(rr.Type))

	switch rr.Type {
	case OPT:
		res = binary.BigEndian.AppendUint16(res, uint16(rr.udpPayloadSize))
		res = append(res, rr.extRCODE)
		res = append(res, rr.version)
		res = binary.BigEndian.AppendUint16(res, bool2uint16(rr.d0)<<15+rr.z)
	default:
		res = binary.BigEndian.AppendUint16(res, uint16(rr.Class))
		res = binary.BigEndian.AppendUint32(res, rr.Ttl)
	}
	res = binary.BigEndian.AppendUint16(res, uint16(len(rr.Rdata)))
	res = append(res, rr.Rdata...)

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
	res = binary.BigEndian.AppendUint16(res, uint16(len(message.Answers)))
	res = binary.BigEndian.AppendUint16(res, uint16(len(message.Authority)))
	res = binary.BigEndian.AppendUint16(res, uint16(len(message.Additional)))

	res = append(res, DumpName(message.Question.Name)...)
	res = binary.BigEndian.AppendUint16(res, uint16(message.Question.Type))
	res = binary.BigEndian.AppendUint16(res, uint16(message.Question.Class))

	for i := range message.Answers {
		res = append(res, DumpRR(message.Answers[i])...)
	}

	for i := range message.Authority {
		res = append(res, DumpRR(message.Authority[i])...)
	}

	for i := range message.Additional {
		res = append(res, DumpRR(message.Additional[i])...)
	}

	return res
}
