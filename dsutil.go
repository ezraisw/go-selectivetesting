package selectivetesting

import (
	"strings"

	"github.com/pwnedgod/go-selectivetesting/internal/util"
)

type traversal struct {
	objName   string
	stepsLeft int
	index     int
}

type traversalPQ []*traversal

func (h traversalPQ) Len() int {
	return len(h)
}

func (h traversalPQ) Less(i, j int) bool {
	return h[i].stepsLeft > h[j].stepsLeft
}

func (h traversalPQ) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *traversalPQ) Push(x any) {
	index := len(*h)
	t := x.(*traversal)
	t.index = index
	*h = append(*h, t)
}

func (h *traversalPQ) Pop() any {
	index := len(*h) - 1
	t := (*h)[index]
	t.index = -1
	(*h)[index] = nil
	*h = (*h)[:index]
	return t
}

type trieNode struct {
	children map[string]*trieNode
	names    util.Set[string]
}

func consolidateTests(testedPkgs map[string]util.Set[string]) map[string]util.Set[string] {
	trieRoot := &trieNode{children: make(map[string]*trieNode)}

	for testedPkg, testNames := range testedPkgs {
		pieces := strings.Split(testedPkg, "/")

		trieCurr := trieRoot
		for _, piece := range pieces {
			child, ok := trieCurr.children[piece]
			if !ok {
				child = &trieNode{children: make(map[string]*trieNode)}
				trieCurr.children[piece] = child
			}
			trieCurr = child
		}

		trieCurr.names = testNames
	}

	newTestedPkgs := make(map[string]util.Set[string])
	traverseTrie("", trieRoot, newTestedPkgs)
	return newTestedPkgs
}

func traverseTrie(path string, curr *trieNode, testedPkgs map[string]util.Set[string]) {
	if child, ok := curr.children["..."]; ok {
		testedPkgs[path+"/..."] = child.names
		return
	}
	for piece, child := range curr.children {
		nextPath := piece
		if path != "" {
			nextPath = path + "/" + piece
		}
		if child.names != nil {
			testedPkgs[nextPath] = child.names
			continue
		}
		traverseTrie(nextPath, child, testedPkgs)
	}
}
