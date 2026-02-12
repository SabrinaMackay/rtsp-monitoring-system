package main

import (
	"context"
	"deepalerttest/service"
	"deepalerttest/utils"
	"fmt"
	"log"
	"time"
)

func main() {

	ticker := time.NewTicker(1 * time.Minute) // Run every minute
	defer ticker.Stop()

	for {
		cameraData, err := utils.GetCameraData()
		if err != nil {
			log.Printf("failed to get camera data: %v", err)
		} else {
			fmt.Println("Starting camera health check run")
			service.CheckCamerasHealthy(context.Background(), cameraData)
		}
		<-ticker.C
	}
}
