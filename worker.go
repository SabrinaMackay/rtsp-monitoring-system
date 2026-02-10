package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func (r CameraTestResult) String() string {
	if r.Error != nil {
		return fmt.Sprintf("ID: %d, Name: %s, ffmpeg error: %v, error_class: %s",
			r.ID, r.Name, r.Error, r.ErrorClass)
	}
	return fmt.Sprintf("ID: %d, Name: %s, ok", r.ID, r.Name)
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
		if testCtx.Err() == context.DeadlineExceeded {
			result.Error = err
			result.ErrorClass = "timeout"
		} else {
			result.Error = err
			result.ErrorClass = classifyFFmpegError(string(out))
		}
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
