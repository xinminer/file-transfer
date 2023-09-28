package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/robfig/cron/v3"

	"file-transfer/internal/core"
)

func transferFile(transferAddr *net.TCPAddr, session *session) {
	// Create transfer connection
	transferConn, err := net.DialTCP("tcp", nil, transferAddr)
	if err != nil {
		core.Log.Errorf("Transfer connection creation error: %v", err)
		sendResponse("error", "Failed to create transfer connection", session.encoder)
		return
	}
	core.Log.Infof("Created transfer connection to %v", transferAddr)
	defer func() {
		if err := transferConn.Close(); err != nil {
			core.Log.Errorf("Transfer connection closing error: %v", err)
			return
		}
		core.Log.Infof("Closed transfer connection to %v", transferAddr)
	}()

	// Open file
	file, err := os.OpenFile(fmt.Sprintf("%s/%s", uploadDir, session.fileName), os.O_RDWR, 0)
	if err != nil {
		core.Log.Errorf("File opening error: %v", err)
		sendResponse("error", "Internal server error", session.encoder)
		return
	}
	core.Log.Infof("Opened file %s/%s", uploadDir, session.fileName)
	defer func() {
		if err := file.Close(); err != nil {
			core.Log.Errorf("File closing error: %v", err)
			return
		}
		core.Log.Infof("Closed file %s/%s", uploadDir, session.fileName)
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

	core.Log.Infof("Start file transfer")
	for session.actualFileSize < session.expectedFileSize {
		received, err := transferConn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			core.Log.Errorf("File data reading error: %v", err)
			sendResponse("error", "Failed to read file data", session.encoder)
			return
		}

		_, err = multiWriter.Write(buffer[:received])
		if err != nil {
			core.Log.Errorf("File writing error: %v", err)
			sendResponse("error", "Internal server error", session.encoder)
			return
		}

		session.actualFileSize += int64(received)
	}

	// Stop scheduler and execute last task
	scheduler.Stop()
	speedCounter.calcSpeed(session)

	core.Log.Infof("Finish file transfer")

	// Update session
	session.fileHashSum = hex.EncodeToString(fileHashSum.Sum(nil))
}
