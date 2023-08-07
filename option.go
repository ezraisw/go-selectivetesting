package selectivetesting

type Option func(*FileAnalyzer)

func WithModuleDir(moduleDir string) Option {
	return func(fa *FileAnalyzer) {
		fa.moduleDir = moduleDir
	}
}

func WithPatterns(patterns ...string) Option {
	return func(fa *FileAnalyzer) {
		fa.patterns = patterns
	}
}

func WithDepth(depth int) Option {
	return func(fa *FileAnalyzer) {
		fa.depth = depth
	}
}

func WithBuildFlags(buildFlags ...string) Option {
	return func(fa *FileAnalyzer) {
		fa.buildFlags = buildFlags
	}
}

func WithMiscUsages(miscUsages ...MiscUsage) Option {
	return func(fa *FileAnalyzer) {
		fa.miscUsages = miscUsages
	}
}

func WithTestAll(testAll bool) Option {
	return func(fa *FileAnalyzer) {
		fa.testAll = testAll
	}
}
