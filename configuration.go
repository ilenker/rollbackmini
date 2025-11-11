package main

import (
	"os"
	"encoding/json"
)

type Config struct {
    LocalPlayer       int `json:"localPlayer"`
    MoveSpeed         int `json:"moveSpeed"`
    RollbackWindow    int `json:"rollbackWindow"`
    SimulationSpeedMS int `json:"simulationSpeedMS"`
}

func loadConfig(filename string) *Config {
    var config Config

    data, err := os.ReadFile(filename)
	E(err, "Error: could not load config - loaded defaults")
	if err != nil {
		config.LocalPlayer = 1
		config.MoveSpeed = 8
		config.RollbackWindow = 15
		config.SimulationSpeedMS = 16
		return &config
	}
    
    err = json.Unmarshal(data, &config)
	E(err, "Error: could not parse config.json - loaded defaults")
	if err != nil {
		config.LocalPlayer = 1
		config.MoveSpeed = 8
		config.RollbackWindow = 15
		config.SimulationSpeedMS = 16
		return &config
	}
	return &config
}
