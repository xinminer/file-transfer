package main

import (
	"file-transfer/internal/consul"
	"file-transfer/internal/log"
	"fmt"
)

func main() {
	service, err := consul.Discovery("file-server", fmt.Sprintf("%s:%d", "10.0.8.10", 8500), "n14", 10)
	if err != nil {
		log.Log.Errorf("Discovery service error: %v", err)
		return
	}
	fmt.Println(service.Service.Address)
}
