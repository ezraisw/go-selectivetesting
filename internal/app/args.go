package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pwnedgod/go-selectivetesting"
)

func parseArgs() (cfg config, notablePaths []string, err error) {
	var (
		cfgPath     string
		cfgFromFlag config
	)

	flag.StringVar(&cfgPath, "cfgpath", "", "config file to use for command configuration")

	flag.StringVar(&cfgFromFlag.RelativePath, "relativepath", "", "relative path from current working directory for input files")
	flag.BoolVar(&cfgFromFlag.PrettyOutput, "prettyoutput", false, "whether to output indented json")
	flag.Var(&cfgFromFlag.Patterns, "patterns", "patterns to use for package search")
	flag.StringVar(&cfgFromFlag.ModuleDir, "moduledir", "", "path to the directory of the module")
	flag.StringVar(&cfgFromFlag.BasePkg, "basepkg", "", "base package path/module name, will be used instead of <modulepath>/go.mod")
	flag.IntVar(&cfgFromFlag.Depth, "depth", 0, "depth of the test search from input files")
	flag.Var(&cfgFromFlag.BuildFlags, "buildflags", "build flags to use")
	flag.StringVar(&cfgFromFlag.AnalyzerOutPath, "analyzeroutpath", "", "path to output debug information for analyzer")

	flag.Parse()

	var cfgFromFile config
	if cfgPath != "" {
		cfgFile, err := os.Open(cfgPath)
		if err != nil {
			return config{}, nil, err
		}
		if err := json.NewDecoder(cfgFile).Decode(&cfgFromFile); err != nil {
			return config{}, nil, err
		}
	}

	cfg = cfgMerge(cfgFromFile, cfgFromFlag)

	notablePaths = make([]string, 0)
	for i := 0; i < flag.NArg(); i++ {
		notablePaths = append(notablePaths, flag.Arg(i))
	}

	return cfg, notablePaths, nil
}

func forAnalyzer(cfg config, inputPaths []string) (string, []string, []selectivetesting.Option, error) {
	basePkg, err := cfg.getBasePkg()
	if err != nil {
		return "", nil, nil, fmt.Errorf("error while getting base package: %w", err)
	}

	inputBasePath, err := cfg.getInputBasePath()
	if err != nil {
		return "", nil, nil, fmt.Errorf("error while getting base path: %w", err)
	}

	absInputPaths := make([]string, len(inputPaths))
	for _, input := range inputPaths {
		absInput := filepath.Join(inputBasePath, input)
		if _, err := os.Stat(absInput); err != nil {
			return "", nil, nil, fmt.Errorf("error checking file: %w", err)
		}
		absInputPaths = append(absInputPaths, absInput)
	}

	pathReplacements := map[string]string{
		"<<basepath>>": inputBasePath,
	}

	options, err := cfg.asOptions(pathReplacements)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error setting options: %w", err)
	}

	return basePkg, absInputPaths, options, nil
}
