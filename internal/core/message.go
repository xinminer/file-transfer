package core

type FileInfoRequest struct {
	Type     string `json:"type"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
}

type StartTransferRequest struct {
	Type string `json:"type"`
	Port int    `json:"port"`
}

type EndTransferRequest struct {
	Type        string `json:"type"`
	FileHashSum string `json:"fileHashSum"`
}

type Response struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
