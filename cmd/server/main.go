package main

import (
	"fmt"
	"net"

	"file-transfer/internal/log"
	"file-transfer/internal/server"
	"file-transfer/internal/server/cli"
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

	port := cli.Parse()

	serverAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Log.Errorf("Resolving error: %v", serverAddr)
		return
	}

	server.Start(serverAddr)
}
