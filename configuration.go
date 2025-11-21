package main

import (
	"os"
	"time"
	"encoding/json"
)

type Config struct {
    LocalPlayer			int  `json:"localPlayer"`
    MoveSpeed			int  `json:"moveSpeed"`
    SimulationSpeedMS	int  `json:"simulationSpeedMS"`
	InputDelayFrames	int  `json:"inputDelayFrames"`
    Online				bool `json:"online"`
}

func loadConfig(filename string) {
    var config Config

    data, err := os.ReadFile(filename)
	F(err, "Error: could not load config - loaded defaults")
	if err != nil {
		config.LocalPlayer = 1
		config.MoveSpeed = 8
		config.SimulationSpeedMS = 16
		config.Online = false
	}
    
    err = json.Unmarshal(data, &config)
	F(err, "Error: could not parse config.json - loaded defaults")
	if err != nil {
		config.LocalPlayer = 1
		config.MoveSpeed = 8
		config.SimulationSpeedMS = 16
		config.Online = false
	}

	LOCAL = config.LocalPlayer
	if LOCAL == 1 {
		PEER = 2
		player1.isLocal = true
		player2.isLocal = false
		localPlayerPtr = &player1
		peerPlayerPtr  = &player2
	}
	if LOCAL == 2 {
		PEER = 1
		player1.isLocal = false
		player2.isLocal = true
		localPlayerPtr = &player2
		peerPlayerPtr  = &player1
	}

	SIM_TIME = time.Millisecond * time.Duration(config.SimulationSpeedMS)

	player1.scpt = int16(config.MoveSpeed)
	player2.scpt = int16(config.MoveSpeed)

	PACKET_BUFFER_LEN = config.InputDelayFrames
	online = config.Online
}
