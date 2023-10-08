package server

import "fmt"

func speedToString(speed float64) string {
	if speed < 1024 {
		return fmt.Sprintf("%.2f B/s", speed)
	} else if speed < 1024*1024 {
		return fmt.Sprintf("%.2f KB/s", speed/1024)
	} else if speed < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB/s", speed/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB/s", speed/(1024*1024*1024))
	}
}
