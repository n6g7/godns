package resolvers

import "github.com/n6g7/godns/proto"

type StaticResolver struct {
	answers             []*proto.ResourceRecord
	overrideFirstAnswer bool
}

func NewStaticResolver(answers []*proto.ResourceRecord) *StaticResolver {
	return &StaticResolver{
		answers:             answers,
		overrideFirstAnswer: true,
	}
}

func (sr *StaticResolver) Resolve(q proto.Question) ([]*proto.ResourceRecord, error) {
	if sr.overrideFirstAnswer && len(sr.answers) > 0 {
		sr.answers[0].Name = q.Name
	}
	return sr.answers, nil
}
