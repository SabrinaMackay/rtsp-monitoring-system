package main

import "encoding/json"

type Camera struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Rtsp_url string `json:"rtsp_url"`
}

type CameraTestResult struct {
	ID         int
	Name       string
	Error      error
	ErrorClass string
}

func (c *Camera) Bytes() []byte {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return nil
	}
	return jsonData
}
