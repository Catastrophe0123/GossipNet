package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/catastrophe0123/gossipnet/delegate"
	"github.com/hashicorp/memberlist"
)

func main() {

	nodeName := flag.String("name", "", "Node name")
	bindPort := flag.String("bind", "", "Bind port")
	peerAddr := flag.String("peer", "", "Peer address")
	flag.Parse()

	fmt.Println("nodename :", *nodeName, *bindPort, *peerAddr)
	config := memberlist.DefaultLocalConfig()
	config.Name = *nodeName
	if len(os.Args) > 1 {
		config.BindPort = parseArg(*bindPort)
	}

	globalRegistry := &delegate.ServicesRegistry{Nodes: make(map[string][]delegate.Service)}
	localServices := &delegate.NodeServices{
		NodeID: config.Name,
		Services: []delegate.Service{
			{"MyService" + config.Name, "localhost", 8080, "active"},
		},
	}
	globalRegistry.Nodes[config.Name] = localServices.Services

	config.Events = &delegate.EventDelegate{GlobalRegistry: globalRegistry}

	config.Delegate = &delegate.CustomDelegate{
		LocalServices:  localServices,
		GlobalRegistry: globalRegistry,
	}

	list, err := memberlist.Create(config)
	if err != nil {
		log.Fatal("Failed to create memberlist: ", err)
	}

	if peerAddr != nil {
		_, err := list.Join([]string{*peerAddr})
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
