package resolvers

import (
	"fmt"
	"net"
	"reflect"

	randmap "github.com/lukechampine/randmap/safe"
	"github.com/n6g7/godns/client"
	"github.com/n6g7/godns/proto"
)

var rootServers = map[string]net.IP{
	"a.root-servers.net": net.ParseIP("198.41.0.4"),
	"b.root-servers.net": net.ParseIP("199.9.14.201"),
	"c.root-servers.net": net.ParseIP("192.33.4.12"),
	"d.root-servers.net": net.ParseIP("199.7.91.13"),
	"e.root-servers.net": net.ParseIP("192.203.230.10"),
	"f.root-servers.net": net.ParseIP("192.5.5.241"),
	"g.root-servers.net": net.ParseIP("192.112.36.4"),
	"h.root-servers.net": net.ParseIP("198.97.190.53"),
	"i.root-servers.net": net.ParseIP("192.36.148.17"),
	"j.root-servers.net": net.ParseIP("192.58.128.30"),
	"k.root-servers.net": net.ParseIP("193.0.14.129"),
	"l.root-servers.net": net.ParseIP("199.7.83.42"),
	"m.root-servers.net": net.ParseIP("202.12.27.33"),
}

type RecursiveResolver struct {
}

func NewRecursiveResolver() *RecursiveResolver {
	return &RecursiveResolver{}
}

func (rr *RecursiveResolver) Resolve(q proto.Question) ([]*proto.ResourceRecord, error) {
	var nextNS net.IP = randmap.Val(rootServers).(net.IP)
	var response *proto.DNSMessage
	var err error

MAIN:
	for {
		cli := client.NewClient(nextNS)
		response, err = cli.Resolve(q)

		if err != nil {
			return nil, err
		}

		// We have an answer!
		if len(response.Answers) > 0 {
			return response.Answers, nil
		}

		// No answer, looking for next NS
		var potentialAuthorities []*proto.ResourceRecord

	AUTHORITIES:
		for i := range response.Authority {
			authority := response.Authority[i]

			// Check name matches
			if !reflect.DeepEqual(authority.Name, q.Name[len(q.Name)-len(authority.Name):]) {
				continue AUTHORITIES
			}

			// Check class and type
			if authority.Class != q.Class || authority.Type != proto.NS {
				continue AUTHORITIES
			}

			// This is a valid authority ...
			potentialAuthorities = append(potentialAuthorities, authority)

			// Check ip is glued in additiooal
		ADDITIONALS:
			for j := range response.Additional {
				additional := response.Additional[j]

				// Check name matches
				if !reflect.DeepEqual(authority.DomainTarget, additional.Name) {
					continue ADDITIONALS
				}

				// Check class and type
				if additional.Class != q.Class || additional.Type != proto.A {
					continue ADDITIONALS
				}

				// We have glued IP!
				// Let's recurse from here
				nextNS = additional.IPTarget
				continue MAIN
			}

			// Couldn't find a glued IP for this authority ...
		}

		// We couldn't find any NS with glued IP

		// Can we resolve one of our potential NS domain names?
	RESOLVE_POTENTIAL:
		for i := range potentialAuthorities {
			potentialAuthority := potentialAuthorities[i]
			rq := proto.Question{
				Name:  potentialAuthority.DomainTarget,
				Class: q.Class,
				Type:  proto.A,
			}
			answers, err := rr.Resolve(rq)

			if err != nil {
				// Ignore for now, let's try the next potential NS
				continue RESOLVE_POTENTIAL
			}

		POTENTIAL_ANSWERS:
			for j := range answers {
				answer := answers[j]

				// Check answer name
				if !reflect.DeepEqual(answer.Name, potentialAuthority.DomainTarget) {
					continue POTENTIAL_ANSWERS
				}

				// Check class and type
				if answer.Class != q.Class || answer.Type != proto.A {
					continue POTENTIAL_ANSWERS
				}

				// We have resolved a potential NS domain:
				nextNS = answer.IPTarget
				continue MAIN
			}

			// Couldn't resolve this potential NS
		}

		// Couldn't resolve any of the potential NS domains, we're truly stuck here.
		return nil, fmt.Errorf("Couldn't find a nameserver IP for %s when asking %s", q.Name, nextNS)
	}
}
