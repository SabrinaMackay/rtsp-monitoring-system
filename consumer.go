package main

import (
	"fmt"
	"sync"
)

func startResultConsumer(results <-chan CameraTestResult, expected int, wg *sync.WaitGroup) {
	go func() {

		for i := 0; i < expected; i++ {
			result := <-results
			fmt.Println("message:", result)
			wg.Done()
		}
	}()
}
