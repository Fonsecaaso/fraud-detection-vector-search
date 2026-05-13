package v1

import (
	"container/heap"
	"runtime"
	"sync"
)

func (idx *Index) SearchOptimized(query Vector, k int) []Neighbor {
	h := &MinHeap{}
	heap.Init(h)

	for _, r := range idx.Data {
		score := cosine(query, r.Vec)

		if h.Len() < k {
			heap.Push(h, Neighbor{Score: score, Fraud: r.Fraud})
		} else if score > (*h)[0].Score {
			(*h)[0] = Neighbor{Score: score, Fraud: r.Fraud}
			heap.Fix(h, 0)
		}
	}

	// extrai resultados
	result := make([]Neighbor, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(Neighbor)
	}

	return result
}

func (idx *Index) SearchParallel(query Vector, k int) []Neighbor {
	numWorkers := runtime.NumCPU()
	chunkSize := (len(idx.Data) + numWorkers - 1) / numWorkers

	results := make(chan []Neighbor, numWorkers)
	var wg sync.WaitGroup

	// worker
	worker := func(start, end int) {
		defer wg.Done()

		h := &MinHeap{}
		heap.Init(h)

		for i := start; i < end && i < len(idx.Data); i++ {
			r := idx.Data[i]
			score := cosine(query, r.Vec)

			if h.Len() < k {
				heap.Push(h, Neighbor{Score: score, Fraud: r.Fraud})
			} else if score > (*h)[0].Score {
				(*h)[0] = Neighbor{Score: score, Fraud: r.Fraud}
				heap.Fix(h, 0)
			}
		}

		// converte heap em slice
		local := make([]Neighbor, h.Len())
		for i := len(local) - 1; i >= 0; i-- {
			local[i] = heap.Pop(h).(Neighbor)
		}

		results <- local
	}

	// spawn workers
	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize

		wg.Add(1)
		go worker(start, end)
	}

	// fechar canal quando terminar
	go func() {
		wg.Wait()
		close(results)
	}()

	// merge final (top-K global)
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

	// extrai resultado final
	result := make([]Neighbor, finalHeap.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(finalHeap).(Neighbor)
	}

	return result
}
