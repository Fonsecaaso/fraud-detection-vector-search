package v1

import "container/heap"

type HNSWNode struct {
	Vec       Vector
	Fraud     bool
	Neighbors []int
}

type HNSWIndex struct {
	Nodes      []HNSWNode
	EntryPoint int
	M          int // conexões por nó
}

func NewHNSWIndex(idx *Index, m int) *HNSWIndex {
	h := &HNSWIndex{
		Nodes:      make([]HNSWNode, 0, len(idx.Data)),
		EntryPoint: -1,
		M:          m,
	}

	for _, r := range idx.Data {
		h.Add(r.Vec, r.Fraud)
	}

	return h
}

func (h *HNSWIndex) Add(vec Vector, fraud bool) {
	newIdx := len(h.Nodes)

	node := HNSWNode{
		Vec:       vec,
		Fraud:     fraud,
		Neighbors: []int{},
	}

	// primeiro nó
	if h.EntryPoint == -1 {
		h.EntryPoint = newIdx
		h.Nodes = append(h.Nodes, node)
		return
	}

	// busca vizinhos próximos
	neighbors := h.searchLayer(vec, h.EntryPoint, h.M*2)

	for i, n := range neighbors {
		if i >= h.M {
			break
		}

		node.Neighbors = append(node.Neighbors, n.Index)
		h.Nodes[n.Index].Neighbors = append(h.Nodes[n.Index].Neighbors, newIdx)
	}

	h.Nodes = append(h.Nodes, node)
}

func (h *HNSWIndex) SearchHNSW(query Vector, k int) []Neighbor {
	hres := h.searchLayer(query, h.EntryPoint, k)

	result := make([]Neighbor, len(hres))
	for i, n := range hres {
		result[i] = Neighbor{
			Score: n.Score,
			Fraud: n.Fraud,
		}
	}

	return result
}

type HNeighbor struct {
	Score float32
	Index int
	Fraud bool
}

// max heap (para best)
type MaxHeap []HNeighbor

func (h MaxHeap) Len() int           { return len(h) }
func (h MaxHeap) Less(i, j int) bool { return h[i].Score < h[j].Score }
func (h MaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x interface{}) {
	*h = append(*h, x.(HNeighbor))
}

func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func (h *HNSWIndex) searchLayer(query Vector, entry int, k int) []HNeighbor {
	visited := make(map[int]bool)

	candidates := &MaxHeap{}
	best := &MaxHeap{}

	heap.Init(candidates)
	heap.Init(best)

	startScore := cosine(query, h.Nodes[entry].Vec)

	start := HNeighbor{
		Score: startScore,
		Index: entry,
		Fraud: h.Nodes[entry].Fraud,
	}

	heap.Push(candidates, start)
	heap.Push(best, start)

	visited[entry] = true

	for candidates.Len() > 0 {
		current := heap.Pop(candidates).(HNeighbor)

		// early stop
		if best.Len() >= k && current.Score < (*best)[0].Score {
			break
		}

		for _, nb := range h.Nodes[current.Index].Neighbors {
			if visited[nb] {
				continue
			}
			visited[nb] = true

			score := cosine(query, h.Nodes[nb].Vec)

			n := HNeighbor{
				Score: score,
				Index: nb,
				Fraud: h.Nodes[nb].Fraud,
			}

			if best.Len() < k || score > (*best)[0].Score {
				heap.Push(best, n)

				if best.Len() > k {
					heap.Pop(best)
				}

				heap.Push(candidates, n)
			}
		}
	}

	// converter para saída
	result := make([]HNeighbor, best.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(best).(HNeighbor)
	}

	return result
}
