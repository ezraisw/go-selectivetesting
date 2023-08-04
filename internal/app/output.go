package app

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/pwnedgod/go-selectivetesting/internal/util"
)

type testedPkg struct {
	PkgPath         string   `json:"pkgPath"`
	RelativePkgPath string   `json:"relativePkgPath"`
	TestNames       []string `json:"testNames"`
	RunRegex        string   `json:"runRegex"`
}

func cleanTestedPkgs(basePkg string, crudeTestedPkgs map[string]util.Set[string]) []testedPkg {
	testedPkgs := make([]testedPkg, 0, len(crudeTestedPkgs))
	for pkgPath, testNameSet := range crudeTestedPkgs {
		testNames := testNameSet.ToSlice()
		sort.Strings(testNames)

		runRegex := ".*"
		if !testNameSet.Has("*") {
			sanitizedTestNames := make([]string, 0, len(testNames))
			for _, testName := range testNames {
				sanitizedTestNames = append(sanitizedTestNames, regexp.QuoteMeta(testName))
			}

			regexPiece := strings.Join(sanitizedTestNames, "|")
			if len(sanitizedTestNames) == 1 {
				runRegex = "^" + regexPiece + "$"
			} else {
				runRegex = "^(" + regexPiece + ")$"
			}
		}

		testedPkgs = append(testedPkgs, testedPkg{
			PkgPath:         pkgPath,
			RelativePkgPath: util.RelatifyPath(basePkg, pkgPath),
			TestNames:       testNames,
			RunRegex:        runRegex,
		})
	}

	sort.Slice(testedPkgs, func(i, j int) bool {
		return testedPkgs[i].PkgPath < testedPkgs[j].PkgPath
	})

	return testedPkgs
}

func jsonTo(out io.Writer, prettyOutput bool, content any) error {
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	if prettyOutput {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(content); err != nil {
		return fmt.Errorf("could not encode json: %w", err)
	}
	return nil
}
