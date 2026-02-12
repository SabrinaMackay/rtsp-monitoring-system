package producer

import (
	"context"
	"deepalerttest/utils"
	"os/exec"
	"strings"
)

func TestCamera(ctx context.Context, camera utils.Camera) utils.CameraTestResult {
	result := utils.CameraTestResult{
		ID:   camera.ID,
		Name: camera.Name,
	}

	testCtx, cancel := context.WithTimeout(ctx, utils.FfmpegTimeout)
	defer cancel()

	cmd := exec.CommandContext(testCtx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-timeout", utils.RtspTimeout,
		"-i", camera.Rtsp_url,
		"-frames:v", "1",
		"-f", "null", "-",
	)

	out, err := cmd.CombinedOutput()

	if err != nil {
		if testCtx.Err() == context.DeadlineExceeded {
			result.Status = "context_timeout"
		} else {
			result.Status = classifyFFmpegErrorStatus(string(out))
		}
	} else {
		result.Status = "healthy"
	}

	return result
}

func classifyFFmpegErrorStatus(log string) string {
	l := strings.ToLower(log)

	errorPatterns := map[string][]string{
		"unauthorised": {"401", "unauthorised"},
		"offline":      {"404", "port missing", "invalid argument", "connection refused", "no route to host", "timed out", "could not resolve", "could not find codec", "invalid data"},
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
