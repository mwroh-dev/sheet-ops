package runtimeconfig

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	RetentionModeEnv = "SHEET_OPS_RETENTION_MODE"
	RenderModeEnv    = "SHEET_OPS_RENDER_MODE"
)

type RetentionMode string

const (
	RetentionModeFull     RetentionMode = "full"
	RetentionModeRedacted RetentionMode = "redacted"
)

type RenderMode string

const (
	RenderModeNever     RenderMode = "never"
	RenderModeOnFailure RenderMode = "on_failure"
	RenderModeAlways    RenderMode = "always"
)

func CurrentRetentionMode() RetentionMode {
	switch strings.TrimSpace(os.Getenv(RetentionModeEnv)) {
	case string(RetentionModeRedacted):
		return RetentionModeRedacted
	default:
		return RetentionModeFull
	}
}

func CurrentRenderMode() RenderMode {
	switch strings.TrimSpace(os.Getenv(RenderModeEnv)) {
	case string(RenderModeNever):
		return RenderModeNever
	case string(RenderModeOnFailure):
		return RenderModeOnFailure
	default:
		return RenderModeAlways
	}
}

func UsesRedactedArtifacts() bool {
	return CurrentRetentionMode() == RetentionModeRedacted
}

func UseBasename(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	return filepath.Base(trimmed)
}

func RedactedString(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "[redacted]"
}
