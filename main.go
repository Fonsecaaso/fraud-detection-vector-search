package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var index *Index    // global simples pra POC
var hnsw *HNSWIndex // global simples pra POC
var ivf *IVFIndex   // global simples pra POC

// ======= REQUEST / RESPONSE =======

type Request struct {
	Vector Vector `json:"vector"`
	Method string `json:"method"`
}

type NeighborResponse struct {
	Score float32 `json:"score"`
	Fraud bool    `json:"fraud"`
}

type Response struct {
	Fraud     bool               `json:"fraud"`
	Neighbors []NeighborResponse `json:"neighbors"`
}

// ======= HANDLER =======

func predict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// valida dimensão
	if len(req.Vector) != 3 {
		http.Error(w, "vector must have 3 dimensions", http.StatusBadRequest)
		return
	}

	var neighbors []Neighbor
	switch req.Method {
	case "optimized":
		neighbors = index.SearchOptimized(req.Vector, 5)
	case "parallel":
		neighbors = index.SearchParallel(req.Vector, 5)
	case "hsnw":
		neighbors = hnsw.SearchHNSW(req.Vector, 5)
	// case "ivf":
	// 	neighbors = ivf.SearchIVFParallel(req.Vector, 5)
	default:
		neighbors = index.Search(req.Vector, 5)
	}

	fraudCount := 0
	respNeighbors := make([]NeighborResponse, 0, len(neighbors))

	for _, n := range neighbors {
		if n.Fraud {
			fraudCount++
		}

		respNeighbors = append(respNeighbors, NeighborResponse{
			Score: n.Score,
			Fraud: n.Fraud,
		})
	}

	isFraud := fraudCount >= 3

	resp := Response{
		Fraud:     isFraud,
		Neighbors: respNeighbors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ======= MAIN =======

func main() {
	// inicializa index mockado
	fmt.Println("Initializing indexes...")
	index = NewIndex(1_000_000)
	hnsw = NewHNSWIndex(index, 8)
	// ivf = NewIVFIndex(index, 1024, 6)
	fmt.Println("Indexes initialized")

	mux := http.NewServeMux()
	mux.HandleFunc("/fraud-score", predict)

	addr := ":9999"
	log.Printf("Servidor ouvindo em %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
