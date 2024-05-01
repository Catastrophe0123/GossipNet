package delegate

import (
	"encoding/json"
	"fmt"
)

type Service struct {
	Name string
}

type NodeServices struct {
	NodeID   string
	Services *[]Service
}

type SDDelegate struct {
	LocalServices *NodeServices
	Registry      *Registry
}

func NewSDDelegate(registry *Registry) *SDDelegate {
	return &SDDelegate{
		LocalServices: &NodeServices{},
		Registry:      registry,
	}
}

func (d *SDDelegate) SetLocalServices(services *NodeServices) {
	d.LocalServices = services
}

func (d *SDDelegate) NodeMeta(limit int) []byte {
	fmt.Println("NodeMeta")
	data, _ := json.Marshal(d.LocalServices)
	return data
}

func (d *SDDelegate) NotifyMsg(b []byte) {
	// fmt.Println("NotifyMsg")
	var updates NodeServices
	if err := json.Unmarshal(b, &updates); err == nil {
		// fmt.Printf("updates: %v\n", updates)
		d.Registry.UpdateRegistry(&updates)
	}
}

func (d *SDDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	// fmt.Println("getbroadcapsts")
	data, _ := json.Marshal(d.LocalServices)
	return [][]byte{data}
}

func (d *SDDelegate) MergeRemoteState(buf []byte, join bool) {
	var receivedServices NodeServices
	fmt.Println("merge remotestate")
	if err := json.Unmarshal(buf, &receivedServices); err == nil {
		fmt.Printf("receivedServices: %v\n", receivedServices)
		d.Registry.UpdateRegistry(&receivedServices)
		return
	}
}

func (d *SDDelegate) LocalState(join bool) []byte {
	fmt.Println("LocalState")
	data, _ := json.Marshal(d.LocalServices)
	return data
}
