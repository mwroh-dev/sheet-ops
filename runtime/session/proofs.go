package session

import (
	"errors"
	"path/filepath"
	"strings"
)

func WriteProof(runSession RunSession, name string, value any) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("proof name required")
	}

	cleanName := filepath.Clean(name)
	if filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", errors.New("proof name must stay within proofs dir")
	}

	path := filepath.Join(runSession.ProofsDir, cleanName)
	if err := writeJSON(path, value); err != nil {
		return "", err
	}
	return path, nil
}
