package client

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"os"

	"file-transfer/internal/core"
)

func createTransferConnectionListener(listenIp net.IP) (*net.TCPListener, bool) {
	listenAddr := &net.TCPAddr{
		IP:   listenIp,
		Port: listenPort,
	}
	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		core.Log.Errorf("Transfer connection listener error: %v", err)
		return nil, false
	}
	core.Log.Infof("Listen to %v for transfer connection", listenAddr)
	return listener, true
}

func acceptTransferConnection(serverIp net.IP, listener *net.TCPListener) (net.Conn, bool) {
	var transferConn net.Conn
	var err error
	for {
		transferConn, err = listener.Accept()
		if err != nil {
			core.Log.Errorf("Transfer connection accepting error: %v", err)
			return nil, false
		}

		if transferConn.RemoteAddr().(*net.TCPAddr).IP.String() == serverIp.String() {
			break
		}
	}
	core.Log.Infof("Created transfer connection from %v", transferConn.RemoteAddr())
	return transferConn, true
}

func sendFileData(transferConn net.Conn, filePath string) (string, bool) {
	// Open file
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		core.Log.Errorf("File opening error: %v", err)
		return "", false
	}
	core.Log.Infof("Open file %s", filePath)
	defer func() {
		file.Close()
		core.Log.Infof("Closed file %s", filePath)
	}()

	// Send file data and calculate file hash sum
	buffer := make([]byte, bufferSize)
	fileHashSum := sha256.New()
	multiWriter := io.MultiWriter(transferConn, fileHashSum)

	core.Log.Infof("Start transfer file %s", file.Name())
	for {
		read, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			core.Log.Errorf("File reading error: %v", err)
			return "", false
		}

		_, err = multiWriter.Write(buffer[:read])
		if err != nil {
			core.Log.Errorf("File data sending error: %v", err)
			return "", false
		}
	}

	core.Log.Infof("Finish transfer file %s", file.Name())

	return hex.EncodeToString(fileHashSum.Sum(nil)), true
}
