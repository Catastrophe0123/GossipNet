package delegate

import (
	"fmt"

	"github.com/hashicorp/memberlist"
)

type EventDelegate struct {
	GlobalRegistry *ServicesRegistry
}

func (e *EventDelegate) NotifyJoin(*memberlist.Node) {

}

func (e *EventDelegate) NotifyLeave(node *memberlist.Node) {
	fmt.Printf("member left : %v", node)
	delete(e.GlobalRegistry.Nodes, node.Name)
}

func (e *EventDelegate) NotifyUpdate(*memberlist.Node) {

}
