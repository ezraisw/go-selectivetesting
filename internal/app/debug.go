package app

import (
	"fmt"
	"os"

	"github.com/ezraisw/go-selectivetesting"
)

func writeFileAnalyzerTo(path string, fa *selectivetesting.FileAnalyzer) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("could create output for analyzer debug: %w", err)
	}
	return jsonTo(file, true, fa)
}
