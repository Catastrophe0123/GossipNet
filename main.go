package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/catastrophe0123/gossipnet/delegate"
	"github.com/hashicorp/memberlist"
)

func main() {
	config := memberlist.DefaultLocalConfig()
	config.Name = os.Args[len(os.Args)-1]
	if len(os.Args) > 1 {
		config.BindPort = parseArg(os.Args[1])
	}

	globalRegistry := &delegate.ServicesRegistry{Nodes: make(map[string][]delegate.Service)}
	localServices := &delegate.NodeServices{
		NodeID: config.Name,
		Services: []delegate.Service{
			{"MyService" + config.Name, "localhost", 8080, "active"},
		},
	}
	globalRegistry.Nodes[config.Name] = localServices.Services

	config.Delegate = &delegate.CustomDelegate{
		LocalServices:  localServices,
		GlobalRegistry: globalRegistry,
	}

	list, err := memberlist.Create(config)
	if err != nil {
		log.Fatal("Failed to create memberlist: ", err)
	}

	if len(os.Args) > 2 {
		_, err := list.Join([]string{os.Args[2]})
		if err != nil {
			log.Println("Failed to join cluster: ", err)
		}
	}
	defer list.Shutdown()
	for {
		time.Sleep(1000 * time.Millisecond)

		// for _, member := range list.Members() {
		// 	fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
		// }
		fmt.Printf("globalRegistry: %v\n", globalRegistry)
	}

	// Block forever
	select {}
}

func parseArg(arg string) int {
	var port int
	_, err := fmt.Sscanf(arg, "%d", &port)
	if err != nil {
		log.Fatalf("Invalid port argument: %v", err)
	}
	return port
}
