package client

import (
	"encoding/json"
	"net"
	"os"

	"file-transfer/internal/dto"
	"file-transfer/internal/log"
)

const (
	listenPort = 4444
	bufferSize = 1024
)

func Start(serverAddr *net.TCPAddr, filePath string) {
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Log.Errorf("File search error: %v", err)
		return
	}
	log.Log.Debugf("Found file %s", filePath)

	// Create control connection
	controlConn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		log.Log.Errorf("Control connection creation error: %v", err)
		return
	}
	log.Log.Debugf("Created control connection to %v", controlConn.RemoteAddr())
	defer func() {
		if err := controlConn.Close(); err != nil {
			log.Log.Errorf("Control connection closing error: %v", err)
			return
		}
		log.Log.Debugf("Closed control connection to %v", controlConn.RemoteAddr())
	}()

	encoder := json.NewEncoder(controlConn)
	decoder := json.NewDecoder(controlConn)

	// Send file info
	fileInfoRequest := dto.FileInfoRequest{
		Type:     "file_info",
		FileName: fileInfo.Name(),
		FileSize: fileInfo.Size(),
	}
	if !sendRequest(fileInfoRequest, encoder) {
		return
	}
	if !receiveResponse(decoder) {
		return
	}

	// Create transfer connection
	transferConnListener, done := createTransferConnectionListener(controlConn.LocalAddr().(*net.TCPAddr).IP)
	if !done {
		return
	}

	startTransferRequest := dto.StartTransferRequest{
		Type: "start_transfer",
		Port: listenPort,
	}
	if !sendRequest(startTransferRequest, encoder) {
		return
	}

	transferConn, done := acceptTransferConnection(serverAddr.IP, transferConnListener)
	if !done {
		return
	}
	if err := transferConnListener.Close(); err != nil {
		log.Log.Errorf("Transfer connection listener closing error: %v", err)
	} else {
		log.Log.Debugf("Closed transfer connection listener to %v", transferConnListener.Addr())
	}

	// Send file data
	fileHashSum, done := sendFileData(transferConn, filePath)
	if !done {
		return
	}
	if err := transferConn.Close(); err != nil {
		log.Log.Errorf("Transfer connection closing error: %v", err)
	} else {
		log.Log.Debugf("Closed transfer connection to %v", transferConn.RemoteAddr())
	}

	// Send end_transfer
	endTransferRequest := dto.EndTransferRequest{
		Type:        "end_transfer",
		FileHashSum: fileHashSum,
	}
	if !sendRequest(endTransferRequest, encoder) {
		return
	}
	if !receiveResponse(decoder) {
		return
	}
	log.Log.Infof("File %s transferred successfully", fileInfo.Name())
}

func sendRequest(request any, encoder *json.Encoder) bool {
	if err := encoder.Encode(request); err != nil {
		log.Log.Errorf("Request sending error: %v", err)
		return false
	}
	return true
}

func receiveResponse(decoder *json.Decoder) bool {
	var response dto.Response
	if err := decoder.Decode(&response); err != nil {
		log.Log.Errorf("Response receiving error: %v", err)
		return false
	}
	if response.Type == "error" {
		log.Log.Errorf("Received error response: %s", response.Message)
		return false
	}
	if response.Type != "success" {
		log.Log.Errorf("Response type is not valid")
		return false
	}
	return true
}
