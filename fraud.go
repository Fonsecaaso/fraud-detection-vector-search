package main

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

type Vector []float32

type Record struct {
	Vec   Vector
	Fraud bool
}

type Index struct {
	Data []Record
}

// cosine similarity
func cosine(a, b Vector) float32 {
	var dot, normA, normB float32

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

type Neighbor struct {
	Score float32
	Fraud bool
}

func (idx *Index) Search(query Vector, k int) []Neighbor {
	neighbors := make([]Neighbor, 0, len(idx.Data))

	for _, r := range idx.Data {
		score := cosine(query, r.Vec)
		neighbors = append(neighbors, Neighbor{
			Score: score,
			Fraud: r.Fraud,
		})
	}

	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].Score > neighbors[j].Score
	})

	return neighbors[:k]
}

func Predict(idx *Index, query Vector) bool {
	top := idx.Search(query, 5)

	frauds := 0
	for _, n := range top {
		if n.Fraud {
			frauds++
		}
	}

	return frauds >= 3
}

// gera número float pequeno em volta de um centro
func jitter(center float32) float32 {
	return center + float32(rand.NormFloat64()*0.1)
}

func NewIndex(size int) *Index {
	rand.Seed(time.Now().UnixNano())

	data := make([]Record, 0, size)

	for i := 0; i < size; i++ {
		var vec Vector
		var fraud bool

		// 20% fraude
		if rand.Float32() < 0.2 {
			// cluster fraude (1,1,1)
			vec = Vector{
				jitter(1),
				jitter(1),
				jitter(1),
			}
			fraud = true
		} else {
			// cluster legit (0,0,0)
			vec = Vector{
				jitter(0),
				jitter(0),
				jitter(0),
			}
			fraud = false
		}

		data = append(data, Record{
			Vec:   vec,
			Fraud: fraud,
		})
	}

	return &Index{
		Data: data,
	}
}
