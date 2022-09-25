package resolvers

import "github.com/n6g7/godns/proto"

type Resolver interface {
	Resolve(q proto.Question) ([]proto.ResourceRecord, error)
}
