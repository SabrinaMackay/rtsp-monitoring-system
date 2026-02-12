package utils

import (
	"encoding/json"
	"fmt"
)

type Camera struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Rtsp_url string `json:"rtsp_url"`
}

func (c *Camera) Bytes() []byte {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return nil
	}
	return jsonData
}

type CameraTestResult struct {
	ID     int
	Name   string
	Status string
}

func (r CameraTestResult) String() string {
	return fmt.Sprintf("ID: %d, Name: %s, Status: %s", r.ID, r.Name, r.Status)
}
