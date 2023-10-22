package server

import (
	"file-transfer/internal/consul"
	"fmt"
	"github.com/gogf/gf/v2/util/guid"
	"net"
	"time"

	"file-transfer/internal/log"
)

const (
	clientTimeout = time.Minute
)

var store *storage

func Start(serverAddr *net.TCPAddr, consulIp string, consulPort int, tag string, destinations []string) {

	// Configuration file storage location
	store = newStorage()
	for _, destination := range destinations {
		if err := store.addPath(destination); err != nil {
			log.Log.Warnf("Configuration file storage location error: %v", err)
		}
	}

	// Create control connection listener
	listener, err := net.ListenTCP("tcp", serverAddr)
	if err != nil {
		log.Log.Errorf("Control connection listener creation error: %v", err)
	}
	log.Log.Infof("Listen to %v for control connection", serverAddr)
	defer func() {
		if err := listener.Close(); err != nil {
			log.Log.Errorf("Control connection listener closing error: %v", err)
			return
		}
		log.Log.Debugf("Closed control connection listener")
	}()

	// Service discovery
	dis := consul.DiscoveryConfig{
		ID:      guid.S(),
		Name:    "file-server",
		Tags:    []string{tag},
		Port:    serverAddr.Port,
		Address: serverAddr.IP.String(),
	}

	if err := consul.RegisterService(fmt.Sprintf("%s:%d", consulIp, consulPort), dis); err != nil {
		log.Log.Errorf("Connect consul error: %v", err)
		return
	}

	// Accepting control connection
	for {
		controlConn, err := listener.AcceptTCP()
		if err != nil {
			log.Log.Errorf("Control connection accepting error: %v", err)
		}
		log.Log.Debugf("Created control connection from %v", controlConn.RemoteAddr())

		go func() {
			handleRequests(controlConn)
		}()
	}
}
