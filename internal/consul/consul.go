package consul

import (
	"errors"
	"file-transfer/internal/log"
	"fmt"
	"github.com/gogf/gf/v2/container/glist"
	"github.com/gogf/gf/v2/text/gstr"
	consulapi "github.com/hashicorp/consul/api"
	"math/rand"
)

var cache = glist.New()

type DiscoveryConfig struct {
	ID      string
	Name    string
	Tags    []string
	Port    int
	Address string
}

func RegisterService(addr string, dis DiscoveryConfig) error {
	config := consulapi.DefaultConfig()
	config.Address = addr
	client, err := consulapi.NewClient(config)
	if err != nil {
		fmt.Printf("create consul client : %v\n", err.Error())
		return err
	}
	registration := &consulapi.AgentServiceRegistration{
		ID:      dis.ID,
		Name:    dis.Name,
		Port:    dis.Port,
		Tags:    dis.Tags,
		Address: dis.Address,
	}

	check := &consulapi.AgentServiceCheck{}
	check.TCP = fmt.Sprintf("%s:%d", registration.Address, registration.Port)
	check.Timeout = "5s"
	check.Interval = "5s"
	check.DeregisterCriticalServiceAfter = "60s"
	registration.Check = check

	if err := client.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	return nil
}

func Discovery(serviceName string, consulAddr string, localIp string, parallel int) (*consulapi.ServiceEntry, error) {

	if cache.Size() > 0 {
		value, ok := cache.PopFront().(*consulapi.ServiceEntry)
		if !ok {
			log.Log.Errorf("Pop service entry error")
		}
		return value, nil
	}

	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	services, _, err := client.Health().Service(serviceName, "", false, nil)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, errors.New("not found service")
	}

	var preferred []*consulapi.ServiceEntry
	var candidate []*consulapi.ServiceEntry

	for _, service := range services {
		serviceAddress := service.Service.Address
		comp := gstr.Explode(".", localIp)
		comp = comp[:3]
		prefix := gstr.Implode(".", comp)
		if gstr.HasPrefix(serviceAddress, prefix) {
			preferred = append(preferred, service)
		} else {
			candidate = append(candidate, service)
		}

	}

	rand.Shuffle(len(preferred), func(i, j int) { preferred[i], preferred[j] = preferred[j], preferred[i] })
	rand.Shuffle(len(candidate), func(i, j int) { candidate[i], candidate[j] = candidate[j], candidate[i] })

	if len(preferred) < parallel {
		count := parallel - len(preferred)
		if len(candidate) < count {
			count = len(candidate)
		}
		for i := 0; i < count; i++ {
			preferred = append(preferred, candidate[i])
		}
	}

	if len(preferred) > parallel {
		preferred = preferred[:parallel]
	}

	for _, entry := range preferred {
		cache.PushBack(entry)
	}

	value, ok := cache.PopFront().(*consulapi.ServiceEntry)
	if !ok {
		log.Log.Errorf("Pop service entry error")
	}

	return value, nil
}
