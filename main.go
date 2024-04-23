package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/memberlist"
)

func main() {
	// Create a configuration for memberlist
	config := memberlist.DefaultLocalConfig()
	config.Name = os.Args[len(os.Args)-1]
	// Allow passing the bind port via args for multiple instances
	if len(os.Args) > 1 {
		config.BindPort = parseArg(os.Args[1])
	}

	// Create a new memberlist using the configuration
	list, err := memberlist.Create(config)
	if err != nil {
		log.Fatal("Failed to create memberlist: ", err)
	}

	// Join an existing cluster by specifying at least one known member.
	if len(os.Args) > 2 {
		_, err := list.Join([]string{os.Args[2]})
		if err != nil {
			log.Println("Failed to join cluster: ", err)
		}
	}
	defer list.Shutdown()
	for {
		time.Sleep(1000 * time.Millisecond)
		// Output the members of the cluster
		for _, member := range list.Members() {
			fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
		}
		fmt.Println("all members listed")
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
