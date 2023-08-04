package app

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"dario.cat/mergo"
	"github.com/pwnedgod/go-selectivetesting"
	"github.com/pwnedgod/go-selectivetesting/internal/util"
	"golang.org/x/mod/modfile"
)

type config struct {
	RelativePath    string          `json:"relativePath"`
	PrettyOutput    bool            `json:"prettyOutput"`
	Patterns        commaSepStrings `json:"patterns"`
	ModuleDir       string          `json:"moduleDir"`
	BasePkg         string          `json:"basePkg"`
	Depth           int             `json:"depth"`
	BuildFlags      commaSepStrings `json:"buildFlags"`
	AnalyzerOutPath string          `json:"analyzerOutPath"`
	MiscUsages      []struct {
		Regexp string `json:"regexp"`
		UsedBy []struct {
			PkgPath   string           `json:"pkgPath"`
			TestNames util.Set[string] `json:"testNames"`
		} `json:"usedBy"`
	} `json:"miscUsages"`
}

func (cfg config) getBasePkg() (string, error) {
	basePkg := cfg.BasePkg
	if basePkg == "" {
		goModData, err := os.ReadFile(filepath.Join(cfg.ModuleDir, "go.mod"))
		if err != nil {
			return "", err
		}
		basePkg = modfile.ModulePath(goModData)
	}
	return basePkg, nil
}

func (cfg config) getInputBasePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, cfg.RelativePath), nil
}

func (cfg config) asOptions(pathReplacements map[string]string) ([]selectivetesting.Option, error) {
	options := make([]selectivetesting.Option, 0)

	if cfg.ModuleDir != "" {
		options = append(options, selectivetesting.WithModuleDir(cfg.ModuleDir))
	}

	if len(cfg.Patterns) > 0 {
		options = append(options, selectivetesting.WithPatterns(cfg.Patterns...))
	}

	if cfg.Depth > 0 {
		options = append(options, selectivetesting.WithDepth(cfg.Depth))
	}

	if len(cfg.BuildFlags) > 0 {
		options = append(options, selectivetesting.WithBuildFlags(cfg.BuildFlags...))
	}

	if len(cfg.MiscUsages) > 0 {
		miscUsages := make([]selectivetesting.MiscUsage, 0, len(cfg.MiscUsages))
		for _, miscUsage := range cfg.MiscUsages {
			usedBy := make([]selectivetesting.MiscUser, 0, len(miscUsage.UsedBy))
			for _, miscUser := range miscUsage.UsedBy {
				if strings.HasSuffix(miscUser.PkgPath, "/...") && (miscUser.TestNames.Len() != 1 || !miscUser.TestNames.Has("*")) {
					return nil, errors.New("recursive path can only accept a single wildcard")
				}
				usedBy = append(usedBy, selectivetesting.MiscUser{
					PkgPath:   miscUser.PkgPath,
					TestNames: miscUser.TestNames,
				})
			}

			regexStr := miscUsage.Regexp
			for old, new := range pathReplacements {
				regexStr = strings.ReplaceAll(regexStr, old, new)
			}

			regex, err := regexp.Compile(regexStr)
			if err != nil {
				return nil, err
			}

			miscUsages = append(miscUsages, selectivetesting.MiscUsage{
				Regexp: regex,
				UsedBy: usedBy,
			})
		}
		options = append(options, selectivetesting.WithMiscUsages(miscUsages...))
	}

	return options, nil
}

type commaSepStrings []string

func (f commaSepStrings) String() string {
	return strings.Join(f, ",")
}

func (f *commaSepStrings) Set(value string) error {
	*f = strings.Split(value, ",")
	for i, str := range *f {
		(*f)[i] = strings.TrimSpace(str)
	}
	return nil
}

func cfgMerge(cfgs ...config) config {
	if len(cfgs) == 0 {
		return config{}
	}
	cfg := cfgs[0]
	for i := 1; i < len(cfgs); i++ {
		_ = mergo.Merge(&cfg, cfgs[i], mergo.WithOverride)
	}
	return cfg
}
