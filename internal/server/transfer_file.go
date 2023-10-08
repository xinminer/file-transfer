package server

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"os"
	"time"

	"github.com/robfig/cron/v3"

	"file-transfer/internal/log"
)

func transferFile(transferAddr *net.TCPAddr, session *session) {
	// Create transfer connection
	transferConn, err := net.DialTCP("tcp", nil, transferAddr)
	if err != nil {
		log.Log.Errorf("Transfer connection creation error: %v", err)
		sendResponse("error", "Failed to create transfer connection", session)
		return
	}
	log.Log.Debugf("Created transfer connection to %v", transferAddr)
	defer func() {
		if err := transferConn.Close(); err != nil {
			log.Log.Errorf("Transfer connection closing error: %v", err)
			return
		}
		log.Log.Debugf("Closed transfer connection to %v", transferAddr)
	}()

	// Open file
	file, err := os.OpenFile(session.filePath, os.O_RDWR, 0)
	if err != nil {
		log.Log.Errorf("File opening error: %v", err)
		sendResponse("error", "Internal server error", session)
		return
	}
	log.Log.Debugf("Opened file %s", session.filePath)
	defer func() {
		if err := file.Close(); err != nil {
			log.Log.Errorf("File closing error: %v", err)
			return
		}
		log.Log.Debugf("Closed file %s", session.filePath)
	}()

	// Start speed counter
	speedCounter := newSpeedCounter()
	scheduler := cron.New()
	scheduler.AddFunc("@every 1s", func() {
		speedCounter.calcSpeed(session)
	})
	scheduler.Start()

	// Transfer bytes
	buffer := make([]byte, bufferSize)
	fileHashSum := sha256.New()
	multiWriter := io.MultiWriter(file, fileHashSum)

	log.Log.Debugf("Start file transfer")
	for session.actualFileSize < session.expectedFileSize {
		transferConn.SetReadDeadline(time.Now().Add(clientTimeout))
		received, err := transferConn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Log.Errorf("File data reading error: %v", err)
			sendResponse("error", "Failed to read file data", session)
			return
		}

		_, err = multiWriter.Write(buffer[:received])
		if err != nil {
			log.Log.Errorf("File writing error: %v", err)
			sendResponse("error", "Internal server error", session)
			return
		}

		session.actualFileSize += int64(received)
	}

	// Stop scheduler and execute last task
	scheduler.Stop()
	speedCounter.calcSpeed(session)

	log.Log.Debugf("Finish file transfer")

	// Update session
	session.fileHashSum = hex.EncodeToString(fileHashSum.Sum(nil))
}
