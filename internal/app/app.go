package app

import (
	"fmt"
	"os"

	"github.com/ezraisw/go-selectivetesting"
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
	crudeTestedPkgs, uniqueTestCount := fa.DetermineTests()
	testedPkgs := cleanTestedPkgs(basePkg, crudeTestedPkgs)
	if cfg.AnalyzerOutPath != "" {
		if err := writeFileAnalyzerTo(cfg.AnalyzerOutPath, fa); err != nil {
			return err
		}
	}
	if !cfg.GoTest.Run {
		testedPkgGroups := groupBy(testedPkgs, cfg.Groups, cfg.OutputEmptyGroups)
		return jsonTo(os.Stdout, cfg.PrettyOutput, testing{
			UniqueTestCount: uniqueTestCount,
			Groups:          testedPkgGroups,
		})
	}
	return runTests(cfg.ModuleDir, cfg.GoTest.Args, cfg.GoTest.Parallel, testedPkgs)
}
