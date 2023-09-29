package server

import (
	"encoding/json"
	"net"
)

type session struct {
	encoder          *json.Encoder
	decoder          *json.Decoder
	controlAddr      *net.TCPAddr
	fileIsCreated    bool
	filePath         string
	expectedFileSize int64
	actualFileSize   int64
	fileHashSum      string
	stage            string
}

func newSession(conn net.Conn) *session {
	return &session{
		encoder:          json.NewEncoder(conn),
		decoder:          json.NewDecoder(conn),
		controlAddr:      conn.RemoteAddr().(*net.TCPAddr),
		fileIsCreated:    false,
		filePath:         "",
		expectedFileSize: 0,
		actualFileSize:   0,
		fileHashSum:      "",
		stage:            "file_info",
	}
}
