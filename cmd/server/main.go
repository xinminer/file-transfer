package main

import (
	"fmt"
	"net"

	"file-transfer/internal/core"
	"file-transfer/internal/server"
	cliparser "file-transfer/internal/server/cli-parser"
)

const title string = "                                                                             \n" +
	"    ____________    ______   __________  ___    _   _______ ________________ \n" +
	"   / ____/  _/ /   / ____/  /_  __/ __ \\/   |  / | / / ___// ____/ ____/ __ \\\n" +
	"  / /_   / // /   / __/      / / / /_/ / /| | /  |/ /\\__ \\/ /_  / __/ / /_/ /\n" +
	" / __/ _/ // /___/ /___     / / / _, _/ ___ |/ /|  /___/ / __/ / /___/ _, _/ \n" +
	"/_/   /___/_____/_____/    /_/ /_/ |_/_/  |_/_/ |_//____/_/   /_____/_/ |_|  \n" +
	"                                                                             \n" +
	"File Transfer Server: 1.0.0                                                  \n"

func main() {
	fmt.Println(title)

	port := cliparser.Parse()

	serverAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		core.Log.Errorf("Resolving error: %v", serverAddr)
		return
	}

	server.Start(serverAddr)
}
