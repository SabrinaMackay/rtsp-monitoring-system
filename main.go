package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-queue/queue"
	"github.com/golang-queue/queue/core"
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

func testCameras(ctx context.Context, cameras []Camera) {
	numCameras := len(cameras)
	results := make(chan CameraTestResult, numCameras)

	fmt.Println("Number of cameras:", numCameras)
	var wg sync.WaitGroup
	wg.Add(numCameras)

	startResultConsumer(results, numCameras, &wg)

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

	for _, camera := range cameras {
		cam := camera
		if err := queueCamera.Queue(&cam); err != nil {
			log.Printf("Failed to queue camera %d: %v", cam.ID, err)
			results <- CameraTestResult{
				ID:         cam.ID,
				Name:       cam.Name,
				Error:      err,
				ErrorClass: "queue_error",
			}
		}
	}
	wg.Wait()

}
