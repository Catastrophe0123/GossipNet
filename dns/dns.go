package dns

import (
	"fmt"
	"strings"

	"github.com/catastrophe0123/gossipnet/delegate"
	"github.com/miekg/dns"
)

type DNS struct {
	registry *delegate.Registry
}

func NewDNS(registry *delegate.Registry) *DNS {
	return &DNS{registry: registry}
}

func (d *DNS) SetupDNSServer(serverAddr string) (*dns.Server, error) {

	if serverAddr == "" {
		serverAddr = "127.0.0.1:5353"
	}

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
				resp, err := dns.Exchange(r, "8.8.8.8:53")
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
	serviceName := strings.TrimSuffix(domain, ".")
	addr, err := d.registry.GetServiceAddress(serviceName)
	fmt.Printf("addr: %v\n", addr)
	fmt.Printf("err: %v\n", err)
	if err != nil {
		fmt.Printf("err getting server addresg: %v\n", err)
		return "", false
	}
	if addr != "" {
		return addr, true
	}

	return "", false
}
