package discovery

import (
	"fmt"
	"log"
	"time"

	"github.com/catastrophe0123/gossipnet/delegate"
	"github.com/catastrophe0123/gossipnet/dns"
	"github.com/hashicorp/memberlist"
)

type ServiceDiscovery struct {
	List     *memberlist.Memberlist
	Registry *delegate.Registry
	DNS      *dns.DNS
}

func NewServiceDiscovery() *ServiceDiscovery {
	return &ServiceDiscovery{}
}

func (s *ServiceDiscovery) GetAddressOfNode(nodeId string) string {
	members := s.List.Members()
	fmt.Printf("members: %v\n", members)
	for _, node := range members {
		if node.Name == nodeId {
			return node.Addr.String()
		}
	}

	return ""
}

func (s *ServiceDiscovery) InitGossip(
	config *memberlist.Config,
	peerAddr string,
) (*memberlist.Memberlist, error) {
	list, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}

	if peerAddr != "" {
		time.Sleep(1000)
		_, err := list.Join([]string{peerAddr})
		if err != nil {
			log.Println("Failed to join cluster: ", err)
		}
	}

	s.List = list

	return list, err
}
