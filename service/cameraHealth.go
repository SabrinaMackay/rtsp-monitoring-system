package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"deepalerttest/consumer"
	"deepalerttest/producer"
	"deepalerttest/utils"

	"github.com/golang-queue/queue"
	"github.com/golang-queue/queue/core"
)

func CheckCamerasHealthy(ctx context.Context, cameras []utils.Camera) {
	var wg sync.WaitGroup

	results := make(chan utils.CameraHealthResult)

	consumer.StartResultConsumer(results)

	queueCamera := queue.NewPool(utils.MaxWorkers, queue.WithFn(func(ctx context.Context, m core.TaskMessage) error {
		defer wg.Done()
		var camera utils.Camera
		if err := json.Unmarshal(m.Payload(), &camera); err != nil {
			return fmt.Errorf("failed to unmarshal camera: %w", err)
		}

		result := producer.CheckCamera(ctx, camera)
		results <- result
		return nil
	}))
	defer queueCamera.Release()

	for _, camera := range cameras {
		wg.Add(1)
		cam := camera
		if err := queueCamera.Queue(&cam); err != nil {
			wg.Done()
			results <- utils.CameraHealthResult{
				ID:     cam.ID,
				Name:   cam.Name,
				Status: err.Error(),
			}
		}
	}
	wg.Wait()
	close(results)
}
