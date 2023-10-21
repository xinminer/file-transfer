package cli

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func Parse() (addr string, port int, path string, suffix string, parallel int) {
	// Create options
	flag.StringVar(&addr, "address", "", "Server address")
	flag.IntVar(&port, "port", 0, "Server port")
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
	if !seen["address"] || !seen["port"] {
		fmt.Println("Missing required flags: -address, -port, -file")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Validate options data
	validateAddress(addr)
	validatePort(port)

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
