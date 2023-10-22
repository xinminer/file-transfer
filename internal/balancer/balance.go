package balancer

import (
	"file-transfer/internal/log"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"math/rand"
)

func RoundRobin(consulAddr string, index *int, service, tags string) (string, error) {
	defer func() {
		*index = *index + 1
	}()
	target, err := getConsulServices(consulAddr, service, tags)
	if err != nil {
		return "", err
	}
	if len(target) <= *index {
		*index = 0
	}

	return fmt.Sprintf("%s:%d", target[*index].Service.Address, target[*index].Service.Port), nil
}

func Random(consulAddr, service, tags string) (string, error) {
	target, err := getConsulServices(consulAddr, service, tags)
	if err != nil {
		return "", err
	}
	lens := len(target)
	index := rand.Intn(lens)
	return fmt.Sprintf("%s:%d", target[index].Service.Address, target[index].Service.Port), nil
}

func getConsulServices(consulAddr, service, tags string) ([]*consulapi.ServiceEntry, error) {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	services, _, err := client.Health().Service(service, "", false, nil)
	log.Log.Infof("services: %d", len(services))
	return services, err
}
