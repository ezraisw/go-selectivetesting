package util

import (
	"path/filepath"
	"strings"
)

func IsWithinPath(basePath, targetPath string) bool {
	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(relPath, "../")
}

func IsSubPackage(basePkg, targetPkg string) bool {
	return IsWithinPath("/"+basePkg, "/"+targetPkg)
}

func RelatifyPath(basePath, targetPath string) string {
	relPath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return targetPath
	}
	if strings.HasPrefix(relPath, "../") {
		return targetPath
	}
	return "./" + relPath
}
