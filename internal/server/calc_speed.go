package server

import (
	"time"

	"file-transfer/internal/core"
)

type speedComputer struct {
	startTime              time.Time
	previousTime           time.Time
	previousActualFileSize int64
	sessionInfo            *session
}

func newSpeedCounter() *speedComputer {
	return &speedComputer{
		startTime:              time.Now(),
		previousTime:           time.Now(),
		previousActualFileSize: 0,
	}
}

func (speedCounter *speedComputer) calcSpeed(sessionInfo *session) {
	instantSpeed := float64(sessionInfo.actualFileSize-speedCounter.previousActualFileSize) * 1e9 /
		float64(time.Now().UnixNano()-speedCounter.previousTime.UnixNano())
	averageSpeed := float64(sessionInfo.actualFileSize) * 1e9 /
		float64(time.Now().UnixNano()-speedCounter.startTime.UnixNano())

	speedCounter.previousTime = time.Now()
	speedCounter.previousActualFileSize = sessionInfo.actualFileSize

	core.Log.Infof("Transfer %s: %s, %s", sessionInfo.filePath, speedToString(instantSpeed),
		speedToString(averageSpeed))
}
