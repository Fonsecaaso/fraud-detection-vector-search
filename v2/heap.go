// v2/heap.go

package v2

type MinHeap []Neighbor

func (h MinHeap) Len() int {
	return len(h)
}

func (h MinHeap) Less(i, j int) bool {
	return h[i].Score < h[j].Score
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(Neighbor))
}

func (h *MinHeap) Pop() interface{} {

	old := *h

	n := len(old)

	x := old[n-1]

	*h = old[:n-1]

	return x
}
