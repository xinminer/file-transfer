package server

import (
	"net"
)

type session struct {
	controlConn      *net.TCPConn
	controlAddr      *net.TCPAddr
	fileIsCreated    bool
	filePath         string
	expectedFileSize int64
	actualFileSize   int64
	fileHashSum      string
	stage            string
}

func newSession(controlConn *net.TCPConn) *session {
	return &session{
		controlConn:      controlConn,
		controlAddr:      controlConn.RemoteAddr().(*net.TCPAddr),
		fileIsCreated:    false,
		filePath:         "",
		expectedFileSize: 0,
		actualFileSize:   0,
		fileHashSum:      "",
		stage:            "file_info",
	}
}
