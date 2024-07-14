package app

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"dario.cat/mergo"
	"github.com/ezraisw/go-selectivetesting"
	"golang.org/x/mod/modfile"
)

type goTest struct {
	Run      bool   `json:"run"`
	Args     string `json:"args"`
	Parallel int    `json:"parallel"`
}

type group struct {
	Name     string   `json:"name"`
	Patterns []string `json:"patterns"`
}

type miscUsage struct {
	Regexp string `json:"regexp"`
	UsedBy []struct {
		PkgPath   string   `json:"pkgPath"`
		All       bool     `json:"all"`
		FileNames []string `json:"fileNames"`
		ObjNames  []string `json:"objNames"`
	} `json:"usedBy"`
}

type config struct {
	RelativePath      string          `json:"relativePath"`
	PrettyOutput      bool            `json:"prettyOutput"`
	Patterns          commaSepStrings `json:"patterns"`
	ModuleDir         string          `json:"moduleDir"`
	BasePkg           string          `json:"basePkg"`
	Depth             int             `json:"depth"`
	BuildFlags        commaSepStrings `json:"buildFlags"`
	TestAll           bool            `json:"testAll"`
	AnalyzerOutPath   string          `json:"analyzerOutPath"`
	GoTest            goTest          `json:"goTest"`
	Groups            []group         `json:"groups"`
	OutputEmptyGroups bool            `json:"outputEmptyGroups"`
	MiscUsages        []miscUsage     `json:"miscUsages"`
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

	if cfg.TestAll {
		options = append(options, selectivetesting.WithTestAll(cfg.TestAll))
	}

	if len(cfg.MiscUsages) > 0 {
		miscUsages := make([]selectivetesting.MiscUsage, 0, len(cfg.MiscUsages))
		for _, miscUsage := range cfg.MiscUsages {
			usedBy := make([]selectivetesting.MiscUser, 0, len(miscUsage.UsedBy))
			for _, miscUser := range miscUsage.UsedBy {
				var (
					recursive bool
					fileNames []string
					objNames  []string
				)
				if strings.HasSuffix(miscUser.PkgPath, "/...") {
					recursive = true
				} else {
					fileNames = miscUser.FileNames
					objNames = miscUser.ObjNames
				}
				usedBy = append(usedBy, selectivetesting.MiscUser{
					PkgPath: miscUser.PkgPath,
					All:     recursive || miscUser.All,

					// Should only be filled when All is false.
					FileNames: fileNames,
					ObjNames:  objNames,
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
