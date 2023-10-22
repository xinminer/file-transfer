package cli

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func Parse() (localIp string, consulIp string, consulPort int, path string, suffix string, parallel int) {
	// Create options
	flag.StringVar(&localIp, "local-ip", "", "Server address")
	flag.StringVar(&consulIp, "consul-ip", "", "Consul ip")
	flag.IntVar(&consulPort, "consul-port", 0, "Consul port")
	flag.StringVar(&path, "path", "", "Transfer file path")
	flag.StringVar(&suffix, "suffix", "", "File extension")
	flag.IntVar(&parallel, "parallel", 10, "Send file parallel")

	// Parse
	flag.Parse()

	// Check required options
	seen := make(map[string]bool)
	flag.Visit(func(flag *flag.Flag) {
		seen[flag.Name] = true
	})
	if !seen["consul-ip"] || !seen["consul-port"] {
		fmt.Println("Missing required flags: -consul-ip, -consul-port")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Validate options data
	validateAddress(consulIp)
	validatePort(consulPort)

	return
}

func validateAddress(addr string) {
	_, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		fmt.Println(addr, "is not valid IP address")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func validatePort(port int) {
	if port < 0 || port > 65535 {
		fmt.Println(port, "is not valid port")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
