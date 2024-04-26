package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/catastrophe0123/gossipnet/config"
	"github.com/catastrophe0123/gossipnet/delegate"
	"github.com/catastrophe0123/gossipnet/dns"
	"github.com/hashicorp/memberlist"
)

func main() {

	nodeName := flag.String("name", "", "Node name")
	bindPort := flag.String("bind", "", "Bind port")
	peerAddr := flag.String("peer", "", "Peer address")
	configFile := flag.String("config-file", "config.json", "configuration file")
	flag.Parse()

	configFilePath, err := filepath.Abs(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	appConfig, err := config.ParseConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	config := memberlist.DefaultLocalConfig()
	config.Name = *nodeName
	if *bindPort != "" {
		config.BindPort = parseArg(*bindPort)
		config.AdvertisePort = parseArg(*bindPort)
	}

	// config.BindAddr = "127.0.0.1"
	// config.AdvertiseAddr = "127.0.0.1"
	config.BindAddr = "0.0.0.0"
	config.ProbeTimeout = 30 * time.Second
	// config.AdvertiseAddr = "0.0.0.0"
	globalRegistry := &delegate.ServicesRegistry{Nodes: make(map[string][]delegate.Service)}
	localServices := &delegate.NodeServices{
		NodeID:   config.Name,
		Services: appConfig.Services,
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

	// init dns server
	DNS := dns.NewDNS(globalRegistry, list)
	dnsServer, err := DNS.SetupDNSServer()
	if err != nil {
		log.Fatal("Failed to initialize DNS server : ", err)
	}

	go (func() {
		fmt.Println("starting DNS server")
		if err := dnsServer.ListenAndServe(); err != nil {
			log.Fatal("failed to start dns server : ", err)
		}
	})()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			fmt.Println("shutting down")
			err = list.Shutdown()
			fmt.Printf("err: %v\n", err)
			dnsServer.ShutdownContext(ctx)
			return
		default:
			time.Sleep(1000 * time.Millisecond)
			fmt.Printf("globalRegistry: %v\n", globalRegistry)
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
