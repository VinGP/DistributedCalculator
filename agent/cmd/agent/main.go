package main

import (
	"github.com/vingp/DistributedCalculator/agent/config"
	"github.com/vingp/DistributedCalculator/agent/internal/app/agent"
)

func main() {

	cfg := config.Get()
	agent.Run(cfg)
}
