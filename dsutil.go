package selectivetesting

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
