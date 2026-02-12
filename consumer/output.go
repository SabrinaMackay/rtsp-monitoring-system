package consumer

import (
	"deepalerttest/utils"
	"fmt"
)

func StartResultConsumer(results <-chan utils.CameraTestResult) {
	go func() {
		for result := range results {
			fmt.Println("message:", result)
		}
	}()
}
