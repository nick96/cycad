package consul

import (
	"fmt"
	"os"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
)

func RegisterService(serviceName string, port string) error {
	cfg := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(cfg)
	if err != nil {
		return err
	}

	registration := new(consulapi.AgentServiceRegistration)
	hostname, _ := os.Hostname()
	registration.ID = fmt.Sprintf("%s-%s", serviceName, hostname)
	registration.Name = serviceName
	registration.Address = hostname
	portNum, _ := strconv.Atoi(port)
	registration.Port = portNum
	return consul.Agent().ServiceRegister(registration)
}

func HealthCheckService(serviceName string) (bool, error) {
	cfg := consuleapi.DefaultConfig()
	consul, err := consulapi.NewClient(cfg)
	if err != nil {
		return false, err
	}

	healthChecks, err := consul.Agent().AgentHealthServiceByName(serviceName)
	if err != nil {
		return false, err
	}

	for _, healthCheck := range healthChecks {
		if healthCheck.AggregateStatus == "passing" {
			return true
		}
	}
	return false
}
