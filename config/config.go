package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/catastrophe0123/gossipnet/delegate"
)

type Config struct {
	Services []delegate.Service
}

func ParseConfig(configFile string) (*Config, error) {
	var config Config
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &config)
	fmt.Printf("config: %v\n", config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
