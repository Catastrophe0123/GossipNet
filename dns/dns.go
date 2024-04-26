package dns

import (
	"fmt"

	"github.com/catastrophe0123/gossipnet/delegate"
	"github.com/hashicorp/memberlist"
	"github.com/miekg/dns"
)

type DNS struct {
	Registry   *delegate.ServicesRegistry
	memberlist *memberlist.Memberlist
}

func NewDNS(registry *delegate.ServicesRegistry, ml *memberlist.Memberlist) *DNS {
	return &DNS{Registry: registry, memberlist: ml}
}

func (d *DNS) SetupDNSServer() (*dns.Server, error) {

	serverAddr := "127.0.0.1:5354"

	server := &dns.Server{Addr: serverAddr, Net: "udp"}

	server.Handler = dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Compress = false

		for _, q := range m.Question {
			fmt.Printf("Received query for %s\n", q.Name)
			ip, found := d.lookupDNS(q.Name)
			if found {
				fmt.Println("found : ", ip)
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err != nil {
					fmt.Println("Error creating DNS response:", err)
					return
				}
				m.Answer = append(m.Answer, rr)
			} else {
				resp, err := dns.Exchange(r, "")
				if err != nil {
					fmt.Println("Error forwarding query:", err)
					return
				}
				m.Answer = resp.Answer
			}
		}

		w.WriteMsg(m)
	})
	return server, nil
}

func (d *DNS) lookupDNS(domain string) (string, bool) {
	fmt.Printf("d.Registry.Nodes: %v\n", d.Registry.Nodes)
	for nodeName, services := range d.Registry.Nodes {
		for _, service := range services {
			if service.Name+"." == domain {
				members := d.memberlist.Members()
				fmt.Printf("members: %v\n", members)
				for _, node := range members {
					if node.Name == nodeName {
						// return node.Addr.String() + ":" + strconv.FormatInt(
						// 	int64(service.Port),
						// 	10,
						// ), true
						return node.Addr.String(), true
					}
				}
				// return service.Host, true
				return "", false
			}
		}
	}

	return "", false
}
