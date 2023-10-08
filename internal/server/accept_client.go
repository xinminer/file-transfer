package server

import (
	"net"
	"os"
	"time"

	"file-transfer/internal/log"
)

const (
	uploadDir     = "uploads"
	clientTimeout = time.Minute
)

func Start(serverAddr *net.TCPAddr) {
	// Make upload directory
	if err := os.Mkdir(uploadDir, os.ModePerm); err != nil && os.IsNotExist(err) {
		log.Log.Errorf("Upload directory creation error: %v", err)
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

	// Accepting control connection
	for {
		controlConn, err := listener.AcceptTCP()
		if err != nil {
			log.Log.Errorf("Control connection accepting error: %v", err)
		}
		log.Log.Debugf("Created control connection from %v", controlConn.RemoteAddr())

		go handleRequests(controlConn)
	}
}
