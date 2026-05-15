// v2/types.go

package v2

type Record struct {
	Vec   Vector
	Fraud bool
}

type Neighbor struct {
	Score float32 `json:"score"`
	Fraud bool    `json:"fraud"`
}
