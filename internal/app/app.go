package app

import (
	"fmt"
	"os"

	"github.com/pwnedgod/go-selectivetesting"
)

func Run() error {
	cfg, inputPaths, err := parseArgs()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	basePkg, absInputPaths, options, err := forAnalyzer(cfg, inputPaths)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	fa := selectivetesting.NewFileAnalyzer(basePkg, absInputPaths, options...)
	if err := fa.Load(); err != nil {
		return fmt.Errorf("could not load packages: %w", err)
	}
	crudeTestedPkgs := fa.DetermineTests()
	testedPkgs := cleanTestedPkgs(basePkg, crudeTestedPkgs)
	if cfg.AnalyzerOutPath != "" {
		file, err := os.OpenFile(cfg.AnalyzerOutPath, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			return fmt.Errorf("could create output for analyzer debug: %w", err)
		}
		if err := jsonTo(file, true, fa); err != nil {
			return err
		}
	}
	return jsonTo(os.Stdout, cfg.PrettyOutput, testedPkgs)
}
