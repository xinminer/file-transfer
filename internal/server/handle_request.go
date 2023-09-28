package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

	"file-transfer/internal/core"
)

const (
	maxFileSize       = 1 << 40 // 1 TB
	maxFileNameLength = 4096
	bufferSize        = 1024
)

func handleRequests(controlConn net.Conn) {
	defer func() {
		if err := controlConn.Close(); err != nil {
			core.Log.Errorf("Control connection closing error: %v", err)
			return
		}
		core.Log.Infof("Closed control connection from %v", controlConn.RemoteAddr())
	}()

	// Create session
	session := newSession(controlConn)

	var data map[string]interface{}
	for {
		// Receive JSON message
		if err := session.decoder.Decode(&data); err != nil {
			core.Log.Errorf("Request receiving error: Expected message with JSON format: %v", err)
			sendResponse("error", "Expected message with JSON format", session.encoder)
			return
		}

		// Check type field
		if _, ok := data["type"]; !ok {
			core.Log.Errorf("Request conversion error: Field \"type\" is missing")
			sendResponse("error", "Field \"type\" is missing", session.encoder)
			return
		}

		// Handle data
		switch data["type"].(string) {
		case "file_info":
			if !handleFileInfoRequest(data, session) {
				return
			}
		case "start_transfer":
			if !handleStartTransferRequest(data, session) {
				return
			}
		case "end_transfer":
			handleEndTransferRequest(data, session)
			return
		default:
			handleUnsupportedRequest(data, session.encoder)
			return
		}
	}
}

func handleFileInfoRequest(data map[string]interface{}, session *session) bool {
	// Convert data to file_info request
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		core.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"file_info\" is not valid", session.encoder)
		return false
	}

	var fileInfoRequest core.FileInfoRequest
	if err := json.Unmarshal(jsonBytes, &fileInfoRequest); err != nil {
		core.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"file_info\" is not valid", session.encoder)
		return false
	}

	// Check stage
	if session.stage != "file_info" {
		core.Log.Errorf("Stage error: Expected message \"file_info\"")
		sendResponse("error", "Expected message \"file_info\"", session.encoder)
		return false
	}

	// Validate file info
	if err := validateFileInfo(fileInfoRequest); err != nil {
		core.Log.Errorf("Message \"file_info\" validation error: %v", err)
		sendResponse("error", err.Error(), session.encoder)
		return false
	}

	// Create file
	if _, err := os.Create(uploadDir + "/" + fileInfoRequest.FileName); err != nil {
		core.Log.Errorf("File creation error: %v", err)
		sendResponse("error",
			fmt.Sprintf("Failed to create file %s", fileInfoRequest.FileName), session.encoder)
		return false
	}
	core.Log.Infof("Created file %s", fileInfoRequest.FileName)

	// Update session info
	session.fileName = fileInfoRequest.FileName
	session.expectedFileSize = fileInfoRequest.FileSize
	session.stage = "start_transfer"

	// Send success response
	sendResponse("success", "", session.encoder)

	return true
}

func handleStartTransferRequest(data map[string]interface{}, session *session) bool {
	// Convert data to start_transfer request
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		core.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"start_transfer\" is not valid", session.encoder)
		return false
	}

	var startTransferRequest core.StartTransferRequest
	if err := json.Unmarshal(jsonBytes, &startTransferRequest); err != nil {
		core.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"start_transfer\" is not valid", session.encoder)
		return false
	}

	// Check stage
	if session.stage != "start_transfer" {
		core.Log.Errorf("Stage error: Expected message \"start_transfer\"")
		sendResponse("error", "Expected message \"start_transfer\"", session.encoder)
		return false
	}

	// Transfer file
	transferAddr, err := net.ResolveTCPAddr("tcp",
		fmt.Sprintf("%s:%d", session.controlAddr.IP, startTransferRequest.Port))
	if err != nil {
		core.Log.Errorf("Transfer address resolving error: %v", err)
		sendResponse("error", "Internal server error", session.encoder)
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
		core.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"end_transfer\" is not valid", session.encoder)
		return false
	}

	var endTransferRequest core.EndTransferRequest
	if err := json.Unmarshal(jsonBytes, &endTransferRequest); err != nil {
		core.Log.Errorf("Request conversion error: %v", err)
		sendResponse("error", "Message \"end_transfer\" is not valid", session.encoder)
		return false
	}

	// Check stage
	if session.stage != "end_transfer" {
		core.Log.Errorf("Stage error: Expected message \"end_transfer\"", session.encoder)
		sendResponse("error", "Expected message \"end_transfer\"", session.encoder)
		return false
	}

	// Check transferred file
	if session.expectedFileSize != session.actualFileSize {
		core.Log.Errorf("Actual and expected file sizes is are different")
		sendResponse("error", "Actual and expected file sizes is are different", session.encoder)
		return false
	}
	if session.fileHashSum != endTransferRequest.FileHashSum {
		core.Log.Errorf("Actual and expected file hash sums is are different")
		sendResponse("error", "Actual and expected file hash sums is are different", session.encoder)
		return false
	}

	// Send success response
	sendResponse("success", "", session.encoder)
	core.Log.Infof("File %s transferred successfully", session.fileName)

	// Update session
	session.stage = ""

	return true
}

func handleUnsupportedRequest(data map[string]interface{}, encoder *json.Encoder) {
	core.Log.Errorf("Request receiving error: Message type \"%s\" is unsupported",
		data["type"].(string))
	sendResponse("error",
		fmt.Sprintf("Message type \"%s\" is unsupported", data["type"].(string)),
		encoder)
}

func sendResponse(status string, message string, encoder *json.Encoder) {
	if err := encoder.Encode(core.Response{Type: status, Message: message}); err != nil {
		core.Log.Errorf("Response sending error: %v", err)
		return
	}
}

func validateFileInfo(fileInfo core.FileInfoRequest) error {
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
	if _, err := os.Stat(uploadDir + "/" + fileInfo.FileName); !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("File %s already exists", fileInfo.FileName))
	}

	return nil
}
