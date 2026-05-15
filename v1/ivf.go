package v1

import (
	"container/heap"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"time"
)

//
// =====================================================
// IVF INDEX (V1 MOCK)
// =====================================================
//
// - vetores mockados
// - 3 dimensões
// - mesma filosofia do HNSWIndex
//

//
// =========================
// HEAP
// =========================
//

type MinHeap []Neighbor

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Score < h[j].Score }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

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

//
// =========================
// IVF TYPES
// =========================
//

type IVFNode struct {
	Vec       Vector
	Fraud     bool
	ClusterID int
}

type IVFIndex struct {
	Nodes       []IVFNode
	Centroids   []Vector
	Inverted    [][]int
	NProbe      int
	NumClusters int
}

//
// =========================
// CONSTRUCTOR
// =========================
//

func NewIVFIndex(
	idx *Index,
	numClusters int,
	nprobe int,
) *IVFIndex {

	rand.Seed(time.Now().UnixNano())

	ivf := &IVFIndex{
		Nodes:       make([]IVFNode, 0, len(idx.Data)),
		Centroids:   make([]Vector, numClusters),
		Inverted:    make([][]int, numClusters),
		NProbe:      nprobe,
		NumClusters: numClusters,
	}

	//
	// init centroides aleatórios
	//

	for i := 0; i < numClusters; i++ {
		ivf.Centroids[i] = idx.Data[rand.Intn(len(idx.Data))].Vec
	}

	//
	// add nodes
	//

	for _, r := range idx.Data {
		ivf.Add(r.Vec, r.Fraud)
	}

	return ivf
}

//
// =========================
// ADD
// =========================
//

func (ivf *IVFIndex) Add(
	vec Vector,
	fraud bool,
) {

	bestCluster := 0

	bestScore := float32(-1e9)

	//
	// nearest centroid
	//

	for clusterIdx, centroid := range ivf.Centroids {

		score := cosine(vec, centroid)

		if score > bestScore {
			bestScore = score
			bestCluster = clusterIdx
		}
	}

	//
	// create node
	//

	nodeIdx := len(ivf.Nodes)

	node := IVFNode{
		Vec:       vec,
		Fraud:     fraud,
		ClusterID: bestCluster,
	}

	ivf.Nodes = append(ivf.Nodes, node)

	//
	// inverted list
	//

	ivf.Inverted[bestCluster] = append(
		ivf.Inverted[bestCluster],
		nodeIdx,
	)
}

//
// =========================
// SEARCH
// =========================
//

func (ivf *IVFIndex) SearchIVF(
	query Vector,
	k int,
) []Neighbor {

	type ClusterScore struct {
		Index int
		Score float32
	}

	//
	// top centroids
	//

	clusterScores := make(
		[]ClusterScore,
		0,
		len(ivf.Centroids),
	)

	for i, centroid := range ivf.Centroids {

		score := cosine(query, centroid)

		clusterScores = append(
			clusterScores,
			ClusterScore{
				Index: i,
				Score: score,
			},
		)
	}

	sort.Slice(clusterScores, func(i, j int) bool {
		return clusterScores[i].Score >
			clusterScores[j].Score
	})

	//
	// local search
	//

	h := &MinHeap{}

	heap.Init(h)

	for i := 0; i < ivf.NProbe &&
		i < len(clusterScores); i++ {

		clusterID := clusterScores[i].Index

		for _, nodeIdx := range ivf.Inverted[clusterID] {

			node := ivf.Nodes[nodeIdx]

			score := cosine(query, node.Vec)

			if h.Len() < k {

				heap.Push(h, Neighbor{
					Score: score,
					Fraud: node.Fraud,
				})

			} else if score > (*h)[0].Score {

				(*h)[0] = Neighbor{
					Score: score,
					Fraud: node.Fraud,
				}

				heap.Fix(h, 0)
			}
		}
	}

	//
	// result
	//

	result := make([]Neighbor, h.Len())

	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(Neighbor)
	}

	return result
}

//
// =========================
// PARALLEL SEARCH
// =========================
//

func (ivf *IVFIndex) SearchIVFParallel(
	query Vector,
	k int,
) []Neighbor {

	type ClusterScore struct {
		Index int
		Score float32
	}

	clusterScores := make(
		[]ClusterScore,
		0,
		len(ivf.Centroids),
	)

	for i, centroid := range ivf.Centroids {

		score := cosine(query, centroid)

		clusterScores = append(
			clusterScores,
			ClusterScore{
				Index: i,
				Score: score,
			},
		)
	}

	sort.Slice(clusterScores, func(i, j int) bool {
		return clusterScores[i].Score >
			clusterScores[j].Score
	})

	//
	// top clusters
	//

	topClusters := clusterScores[:ivf.NProbe]

	//
	// workers
	//

	numWorkers := runtime.NumCPU()

	chunkSize := (len(topClusters) + numWorkers - 1) / numWorkers

	results := make(chan []Neighbor, numWorkers)

	var wg sync.WaitGroup

	worker := func(clusters []ClusterScore) {

		defer wg.Done()

		h := &MinHeap{}

		heap.Init(h)

		for _, cluster := range clusters {

			clusterID := cluster.Index

			for _, nodeIdx := range ivf.Inverted[clusterID] {

				node := ivf.Nodes[nodeIdx]

				score := cosine(query, node.Vec)

				if h.Len() < k {

					heap.Push(h, Neighbor{
						Score: score,
						Fraud: node.Fraud,
					})

				} else if score > (*h)[0].Score {

					(*h)[0] = Neighbor{
						Score: score,
						Fraud: node.Fraud,
					}

					heap.Fix(h, 0)
				}
			}
		}

		local := make([]Neighbor, h.Len())

		for i := len(local) - 1; i >= 0; i-- {
			local[i] = heap.Pop(h).(Neighbor)
		}

		results <- local
	}

	//
	// spawn workers
	//

	for i := 0; i < numWorkers; i++ {

		start := i * chunkSize

		end := start + chunkSize

		if start >= len(topClusters) {
			break
		}

		if end > len(topClusters) {
			end = len(topClusters)
		}

		wg.Add(1)

		go worker(topClusters[start:end])
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	//
	// merge final
	//

	finalHeap := &MinHeap{}

	heap.Init(finalHeap)

	for local := range results {

		for _, n := range local {

			if finalHeap.Len() < k {

				heap.Push(finalHeap, n)

			} else if n.Score > (*finalHeap)[0].Score {

				(*finalHeap)[0] = n

				heap.Fix(finalHeap, 0)
			}
		}
	}

	result := make([]Neighbor, finalHeap.Len())

	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(finalHeap).(Neighbor)
	}

	return result
}
