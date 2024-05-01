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
	"github.com/catastrophe0123/gossipnet/discovery"
	"github.com/catastrophe0123/gossipnet/dns"
	"github.com/hashicorp/memberlist"
)

type ApplicationConfig struct {
	NodeName     string
	BindPort     string
	ProbeTimeout time.Duration
	PeerAddr     string
	BindAddr     string
	DnsAddr      string
}

func getMemberlistConfig(appConfig *ApplicationConfig) *memberlist.Config {
	config := memberlist.DefaultLocalConfig()
	config.Name = appConfig.NodeName
	if appConfig.BindPort != "" {
		config.BindPort = parseArg(appConfig.BindPort)
		config.AdvertisePort = parseArg(appConfig.BindPort)
	}

	if appConfig.BindAddr == "" {
		config.BindAddr = "0.0.0.0"
	} else {
		config.BindAddr = appConfig.BindAddr
	}

	if appConfig.ProbeTimeout != 0 {
		config.ProbeTimeout = appConfig.ProbeTimeout
	} else {
		config.ProbeTimeout = 30 * time.Second
	}

	return config
}

func main() {
	nodeName := flag.String("name", "", "Node name")
	bindPort := flag.String("bind", "", "Bind port")
	peerAddr := flag.String("peer", "", "Peer address")
	dnsAddr := flag.String("dnsAddr", "", "dns address")
	configFile := flag.String("config-file", "config.json", "configuration file")
	flag.Parse()

	appConfig := ApplicationConfig{
		NodeName:     *nodeName,
		BindPort:     *bindPort,
		PeerAddr:     *peerAddr,
		ProbeTimeout: 30 * time.Second,
		BindAddr:     "0.0.0.0",
		DnsAddr:      *dnsAddr,
	}

	configFilePath, err := filepath.Abs(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	serviceConfig, err := config.ParseConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	serviceDiscovery := discovery.NewServiceDiscovery()
	registry := delegate.NewRegistry(serviceDiscovery)

	serviceDiscovery.DNS = dns.NewDNS(registry)
	dnsServer, err := serviceDiscovery.DNS.SetupDNSServer(appConfig.DnsAddr)
	if err != nil {
		log.Fatal("Failed to initialize DNS server : ", err)
	}

	go (func() {
		fmt.Println("starting DNS server")
		if err := dnsServer.ListenAndServe(); err != nil {
			log.Fatal("failed to start dns server : ", err)
		}
	})()

	memberlistConf := getMemberlistConfig(&appConfig)

	serviceDiscovery.Registry = registry

	sdDelegate := delegate.NewSDDelegate(registry)
	eventDelegate := delegate.NewEventDelegate(registry)

	memberlistConf.Delegate = sdDelegate
	memberlistConf.Events = eventDelegate

	localServices := &delegate.NodeServices{}
	localServices.NodeID = memberlistConf.Name
	localServices.Services = serviceConfig.Services

	sdDelegate.SetLocalServices(localServices)
	registry.UpdateRegistry(localServices)

	list, err := serviceDiscovery.InitGossip(memberlistConf, appConfig.PeerAddr)
	if err != nil {
		log.Fatal(err)
	}

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
			// registry.Print()
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
