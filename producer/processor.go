package producer

import (
	"context"
	"deepalerttest/utils"
	"os/exec"
	"strings"
)

func CheckCamera(ctx context.Context, camera utils.Camera) utils.CameraHealthResult {
	result := utils.CameraHealthResult{
		ID:   camera.ID,
		Name: camera.Name,
	}

	healthCtx, cancel := context.WithTimeout(ctx, utils.FfmpegTimeout)
	defer cancel()

	cmd := exec.CommandContext(healthCtx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-timeout", utils.RtspTimeout,
		"-i", camera.Rtsp_url,
		"-frames:v", "1",
		"-f", "null", "-",
	)

	out, err := cmd.CombinedOutput()

	if err != nil {
		if healthCtx.Err() == context.DeadlineExceeded {
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
	}

	for errorType, patterns := range errorPatterns {
		for _, pattern := range patterns {
			if strings.Contains(l, pattern) {
				return errorType
			}
		}
	}

	return "offline"
}
