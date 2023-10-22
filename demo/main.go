package main

import (
	"file-transfer/internal/log"
	"github.com/shirou/gopsutil/v3/disk"
)

func main() {
	stat, err := disk.Usage("/mnt/data1")
	if err != nil {
		log.Log.Infof("stat error: %v", err)
	}
	log.Log.Infof("stat info: %v", stat)
}
