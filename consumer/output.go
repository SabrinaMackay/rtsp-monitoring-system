package consumer

import (
	"deepalerttest/utils"
	"fmt"
)

func StartResultConsumer(results <-chan utils.CameraHealthResult) {
	go func() {
		for result := range results {
			fmt.Println("message:", result)
		}
	}()
}
