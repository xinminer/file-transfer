package balancer

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"math/rand"
)

func RoundRobin(consulAddr string, index *int, service, tags string) (string, error) {
	defer func() {
		*index = *index + 1
	}()
	target, err := getConsulServices(consulAddr, service)
	if err != nil {
		return "", err
	}
	if len(target) <= *index {
		*index = 0
	}

	return fmt.Sprintf("%s:%d", target[*index].Service.Address, target[*index].Service.Port), nil
}

func Random(consulAddr, service, tags string) (string, error) {
	target, err := getConsulServices(consulAddr, service)
	if err != nil {
		return "", err
	}
	lens := len(target)
	index := rand.Intn(lens)
	for i := 0; i < 3; i++ {
		svr := target[index]
		if svr.Service.Tags[0] == tags {
			return fmt.Sprintf("%s:%d", target[index].Service.Address, target[index].Service.Port), nil
		}
		index = rand.Intn(lens)
	}
	return fmt.Sprintf("%s:%d", target[index].Service.Address, target[index].Service.Port), nil
}

func getConsulServices(consulAddr, service string) ([]*consulapi.ServiceEntry, error) {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	services, _, err := client.Health().Service(service, "", false, nil)
	return services, err
}
