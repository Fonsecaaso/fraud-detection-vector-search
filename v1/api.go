package v1

import (
	"encoding/json"
	"net/http"
)

//
// =====================================================
// GLOBAL INDEXES
// =====================================================
//

var index *Index
var hnsw *HNSWIndex
var ivf *IVFIndex

//
// =====================================================
// INIT
// =====================================================
//

func Initialize() {

	index = NewIndex(1_000_000)

	hnsw = NewHNSWIndex(index, 8)

	ivf = NewIVFIndex(
		index,
		1024,
		6,
	)
}

//
// =====================================================
// REQUEST / RESPONSE
// =====================================================
//

type PredictRequest struct {
	Vector Vector `json:"vector"`
	Method string `json:"method"`
}

type NeighborResponse struct {
	Score float32 `json:"score"`
	Fraud bool    `json:"fraud"`
}

type PredictResponse struct {
	Fraud     bool               `json:"fraud"`
	Neighbors []NeighborResponse `json:"neighbors"`
}

//
// =====================================================
// HANDLER
// =====================================================
//

func Predict(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {

		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	var req PredictRequest

	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {

		http.Error(
			w,
			"invalid request body: "+err.Error(),
			http.StatusBadRequest,
		)

		return
	}

	//
	// validate dimensions
	//

	if len(req.Vector) != 3 {

		http.Error(
			w,
			"vector must have 3 dimensions",
			http.StatusBadRequest,
		)

		return
	}

	//
	// search
	//

	var neighbors []Neighbor

	switch req.Method {

	case "optimized":

		neighbors = index.SearchOptimized(
			req.Vector,
			5,
		)

	case "parallel":

		neighbors = index.SearchParallel(
			req.Vector,
			5,
		)

	case "hnsw":

		neighbors = hnsw.SearchHNSW(
			req.Vector,
			5,
		)

	case "ivf":

		neighbors = ivf.SearchIVFParallel(
			req.Vector,
			5,
		)

	default:

		neighbors = index.Search(
			req.Vector,
			5,
		)
	}

	//
	// fraud decision
	//

	fraudCount := 0

	respNeighbors := make(
		[]NeighborResponse,
		0,
		len(neighbors),
	)

	for _, n := range neighbors {

		if n.Fraud {
			fraudCount++
		}

		respNeighbors = append(
			respNeighbors,
			NeighborResponse{
				Score: n.Score,
				Fraud: n.Fraud,
			},
		)
	}

	isFraud := fraudCount >= 3

	//
	// response
	//

	resp := PredictResponse{
		Fraud:     isFraud,
		Neighbors: respNeighbors,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(resp)
}
