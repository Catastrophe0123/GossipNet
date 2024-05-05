package dns

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/catastrophe0123/gossipnet/delegate"
	"github.com/miekg/dns"
)

type DNS struct {
	registry     *delegate.Registry
	nameservers  []string
	resolverPath string
}

const DEFAULT_RESOLVER_PATH = "/etc/resolv.conf"

func NewDNS(registry *delegate.Registry, resolverPath string) (*DNS, error) {
	if resolverPath == "" {
		resolverPath = DEFAULT_RESOLVER_PATH
	}

	nameservers, err := getNameservers(resolverPath)
	fmt.Printf("nameservers: %v\n", nameservers)
	if err != nil {
		return nil, err
	}

	return &DNS{registry: registry, nameservers: nameservers, resolverPath: resolverPath}, nil
}

// function to read resolv.conf and extract nameservers
func getNameservers(resolverPath string) ([]string, error) {
	var nameservers []string

	file, err := os.Open(resolverPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			nameservers = append(nameservers, fields[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nameservers, nil
}

// function to add a nameserver to resolv.conf
func AddNameserver(nameserver string) error {
	file, err := os.OpenFile("/etc/resolv.conf", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	existing, err := nameserverExists(nameserver)
	fmt.Printf("existing: %v\n", existing)
	if err != nil {
		return err
	}
	if existing {
		return nil
	}

	_, err = file.WriteString("nameserver " + nameserver + "\n")
	if err != nil {
		return err
	}

	return nil
}

func AddNameserverToTop(resolverPath string, nameserver string) error {
	existingContent, err := os.ReadFile(resolverPath)
	if err != nil {
		return err
	}

	newContent := "nameserver " + nameserver + "\n" + string(existingContent)

	err = os.WriteFile(resolverPath, []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

// function to check if a nameserver already exists in resolv.conf
func nameserverExists(nameserver string) (bool, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver "+nameserver+"\n") {
			// if strings.HasPrefix(line, "nameserver "+nameserver) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func (d *DNS) SetupDNSServer(serverAddr string) (*dns.Server, error) {

	if serverAddr == "" {
		serverAddr = "127.0.0.1:5353"
	}

	server := &dns.Server{Addr: serverAddr, Net: "udp"}

	if strings.Contains(serverAddr, ":") {
		serverAddr = strings.Split(serverAddr, ":")[0]
	}
	fmt.Println("serveraddrr ; ", serverAddr)

	// err := AddNameserver(serverAddr)
	err := AddNameserverToTop(d.resolverPath, serverAddr)
	if err != nil {
		return nil, err
	}

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
				// resp, err := dns.Exchange(r, "8.8.8.8:53")
				resp, err := dns.Exchange(r, d.nameservers[0]+":53")
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
