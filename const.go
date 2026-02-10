package main

import "time"

const (
	maxWorkers    = 10
	ffmpegTimeout = 15 * time.Second
	rtspTimeout   = "5000000" // 5s in microseconds
)
