package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/catastrophe0123/gossipnet/delegate"
	"github.com/hashicorp/memberlist"
)

func main() {

	nodeName := flag.String("name", "", "Node name")
	bindPort := flag.String("bind", "", "Bind port")
	peerAddr := flag.String("peer", "", "Peer address")
	flag.Parse()

	config := memberlist.DefaultLocalConfig()

	config.Name = *nodeName
	if *bindPort != "" {
		config.BindPort = parseArg(*bindPort)
		config.AdvertisePort = parseArg(*bindPort)
	}

	config.BindAddr = "127.0.0.1"
	config.AdvertiseAddr = "127.0.0.1"
	fmt.Printf("config.AdvertiseAddr: %v\n", config.AdvertiseAddr)

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block forever
loop:
	for {
		select {
		case <-sigChan:
			fmt.Println("shutting down")
			err = list.Shutdown()
			fmt.Printf("err: %v\n", err)
			break loop
		default:
			time.Sleep(1000 * time.Millisecond)

			fmt.Printf("globalRegistry: %v\n", globalRegistry)
			fmt.Printf("globalRegistryppp: %p\n", globalRegistry)
		}
	}

}

func parseArg(arg string) int {
	var port int
	_, err := fmt.Sscanf(arg, "%d", &port)
	if err != nil {
		log.Fatalf("Invalid port argument: %v", err)
	}
	return port
}
