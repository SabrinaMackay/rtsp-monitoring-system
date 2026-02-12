package utils

import "time"

const (
	MaxWorkers    = 10
	FfmpegTimeout = 10 * time.Second
	RtspTimeout   = "5000000" // 5s in microseconds
)
