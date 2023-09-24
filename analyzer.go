package selectivetesting

import (
	"container/heap"
	"encoding/json"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pwnedgod/go-selectivetesting/internal/util"
	"golang.org/x/tools/go/packages"
)

type definition struct {
	obj            types.Object
	fileName       string
	node           ast.Node
	usedByObjNames util.Set[string]
	usingObjNames  util.Set[string]
}

type MiscUser struct {
	PkgPath   string
	All       bool
	FileNames []string
	ObjNames  []string
}

type MiscUsage struct {
	Regexp *regexp.Regexp
	UsedBy []MiscUser
}

type TestedPackage struct {
	Names      util.Set[string]
	HasNotable bool
}

type FileAnalyzer struct {
	basePkg          string
	notableFileNames util.Set[string]

	moduleDir  string
	patterns   []string
	depth      int
	buildFlags []string
	miscUsages []MiscUsage
	testAll    bool

	pkgDirs          map[string]string
	testFuncs        util.Set[*types.Func]
	definitions      map[string]*definition
	pkgTestUniqNames map[string]util.Set[string]
	pkgObjNames      map[string]util.Set[string]
	pkgLocalObjNames map[string]map[string]string
	fileObjNames     map[string]util.Set[string]
}

var defaultOptions = []Option{
	WithDepth(1),
	WithPatterns("./..."),
}

func NewFileAnalyzer(basePkg string, notableFileNames []string, options ...Option) *FileAnalyzer {
	fa := &FileAnalyzer{
		basePkg:          basePkg,
		notableFileNames: util.SetFrom(notableFileNames),
		pkgDirs:          make(map[string]string),
		testFuncs:        util.NewSet[*types.Func](),
		definitions:      make(map[string]*definition),
		pkgTestUniqNames: make(map[string]util.Set[string]),
		pkgObjNames:      make(map[string]util.Set[string]),
		pkgLocalObjNames: make(map[string]map[string]string),
		fileObjNames:     make(map[string]util.Set[string]),
	}

	fa.applyOptions(defaultOptions)
	fa.applyOptions(options)

	return fa
}

func (fa *FileAnalyzer) applyOptions(options []Option) {
	for _, option := range options {
		option(fa)
	}
}

func (fa *FileAnalyzer) addDefinition(obj types.Object, fileName string, node ast.Node) {
	fa.definitions[types.ObjectString(obj, nil)] = &definition{
		obj:            obj,
		fileName:       fileName,
		node:           node,
		usedByObjNames: util.NewSet[string](),
		usingObjNames:  util.NewSet[string](),
	}
}

func (fa *FileAnalyzer) getDefinition(obj types.Object) *definition {
	return fa.definitions[types.ObjectString(obj, nil)]
}

func (fa *FileAnalyzer) addObj(pkgPath, fileName string, obj types.Object) {
	objName := types.ObjectString(obj, nil)

	pkgObjs := util.MapGetOrCreate(fa.pkgObjNames, pkgPath, func() util.Set[string] { return util.NewSet[string]() })
	pkgObjs.Add(objName)

	pkgLocalObjs := util.MapGetOrCreate(fa.pkgLocalObjNames, pkgPath, func() map[string]string { return make(map[string]string) })
	pkgLocalObjs[obj.Name()] = objName

	fileObjs := util.MapGetOrCreate(fa.fileObjNames, fileName, func() util.Set[string] { return util.NewSet[string]() })
	fileObjs.Add(objName)
}

func (fa *FileAnalyzer) Load() error {
	pkgs, err := packages.Load(&packages.Config{
		Dir: fa.moduleDir,
		Mode: packages.NeedCompiledGoFiles |
			packages.NeedDeps |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedName |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		BuildFlags: fa.buildFlags,
		Tests:      true,
	}, fa.patterns...)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		fa.addPkgPath(pkg)
	}

	for _, pkg := range pkgs {
		fa.searchTopLevelObjects(pkg)
	}

	for _, pkg := range pkgs {
		fa.analyzeUses(pkg)
		fa.analyzeDefs(pkg)
		fa.analyzeImplicits(pkg)
	}

	return nil
}

func (fa *FileAnalyzer) addPkgPath(pkg *packages.Package) {
	pkgPath := pkg.PkgPath
	if strings.HasSuffix(pkgPath, ".test]") || strings.HasSuffix(pkgPath, ".test") {
		return
	}
	pkgPath = strings.TrimSuffix(pkg.PkgPath, "_test")

	var dir string
	if len(pkg.GoFiles) >= 1 {
		dir = filepath.Dir(pkg.GoFiles[0])
	}

	fa.pkgDirs[pkgPath] = dir
}

func (fa *FileAnalyzer) searchTopLevelObjects(pkg *packages.Package) {
	// Collect all nodes from top level declarations.
	// There aren't any good way to obtain AST position from object.
	nodes := make([]ast.Node, 0)
	for _, astFile := range pkg.Syntax {
		for _, d := range astFile.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				nodes = append(nodes, decl)
			case *ast.GenDecl:
				for _, s := range decl.Specs {
					if ts, ok := s.(*ast.TypeSpec); ok {
						if iface, ok := ts.Type.(*ast.InterfaceType); ok {
							nodes = append(nodes, ts.Name)
							for _, method := range iface.Methods.List {
								nodes = append(nodes, method)
							}
						} else {
							nodes = append(nodes, s)
						}
					} else {
						nodes = append(nodes, s)
					}
				}
			}
		}
	}

	for ident, defObj := range pkg.TypesInfo.Defs {
		if defObj == nil {
			continue
		}

		isMemberFunc := false
		switch obj := defObj.(type) {
		case *types.Func:
			sig := obj.Type().(*types.Signature)

			if sig.Recv() != nil {
				switch sig.Recv().Type().(type) {
				case *types.Pointer, *types.Named, *types.Interface:
					isMemberFunc = true
				}
			}
		}

		// Ignore non top-level objects that are not methods of a struct/interface.
		if !isMemberFunc && defObj.Parent() != pkg.Types.Scope() {
			continue
		}

		file := pkg.Fset.File(ident.Pos())

		// Prevent object definitions from cache files.
		if util.IsWithinPath(util.GoCacheFolder(), file.Name()) {
			continue
		}

		// Record test files.
		if strings.HasSuffix(file.Name(), "_test.go") {
			if f, ok := defObj.(*types.Func); ok && strings.HasPrefix(f.Name(), "Test") && f.Name() != "TestMain" {
				fa.testFuncs.Add(f)
				uniqNames := util.MapGetOrCreate(fa.pkgTestUniqNames, strings.TrimSuffix(pkg.PkgPath, "_test"), func() util.Set[string] {
					return util.NewSet[string]()
				})
				uniqNames.Add(f.Name())
			}
		}

		var node ast.Node
		for _, d := range nodes {
			if d.Pos() <= ident.Pos() && ident.End() <= d.End() {
				node = d
				break
			}
		}

		fa.addDefinition(defObj, file.Name(), node)
		fa.addObj(strings.TrimSuffix(pkg.PkgPath, "_test"), file.Name(), defObj)
	}
}

func (fa *FileAnalyzer) analyzeUses(pkg *packages.Package) {
	for ident, usedObj := range pkg.TypesInfo.Uses {
		fa.addUsage(pkg.Fset, ident.Pos(), usedObj)
	}
}

func (fa *FileAnalyzer) analyzeDefs(pkg *packages.Package) {
	for ident, defObj := range pkg.TypesInfo.Defs {
		fa.addUsageToObjectType(pkg.Fset, ident.Pos(), defObj)
	}
}

func (fa *FileAnalyzer) analyzeImplicits(pkg *packages.Package) {
	for node, implicitObj := range pkg.TypesInfo.Implicits {
		fa.addUsageToObjectType(pkg.Fset, node.Pos(), implicitObj)
	}
}

func (fa *FileAnalyzer) addUsageToObjectType(fset *token.FileSet, usagePos token.Pos, obj types.Object) {
	if obj == nil || obj.Pkg() == nil {
		return
	}

	// Ignore package names.
	if _, ok := obj.(*types.PkgName); ok {
		return
	}

	for _, usedTypeName := range getUsedTypeNames(obj.Type()) {
		fa.addUsage(fset, usagePos, usedTypeName)
	}
}

func (fa *FileAnalyzer) addUsage(fset *token.FileSet, usagePos token.Pos, usedObj types.Object) {
	if usedObj == nil {
		return
	}

	file := fset.File(usagePos)

	// Prevent object definitions from cache files.
	if util.IsWithinPath(util.GoCacheFolder(), file.Name()) {
		return
	}

	// Prevent usages from outside the main package in question.
	// Nil package indicate native objects.
	if usedObj.Pkg() == nil || !util.IsSubPackage(fa.basePkg, usedObj.Pkg().Path()) {
		return
	}

	usedDef := fa.getDefinition(usedObj)
	// Could be using some non top-level objects.
	if usedDef == nil {
		return
	}

	usedObjName := types.ObjectString(usedObj, nil)

	for objName := range fa.fileObjNames[file.Name()] {
		// Prevent self-usage.
		if objName == usedObjName {
			continue
		}

		def := fa.definitions[objName]

		// Ignore objects not within the user object.
		if usagePos < def.node.Pos() || def.node.End() < usagePos {
			continue
		}

		def.usingObjNames.Add(usedObjName)
		usedDef.usedByObjNames.Add(objName)
	}
}

func (fa *FileAnalyzer) DetermineTests() (map[string]*TestedPackage, int) {
	testedPkgs := make(map[string]*TestedPackage)
	if fa.testAll {
		for pkgPath := range fa.pkgDirs {
			testedPkgs[pkgPath] = &TestedPackage{
				Names:      util.NewSet("*"),
				HasNotable: true,
			}
		}
		return testedPkgs, -1
	}

	fa.testsFromUsages(testedPkgs)

	uniqueTestCount := 0

	// Consolidate test packages that test everything.
	for pkgPath, testedPkg := range testedPkgs {
		uniqueTestCount += testedPkg.Names.Len()
		if fa.pkgTestUniqNames != nil && testedPkg.Names.Len() == fa.pkgTestUniqNames[pkgPath].Len() {
			testedPkgs[pkgPath].Names = util.NewSet("*")
		}
	}

	return testedPkgs, uniqueTestCount
}

func (fa *FileAnalyzer) testsFromUsages(testedPkgs map[string]*TestedPackage) {
	// Multi-source BFS.
	queued := make(map[string]*traversal)
	queue := make(traversalPQ, 0)

	notablePkgs := util.NewSet[string]()

	fa.queueUp(func(objName string) {
		if _, ok := queued[objName]; ok {
			return
		}

		t := &traversal{
			objName:   objName,
			stepsLeft: fa.depth,
		}
		heap.Push(&queue, t)
		queued[objName] = t

		def := fa.definitions[t.objName]
		if def == nil {
			return
		}
		pkg := strings.TrimSuffix(def.obj.Pkg().Path(), "_test")
		notablePkgs.Add(pkg)
	})

	for queue.Len() > 0 {
		t := heap.Pop(&queue).(*traversal)

		def := fa.definitions[t.objName]
		if def == nil {
			continue
		}

		if f, ok := def.obj.(*types.Func); ok && fa.testFuncs.Has(f) {
			pkg := strings.TrimSuffix(f.Pkg().Path(), "_test")

			testedPkg := util.MapGetOrCreate(testedPkgs, pkg, func() *TestedPackage {
				return &TestedPackage{
					Names:      util.NewSet[string](),
					HasNotable: false,
				}
			})
			testedPkg.Names.Add(f.Name())

			if notablePkgs.Has(pkg) {
				testedPkg.HasNotable = true
			}
		}

		if t.stepsLeft <= 0 {
			continue
		}
		nextStepsLeft := t.stepsLeft - 1

		for userObjName := range def.usedByObjNames {
			nt, ok := queued[userObjName]
			if !ok {
				nt = &traversal{
					objName:   userObjName,
					stepsLeft: nextStepsLeft,
				}
				heap.Push(&queue, nt)
				queued[userObjName] = nt
			} else if nt.stepsLeft < nextStepsLeft {
				nt.stepsLeft = nextStepsLeft
				heap.Fix(&queue, nt.index)
			}
		}
	}
}

func (fa *FileAnalyzer) queueUp(addToQueue func(string)) {
	for notableFileName := range fa.notableFileNames {
		for objName := range fa.fileObjNames[notableFileName] {
			addToQueue(objName)
		}

		// Add all that are related to misc usage.
		for _, miscUsage := range fa.miscUsages {
			if !miscUsage.Regexp.MatchString(notableFileName) {
				continue
			}

			for _, user := range miscUsage.UsedBy {
				// Is recursive?
				if strings.HasSuffix(user.PkgPath, "/...") {
					for pkgPath, objNames := range fa.pkgObjNames {
						if !strings.HasPrefix(pkgPath, user.PkgPath[:len(user.PkgPath)-4]) {
							continue
						}
						for objName := range objNames {
							addToQueue(objName)
						}
					}

					continue
				}

				if user.All {
					for objName := range fa.pkgObjNames[user.PkgPath] {
						addToQueue(objName)
					}

					continue
				}

				for _, fileName := range user.FileNames {
					for objName := range fa.fileObjNames[filepath.Join(fa.pkgDirs[user.PkgPath], fileName)] {
						addToQueue(objName)
					}
				}

				for _, localObjName := range user.ObjNames {
					if fa.pkgLocalObjNames[user.PkgPath] == nil {
						continue
					}
					fullObjName, ok := fa.pkgLocalObjNames[user.PkgPath][localObjName]
					if !ok {
						continue
					}
					addToQueue(fullObjName)
				}
			}
		}
	}
}

func (fa *FileAnalyzer) MarshalJSON() ([]byte, error) {
	type jsonDefinition struct {
		File   string           `json:"file"`
		UsedBy util.Set[string] `json:"usedBy"`
		Using  util.Set[string] `json:"using"`
	}

	type jsonAnalyzer struct {
		TestFuncs   []string                    `json:"testFuncs"`
		Definitions map[string]jsonDefinition   `json:"definitions"`
		FileObjs    map[string]util.Set[string] `json:"fileObjs"`
	}

	x := jsonAnalyzer{
		TestFuncs:   make([]string, 0, len(fa.testFuncs)),
		Definitions: make(map[string]jsonDefinition, len(fa.definitions)),
		FileObjs:    make(map[string]util.Set[string], len(fa.fileObjNames)),
	}
	for testFunc := range fa.testFuncs {
		x.TestFuncs = append(x.TestFuncs, testFunc.FullName())
	}

	for objName, def := range fa.definitions {
		y := jsonDefinition{
			File:   def.fileName,
			UsedBy: util.NewSet[string](),
			Using:  util.NewSet[string](),
		}
		for userObjName := range def.usedByObjNames {
			y.UsedBy.Add(userObjName)
		}
		for usedObjName := range def.usingObjNames {
			y.Using.Add(usedObjName)
		}
		x.Definitions[objName] = y
	}

	for fileName, objNames := range fa.fileObjNames {
		y := util.NewSet[string]()
		for objName := range objNames {
			y.Add(objName)
		}
		x.FileObjs[fileName] = y
	}

	return json.Marshal(x)
}
