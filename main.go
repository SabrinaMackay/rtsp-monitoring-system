package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/golang-queue/queue"
	"github.com/golang-queue/queue/core"
)

const (
	maxWorkers    = 10
	ffmpegTimeout = 15 * time.Second
	rtspTimeout   = "5000000" // 5s in microseconds
)

func main() {
	databaseConnection, err := NewDatabaseConnection()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	fmt.Println("Database connection established")
	defer databaseConnection.Close()

	ticker := time.NewTicker(1 * time.Minute) // Run every minute
	defer ticker.Stop()

	for {
		cameraData, err := getCameraData()
		if err != nil {
			log.Printf("failed to get camera data: %v", err)
		} else {
			fmt.Println("Starting camera health check run")
			testCameras(context.Background(), cameraData)
		}

		<-ticker.C
	}
}

func (r CameraTestResult) String() string {
	if r.Error != nil {
		return fmt.Sprintf("ID: %d, Name: %s, ffmpeg error: %v, error_class: %s",
			r.ID, r.Name, r.Error, r.ErrorClass)
	}
	return fmt.Sprintf("ID: %d, Name: %s, ok", r.ID, r.Name)
}

func testCameras(ctx context.Context, cameras []Camera) {
	numCameras := len(cameras)
	results := make(chan CameraTestResult, numCameras)

	queueCamera := queue.NewPool(maxWorkers, queue.WithFn(func(ctx context.Context, m core.TaskMessage) error {
		var camera Camera
		if err := json.Unmarshal(m.Payload(), &camera); err != nil {
			return fmt.Errorf("failed to unmarshal camera: %w", err)
		}

		result := testCamera(ctx, camera)
		results <- result
		return nil
	}))
	defer queueCamera.Release()

	// Queue all camera tests
	for _, camera := range cameras {
		cam := camera
		go func() {
			if err := queueCamera.Queue(&cam); err != nil {
				log.Printf("Failed to queue camera %d: %v", cam.ID, err)
				results <- CameraTestResult{
					ID:         cam.ID,
					Name:       cam.Name,
					Error:      err,
					ErrorClass: "queue_error",
				}
			}
		}()
	}

	// Display results
	for i := 0; i < numCameras; i++ {
		result := <-results
		fmt.Println("message:", result)
		time.Sleep(50 * time.Millisecond)
	}
}

func testCamera(ctx context.Context, camera Camera) CameraTestResult {
	result := CameraTestResult{
		ID:   camera.ID,
		Name: camera.Name,
	}

	testCtx, cancel := context.WithTimeout(ctx, ffmpegTimeout)
	defer cancel()

	cmd := exec.CommandContext(testCtx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-timeout", rtspTimeout,
		"-i", camera.Rtsp_url,
		"-frames:v", "1",
		"-f", "null", "-",
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		result.Error = err
		result.ErrorClass = classifyFFmpegError(string(out))
	}

	return result
}

func classifyFFmpegError(log string) string {
	l := strings.ToLower(log)

	errorPatterns := map[string][]string{
		"unauthorised": {"401"},
		"not_found":    {"404"},
		"invalid_url":  {"port missing in uri", "invalid argument"},
		"unreachable":  {"connection refused", "no route to host", "timed out", "could not resolve"},
		"stream_error": {"could not find codec", "invalid data"},
	}

	for errorType, patterns := range errorPatterns {
		for _, pattern := range patterns {
			if strings.Contains(l, pattern) {
				return errorType
			}
		}
	}

	return "unknown_error"
}
