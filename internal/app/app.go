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
		if err := writeFileAnalyzerTo(cfg.AnalyzerOutPath, fa); err != nil {
			return err
		}
	}
	if !cfg.GoTest.Run {
		testedPkgGroups := groupBy(testedPkgs, cfg.Groups)
		return jsonTo(os.Stdout, cfg.PrettyOutput, testedPkgGroups)
	}
	return runTests(cfg.ModuleDir, cfg.GoTest.Args, cfg.GoTest.Parallel, testedPkgs)
}
