package resolvers

import (
	"fmt"
	"net"

	"github.com/n6g7/godns/client"
	"github.com/n6g7/godns/proto"
)

type ForwardResolver struct {
	client *client.Client
}

func NewForwardResolver(ip net.IP) *ForwardResolver {
	return &ForwardResolver{
		client: client.NewClient(ip),
	}
}

func (fr *ForwardResolver) Resolve(q proto.Question) ([]*proto.ResourceRecord, error) {
	response, err := fr.client.Resolve(q)
	if err != nil {
		return nil, fmt.Errorf("Error while forwarding to %s: %w", fr.client.Ip, err)
	}
	return response.Answers, nil
}
