package util

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	defaultDirOnce sync.Once
	defaultDir     string
)

func GoCacheFolder() string {
	defaultDirOnce.Do(func() {
		defaultDir = os.Getenv("GOCACHE")
		if filepath.IsAbs(defaultDir) || defaultDir == "off" {
			return
		}
		if defaultDir != "" {
			defaultDir = "off"
			return
		}

		// Compute default location.
		dir, err := os.UserCacheDir()
		if err != nil {
			defaultDir = "off"
			return
		}
		defaultDir = filepath.Join(dir, "go-build")
	})

	return defaultDir
}
