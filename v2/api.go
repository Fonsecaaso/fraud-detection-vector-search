package v2

import (
	"encoding/json"
	"net/http"
)

var ivf *IVFIndex

func Initialize() error {

	index, err := NewIVFIndex(
		"./datasets/references.json",
		1024,
		8,
	)

	if err != nil {
		return err
	}

	ivf = index

	return nil
}

type FraudResponse struct {
	Approved   bool    `json:"approved"`
	FraudScore float32 `json:"fraud_score"`
}

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

	var req FraudRequest

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

	vector := BuildVector(req)

	neighbors := ivf.Search(
		vector,
		5,
	)

	var fraudWeight float32
	var totalWeight float32

	for _, n := range neighbors {

		totalWeight += n.Score

		if n.Fraud {
			fraudWeight += n.Score
		}
	}

	var fraudScore float32

	if totalWeight > 0 {
		fraudScore = fraudWeight / totalWeight
	}

	resp := FraudResponse{
		Approved:   fraudScore < 0.6,
		FraudScore: fraudScore,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(resp)
}
