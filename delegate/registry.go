package delegate

import (
	"fmt"
	"sync"
)

type NodeAddressManager interface {
	GetAddressOfNode(nodeID string) string
}

type Registry struct {
	mu                 sync.RWMutex
	Nodes              map[string]*[]Service
	serviceIndexes     map[string]int
	NodeAddressManager NodeAddressManager
}

func NewRegistry(nam NodeAddressManager) *Registry {
	return &Registry{
		Nodes:              make(map[string]*[]Service),
		serviceIndexes:     make(map[string]int),
		NodeAddressManager: nam,
	}
}

func (r *Registry) UpdateRegistry(updates *NodeServices) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Nodes[updates.NodeID] = updates.Services
}

func (r *Registry) Print() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for NodeID, services := range r.Nodes {
		fmt.Printf("NodeID: %s, Services: %v\n", NodeID, services)
	}
}

func (r *Registry) RemoveNode(nodeId string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Nodes, nodeId)
}

func (r *Registry) GetServicesFromNode(nodeId string) (*[]Service, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	services, ok := r.Nodes[nodeId]
	return services, ok
}

func (r *Registry) GetServiceAddress(serviceName string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodeCount := len(r.Nodes)
	if nodeCount == 0 {
		return "", fmt.Errorf("no nodes available")
	}

	r.serviceIndexes[serviceName] = (r.serviceIndexes[serviceName] + 1) % nodeCount
	selectedIndex := r.serviceIndexes[serviceName]

	var selectedAddress string
	var found bool

	for nodeId, services := range r.Nodes {
		for _, service := range *services {
			if service.Name == serviceName {
				if selectedIndex == 0 {
					address := r.NodeAddressManager.GetAddressOfNode(nodeId)
					selectedAddress = address
					found = true
					break
				}
				selectedIndex--
			}
		}
		if found {
			break
		}
	}

	if !found {
		return "", fmt.Errorf("service %s not found", serviceName)
	}

	return selectedAddress, nil
}
