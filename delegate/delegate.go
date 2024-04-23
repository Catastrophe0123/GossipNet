package delegate

import (
	"encoding/json"
	"fmt"
)

type Service struct {
	Name   string
	Host   string
	Port   int
	Status string
}

type NodeServices struct {
	NodeID   string
	Services []Service
}

type ServicesRegistry struct {
	Nodes map[string][]Service
}

type CustomDelegate struct {
	LocalServices  *NodeServices
	GlobalRegistry *ServicesRegistry
}

func (d *CustomDelegate) NodeMeta(limit int) []byte {
	fmt.Println("NodeMeta")
	data, _ := json.Marshal(d.LocalServices)
	return data
}

func (d *CustomDelegate) NotifyMsg(b []byte) {
	// fmt.Println("NotifyMsg")
	var updates NodeServices
	if err := json.Unmarshal(b, &updates); err == nil {
		// fmt.Printf("updates: %v\n", updates)
		d.GlobalRegistry.Nodes[updates.NodeID] = updates.Services
	}
}

func (d *CustomDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	// fmt.Println("getbroadcapsts")
	data, _ := json.Marshal(d.LocalServices)
	return [][]byte{data}
}

// func (d *CustomDelegate) MergeRemoteState(buf []byte, join bool) {
// 	var receivedRegistry ServicesRegistry
// 	if err := json.Unmarshal(buf, &receivedRegistry); err == nil {
// 		for nodeID, services := range receivedRegistry.Nodes {
// 			d.globalRegistry.Nodes[nodeID] = services
// 		}
// 	}
// }

func (d *CustomDelegate) MergeRemoteState(buf []byte, join bool) {
	var receivedServices NodeServices
	fmt.Println("merge remotestate")
	if err := json.Unmarshal(buf, &receivedServices); err == nil {
		fmt.Printf("receivedServices: %v\n", receivedServices)
		currentServices, exists := d.GlobalRegistry.Nodes[receivedServices.NodeID]
		if !exists || join {
			d.GlobalRegistry.Nodes[receivedServices.NodeID] = receivedServices.Services
		} else {
			for _, receivedService := range receivedServices.Services {
				updated := false
				for i, currentService := range currentServices {
					if currentService.Name == receivedService.Name {
						currentServices[i] = receivedService
						updated = true
						break
					}
				}
				if !updated {
					currentServices = append(currentServices, receivedService)
				}
			}
			d.GlobalRegistry.Nodes[receivedServices.NodeID] = currentServices
		}
	}
}

func (d *CustomDelegate) LocalState(join bool) []byte {
	fmt.Println("LocalState")
	data, _ := json.Marshal(d.LocalServices)
	return data
}
