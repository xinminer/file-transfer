package cli

import (
	"flag"
	"fmt"
	"os"
)

func Parse() (svrIp string, svrPort int, consulIp string, consulPort int, destinations []string) {
	// Create options
	flag.StringVar(&svrIp, "server-ip", "", "Server ip")
	flag.IntVar(&svrPort, "server-port", 0, "Server port")
	flag.StringVar(&consulIp, "consul-ip", "", "Consul ip")
	flag.IntVar(&consulPort, "consul-port", 0, "Consul port")

	// Parse
	flag.Parse()

	// Check required options
	seen := make(map[string]bool)
	flag.Visit(func(flag *flag.Flag) {
		seen[flag.Name] = true
	})
	if !seen["svrPort"] {
		fmt.Println("Missing required flags: -svrPort")
		flag.PrintDefaults()
		os.Exit(1)
	}

	destinations = flag.Args()

	// Validate options data
	validatePort(svrPort)
	validateDestinations(destinations)

	return
}

func validatePort(port int) {
	if port < 0 || port > 65535 {
		fmt.Println(port, "is not valid port")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func validateDestinations(destinations []string) {
	if len(destinations) == 0 {
		fmt.Println("the file save location is empty")
		flag.PrintDefaults()
		os.Exit(1)
	}

	for _, destination := range destinations {
		dir, err := os.Stat(destination)
		if err != nil {
			fmt.Println(err.Error())
			flag.PrintDefaults()
			os.Exit(1)
		}
		if !dir.IsDir() {
			fmt.Println(destination, "is not a folder")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}
}
