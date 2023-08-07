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
	exists   bool
	names    util.Set[string]
}

func consolidateTests(trieRoot *trieNode, testedPkgs map[string]util.Set[string]) map[string]util.Set[string] {
	for testedPkg, testNames := range testedPkgs {
		pieces := strings.Split(testedPkg, "/")

		cutoff := false
		trieCurr := trieRoot
		for _, piece := range pieces {
			child, ok := trieCurr.children[piece]
			if !ok {
				if piece != "..." {
					cutoff = true
					break
				}

				// We just need "..." to be not nil. It should not have a children.
				child = &trieNode{}
				trieCurr.children[piece] = child
			}
			trieCurr = child
		}

		if !cutoff {
			trieCurr.names = testNames
		}
	}

	newTestedPkgs := make(map[string]util.Set[string])
	traverseTrie("", trieRoot, false, newTestedPkgs)
	return newTestedPkgs
}

func traverseTrie(path string, curr *trieNode, alwaysAdd bool, testedPkgs map[string]util.Set[string]) {
	if _, ok := curr.children["..."]; ok {
		alwaysAdd = true
	}
	for piece, child := range curr.children {
		// Do not handle "...".
		// It is technically just a marker so that everything below this parent gets added.
		if piece == "..." {
			continue
		}

		nextPath := piece
		if path != "" {
			nextPath = path + "/" + piece
		}

		if child.exists {
			if alwaysAdd {
				testedPkgs[nextPath] = util.NewSet("*")
			} else if child.names != nil {
				testedPkgs[nextPath] = child.names
			}
		}

		traverseTrie(nextPath, child, alwaysAdd, testedPkgs)
	}
}
