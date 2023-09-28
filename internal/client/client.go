package client

import (
	"encoding/json"
	"net"
	"os"

	"file-transfer/internal/core"
)

const (
	listenerPort = 4444
	bufferSize   = 1024
)

func Start(serverAddr *net.TCPAddr, filePath string) {
	// Get file info
	fileInfo, done := getFileInfo(filePath)
	if !done {
		return
	}

	// Create control connection
	controlConn, done := createControlConnection(serverAddr)
	if !done {
		return
	}
	defer controlConn.Close()
	encoder := json.NewEncoder(controlConn)
	decoder := json.NewDecoder(controlConn)

	// Send file info
	fileInfoRequest := core.FileInfoRequest{
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

	// Create transfer connection listener
	listenAddr := net.TCPAddr{
		IP:   controlConn.LocalAddr().(*net.TCPAddr).IP,
		Port: listenerPort,
	}
	listener, done := createTransferConnListener(&listenAddr)
	if !done {
		return
	}

	// Send start_transfer
	startTransferRequest := core.StartTransferRequest{
		Type: "start_transfer",
		Port: listenerPort,
	}
	if !sendRequest(startTransferRequest, encoder) {
		return
	}

	// Create transfer connection
	transferConn, done := acceptTransferConn(listener, serverAddr.IP)
	listener.Close()
	if !done {
		return
	}

	// Send file data
	fileHashSum, done := sendFileData(transferConn, filePath)
	if !done {
		return
	}
	transferConn.Close()

	// Send end_transfer
	endTransferRequest := core.EndTransferRequest{
		Type:        "end_transfer",
		FileHashSum: fileHashSum,
	}
	if !sendRequest(endTransferRequest, encoder) {
		return
	}
	if !receiveResponse(decoder) {
		return
	}
	core.Log.Infof("File %s transferred successfully", fileInfo.Name())
}

func getFileInfo(filePath string) (os.FileInfo, bool) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		core.Log.Errorf("File search error: %v", err)
		return nil, false
	}
	core.Log.Infof("Found file %s", filePath)
	return fileInfo, true
}

func createControlConnection(serverAddr *net.TCPAddr) (*net.TCPConn, bool) {
	controlConn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		core.Log.Errorf("Control connection error: %v", err)
		return nil, false
	}
	core.Log.Infof("Created control connection to %v", controlConn.RemoteAddr())
	return controlConn, true
}

func sendRequest(request any, encoder *json.Encoder) bool {
	if err := encoder.Encode(request); err != nil {
		core.Log.Errorf("Request sending error: %v", err)
		return false
	}
	return true
}

func receiveResponse(decoder *json.Decoder) bool {
	var response core.Response
	if err := decoder.Decode(&response); err != nil {
		core.Log.Errorf("Response receiving error: %v", err)
		return false
	}
	if response.Type == "error" {
		core.Log.Errorf("Received error response: %s", response.Message)
		return false
	}
	if response.Type != "success" {
		core.Log.Errorf("Response type is not valid")
		return false
	}
	return true
}
