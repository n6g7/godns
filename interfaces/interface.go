package interfaces

import (
	"net"

	"github.com/n6g7/godns/proto"
)

type Interface interface {
	Start() error
	GetQuery() (*QueryContext, error)
	Respond(context *QueryContext, response *proto.DNSMessage) error
	// RespondWithError(context *QueryContext, err error) error
	Stop() error
}

type QueryContext struct {
	Query *proto.DNSMessage
	caddr *net.UDPAddr
}
