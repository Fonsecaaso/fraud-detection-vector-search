package v2

import "math"

func cosine(a, b Vector) float32 {

	var dot float32
	var normA float32
	var normB float32

	for i := range a {

		dot += a[i] * b[i]

		normA += a[i] * a[i]

		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}
