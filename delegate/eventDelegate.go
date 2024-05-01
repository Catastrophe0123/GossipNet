package delegate

import (
	"fmt"

	"github.com/hashicorp/memberlist"
)

type EventDelegate struct {
	GlobalRegistry *Registry
}

func NewEventDelegate(registry *Registry) *EventDelegate {
	return &EventDelegate{
		GlobalRegistry: registry,
	}
}

func (e *EventDelegate) NotifyJoin(node *memberlist.Node) {
	fmt.Println("member joined :", node)
}

func (e *EventDelegate) NotifyLeave(node *memberlist.Node) {
	fmt.Printf("member left : %v", node)
	e.GlobalRegistry.RemoveNode(node.Name)
}

func (e *EventDelegate) NotifyUpdate(node *memberlist.Node) {
}
