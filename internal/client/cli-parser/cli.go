package cliparser

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func Parse() (addr string, port int, filePath string) {
	// Create options
	flag.StringVar(&addr, "address", "", "Server address")
	flag.IntVar(&port, "port", 0, "Server port")
	flag.StringVar(&filePath, "file", "", "Transfer file path")

	// Parse
	flag.Parse()

	// Check required options
	seen := make(map[string]bool)
	flag.Visit(func(flag *flag.Flag) {
		seen[flag.Name] = true
	})
	if !seen["address"] || !seen["port"] || !seen["file"] {
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
	_, err := net.LookupHost(addr)
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
