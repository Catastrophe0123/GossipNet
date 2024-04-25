package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/catastrophe0123/gossipnet/delegate"
)

type Config struct {
	Services []delegate.Service
}

func WatchConfigFile(configFile string, config *Config) {
	for {
		time.Sleep(5 * time.Second)
		err := readConfig(configFile, config)
		if err != nil {
			fmt.Println("error reading configuration file:", err)
		}
	}
}

func readConfig(configFile string, config *Config) error {
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &config)
	fmt.Printf("config: %v\n", config)
	if err != nil {
		return err
	}

	return nil
}

func ParseConfig(configFile string) (*Config, error) {
	var config Config
	err := readConfig(configFile, &config)
	if err != nil {
		return nil, err
	}
	go WatchConfigFile(configFile, &config)
	return &config, nil
}
