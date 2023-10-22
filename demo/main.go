package main

import (
	"file-transfer/internal/balancer"
	"file-transfer/internal/log"
	"fmt"
)

func main() {
	var index int
	s, err := balancer.RoundRobin(fmt.Sprintf("%s:%d", "10.0.8.10", 8500), &index, "file-server", "")
	if err != nil {
		log.Log.Errorf("Discovery service error: %v", err)
		return
	}
	fmt.Println(s)
}
