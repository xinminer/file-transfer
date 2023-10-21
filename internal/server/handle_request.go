package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"file-transfer/internal/dto"
	"file-transfer/internal/log"
)

const (
	maxFileSize       = 1 << 40 // 1 TB
	maxFileNameLength = 4096
	bufferSize        = 1024
)

func handleRequests(controlConn *net.TCPConn) {
	defer func() {
		if err := controlConn.Close(); err != nil {
			log.Log.Errorf("Control connection closing error: %v", err)
			return
		}
		log.Log.Debugf("Closed control connection from %v", controlConn.RemoteAddr())
	}()

	// Create session
	session := newSession(controlConn)

	reader := bufio.NewReader(session.controlConn)
	var data map[string]interface{}
	for {
		// Receive JSON message
		session.controlConn.SetReadDeadline(time.Now().Add(clientTimeout))
		jsonBytes, err := reader.ReadBytes('\n')
		if err != nil {
			log.Log.Errorf("Request receiving error: %v", err)
			sendResponse("error", "Failed to read request", session)
			removeUntransferredFile(session)
			time.Sleep(2 * time.Second)
			return
		}

		// Convert JSON to object
		if err := json.Unmarshal(jsonBytes, &data); err != nil {
			log.Log.Errorf("Unmarshalling error: Expected message with JSON format: %v", err)
			sendResponse("error", "Expected message with JSON format", session)
			removeUntransferredFile(session)
			return
		}

		// Check type field
		if _, ok := data["type"]; !ok {
			log.Log.Errorf("Request conversion error: Field \"type\" is missing")
			sendResponse("error", "Field \"type\" is missing", session)
			removeUntransferredFile(session)
			return
		}

		// Handle data
		switch data["type"].(string) {
		case "file_info":
			if !handleFileInfoRequest(data, session) {
				removeUntransferredFile(session)
				return
			}
		case "start_transfer":
			if !handleStartTransferRequest(data, session) {
				removeUntransferredFile(session)
				return
			}
		case "end_transfer":
			if !handleEndTransferRequest(data, session) {
				removeUntransferredFile(session)
			}
			return
		default:
			handleUnsupportedRequest(data, session)
			removeUntransferredFile(session)
			return
		}
	}
}

func handleFileInfoRequest(data map[string]interface{}, session *session) bool {
	// Convert data to file_info request
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"file_info\" is not valid", session)
		return false
	}

	var fileInfoRequest dto.FileInfoRequest
	if err := json.Unmarshal(jsonBytes, &fileInfoRequest); err != nil {
		log.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"file_info\" is not valid", session)
		return false
	}

	// Check stage
	if session.stage != "file_info" {
		log.Log.Errorf("Stage error: Expected message \"%s\"", session.stage)
		sendResponse("error", fmt.Sprintf("Expected message \"%s\"", session.stage), session)
		return false
	}

	// Get storage path
	path, err := store.getPath(fileInfoRequest.FileSize)
	if err != nil {
		log.Log.Errorf("get storage path error: %v", err)
		sendResponse("error", err.Error(), session)
		return false
	}
	session.parentPath = path

	// Validate file info
	if err := validateFileInfo(fileInfoRequest, path); err != nil {
		log.Log.Errorf("Message \"file_info\" validation error: %v", err)
		sendResponse("error", err.Error(), session)
		return false
	}

	// Create file
	if _, err := os.Create(path + "/" + fileInfoRequest.FileName); err != nil {
		log.Log.Errorf("File creation error: %v", err)
		sendResponse("error", fmt.Sprintf("Failed to create file %s", fileInfoRequest.FileName), session)
		return false
	}
	log.Log.Debugf("Created file %s/%s", path, fileInfoRequest.FileName)
	session.fileIsCreated = true

	// Update session info
	session.filePath = path + "/" + fileInfoRequest.FileName
	session.expectedFileSize = fileInfoRequest.FileSize
	session.stage = "start_transfer"

	// Send success response
	sendResponse("success", "", session)

	return true
}

func handleStartTransferRequest(data map[string]interface{}, session *session) bool {
	// Convert data to start_transfer request
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"start_transfer\" is not valid", session)
		return false
	}

	var startTransferRequest dto.StartTransferRequest
	if err := json.Unmarshal(jsonBytes, &startTransferRequest); err != nil {
		log.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"start_transfer\" is not valid", session)
		return false
	}

	// Check stage
	if session.stage != "start_transfer" {
		log.Log.Errorf("Stage error: Expected message \"%s\"", session.stage)
		sendResponse("error", fmt.Sprintf("Expected message \"%s\"", session.stage), session)
		return false
	}

	// Transfer file
	transferAddr, err := net.ResolveTCPAddr("tcp",
		fmt.Sprintf("%s:%d", session.controlAddr.IP, startTransferRequest.Port))
	if err != nil {
		log.Log.Errorf("Transfer address resolving error: %v", err)
		sendResponse("error", "Internal server error", session)
		return false
	}
	transferFile(transferAddr, session)

	// Update session
	session.stage = "end_transfer"

	return true
}

func handleEndTransferRequest(data map[string]interface{}, session *session) bool {
	// Convert data to end_transfer request
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"end_transfer\" is not valid", session)
		return false
	}

	var endTransferRequest dto.EndTransferRequest
	if err := json.Unmarshal(jsonBytes, &endTransferRequest); err != nil {
		log.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"end_transfer\" is not valid", session)
		return false
	}

	// Check stage
	if session.stage != "end_transfer" {
		log.Log.Errorf("Stage error: Expected message \"%s\"", session.stage)
		sendResponse("error", fmt.Sprintf("Expected message \"%s\"", session.stage), session)
		return false
	}

	// Check transferred file
	if session.expectedFileSize != session.actualFileSize {
		log.Log.Errorf("Actual and expected file sizes is are different")
		sendResponse("error", "Actual and expected file sizes is are different", session)
		return false
	}
	if session.fileHashSum != endTransferRequest.FileHashSum {
		log.Log.Errorf("Actual and expected file hash sums is are different")
		sendResponse("error", "Actual and expected file hash sums is are different", session)
		return false
	}

	// Send success response
	sendResponse("success", "", session)
	log.Log.Infof("File %s transferred successfully", session.filePath)

	// Update session
	session.stage = ""

	return true
}

func handleUnsupportedRequest(data map[string]interface{}, session *session) {
	log.Log.Errorf("Request receiving error: Message type \"%s\" is unsupported",
		data["type"].(string))
	sendResponse("error",
		fmt.Sprintf("Message type \"%s\" is unsupported", data["type"].(string)),
		session)
}

func sendResponse(status string, message string, session *session) {
	// Convert response object to JSON
	jsonBytes, err := json.Marshal(dto.Response{Type: status, Message: message})
	if err != nil {
		log.Log.Errorf("Marshalling error: %v", err)
		return
	}

	// Send JSON
	session.controlConn.SetWriteDeadline(time.Now().Add(clientTimeout))
	_, err = session.controlConn.Write(jsonBytes)
	if err != nil {
		log.Log.Errorf("Response sending error: %v", err)
		return
	}
}

func validateFileInfo(fileInfo dto.FileInfoRequest, path string) error {
	// Check file name
	if len(fileInfo.FileName) == 0 {
		return errors.New("File name is empty")
	}
	if len(fileInfo.FileName) > maxFileNameLength {
		return errors.New("File name length exceeds 4096 bytes")
	}

	// Check file size
	if fileInfo.FileSize < 0 {
		return errors.New("File size is negative number")
	}
	if fileInfo.FileSize > maxFileSize {
		return errors.New("File size exceeds 1 TB")
	}

	// Check if file with same name exists
	if _, err := os.Stat(path + "/" + fileInfo.FileName); !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("File %s already exists", fileInfo.FileName))
	}

	return nil
}

func removeUntransferredFile(session *session) {
	if session.fileIsCreated {
		if err := os.Remove(session.filePath); err != nil {
			log.Log.Errorf("Untransferred file removing error: %v", err)
		} else {
			log.Log.Debugf("File %s removed", session.filePath)
			store.release(session.parentPath, session.expectedFileSize)
		}
	}
}
