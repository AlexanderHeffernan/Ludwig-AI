package main

import (
	"ludwig/internal/orchestrator"
	"time"
)

func main() {
	orchestrator.Start()
	time.Sleep(100 * time.Second)
	orchestrator.Stop()
}
