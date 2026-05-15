// v2/ivf.go

package v2

import (
	"container/heap"
	"math/rand"
	"sort"
	"time"
)

type IVFIndex struct {
	Data       []Record
	Centroids  []Vector
	Inverted   [][]int
	NProbe     int
	NumCluster int
}

func NewIVFIndex(
	path string,
	numClusters int,
	nprobe int,
) (*IVFIndex, error) {

	data, err := LoadDataset(path)

	if err != nil {
		return nil, err
	}

	return BuildIVF(
		data,
		numClusters,
		nprobe,
	), nil
}

func BuildIVF(
	data []Record,
	numClusters int,
	nprobe int,
) *IVFIndex {

	rand.Seed(time.Now().UnixNano())

	centroids := make([]Vector, numClusters)

	for i := 0; i < numClusters; i++ {

		centroids[i] = data[rand.Intn(len(data))].Vec
	}

	inverted := make([][]int, numClusters)

	for dataIdx, r := range data {

		bestCluster := 0

		bestScore := float32(-1e9)

		for clusterIdx, centroid := range centroids {

			score := cosine(r.Vec, centroid)

			if score > bestScore {

				bestScore = score
				bestCluster = clusterIdx
			}
		}

		inverted[bestCluster] = append(
			inverted[bestCluster],
			dataIdx,
		)
	}

	return &IVFIndex{
		Data:       data,
		Centroids:  centroids,
		Inverted:   inverted,
		NProbe:     nprobe,
		NumCluster: numClusters,
	}
}

func (ivf *IVFIndex) Search(
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

	h := &MinHeap{}

	heap.Init(h)

	for i := 0; i < ivf.NProbe &&
		i < len(clusterScores); i++ {

		clusterID := clusterScores[i].Index

		for _, dataIdx := range ivf.Inverted[clusterID] {

			r := ivf.Data[dataIdx]

			score := cosine(query, r.Vec)

			if h.Len() < k {

				heap.Push(h, Neighbor{
					Score: score,
					Fraud: r.Fraud,
				})

			} else if score > (*h)[0].Score {

				(*h)[0] = Neighbor{
					Score: score,
					Fraud: r.Fraud,
				}

				heap.Fix(h, 0)
			}
		}
	}

	result := make([]Neighbor, h.Len())

	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(Neighbor)
	}

	return result
}
