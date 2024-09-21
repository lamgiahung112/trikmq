package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type TrikMqConfig struct {
	Port int    `json:"port"`
}

func ReadConfig() *TrikMqConfig {
	filepath := fmt.Sprintf("%v/trikmq.config.json", os.Getenv("TRIKMQ_CONF_DIR"))
	file, err := os.Open(filepath)

	if err != nil {
		fmt.Println(fmt.Errorf("Error reading config file: %v", err))
		os.Exit(1)
	}

	defer file.Close()
	content := make([]byte, 1024)
	ReadLen, err := file.Read(content)
	if err != nil {
		fmt.Println(fmt.Errorf("Error reading config file: %v", err))
	}
	var config TrikMqConfig
	err = json.Unmarshal(content[:ReadLen], &config)
	if err != nil {
		fmt.Println(fmt.Errorf("Error parsing config file: %v", err))
	}
	return &config
}
