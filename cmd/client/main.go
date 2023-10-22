package main

import (
	"file-transfer/internal/client"
	"file-transfer/internal/consul"
	"fmt"
	"github.com/gogf/gf/v2/util/grand"
	"net"
	"time"

	"file-transfer/internal/client/cli"
	"file-transfer/internal/log"
	"github.com/gogf/gf/v2/os/gfile"
)

const title string = "                                                                             \n" +
	"    ____________    ______   __________  ___    _   _______ ________________ \n" +
	"   / ____/  _/ /   / ____/  /_  __/ __ \\/   |  / | / / ___// ____/ ____/ __ \\\n" +
	"  / /_   / // /   / __/      / / / /_/ / /| | /  |/ /\\__ \\/ /_  / __/ / /_/ /\n" +
	" / __/ _/ // /___/ /___     / / / _, _/ ___ |/ /|  /___/ / __/ / /___/ _, _/ \n" +
	"/_/   /___/_____/_____/    /_/ /_/ |_/_/  |_/_/ |_//____/_/   /_____/_/ |_|  \n" +
	"                                                                             \n" +
	"File Transfer Client: 1.0.0                                                  \n"

func main() {
	fmt.Println(title)

	consulIp, consulPort, path, suffix, parallel, tag := cli.Parse()

	ch := make(chan struct{}, parallel)
	for {
		ch <- struct{}{}
		list, err := gfile.ScanDirFile(path, suffix, false)
		if err != nil {
			log.Log.Errorf("Scanning file error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(list) == 0 {
			log.Log.Infof("No matching files (%s) found in %s", suffix, path)
			time.Sleep(5 * time.Second)
			continue
		}

		fileName := list[0]
		tmpFileName := fmt.Sprintf("%s.%s", fileName, "fmv")

		if err := gfile.Move(fileName, tmpFileName); err != nil {
			log.Log.Errorf("Moving file error: %v", err)
			continue
		}

		go func() {
			service, err := consul.Discovery("file-server", fmt.Sprintf("%s:%d", consulIp, consulPort), tag, parallel)
			if err != nil {
				log.Log.Errorf("Discovery service error: %v", err)
				return
			}
			localPort := grand.N(5000, 10000)

			serverAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%d", service.Service.Address, service.Service.Port))
			if err != nil {
				log.Log.Errorf("Resolving error: %v", serverAddr)
				return
			}

			client.Start(serverAddr, tmpFileName, localPort)
			<-ch
		}()

	}

}
