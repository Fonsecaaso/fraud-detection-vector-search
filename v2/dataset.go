// v2/dataset.go

package v2

import (
	"encoding/json"
	"os"
)

type Reference struct {
	Vector []float32 `json:"vector"`
	Label  string    `json:"label"`
}

func LoadDataset(path string) ([]Record, error) {

	raw, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var refs []Reference

	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, err
	}

	data := make([]Record, 0, len(refs))

	for _, r := range refs {

		data = append(data, Record{
			Vec:   Vector(r.Vector),
			Fraud: r.Label == "fraud",
		})
	}

	return data, nil
}
