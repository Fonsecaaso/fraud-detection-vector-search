package v2

import "time"

type Vector []float32

var normalization = struct {
	MaxAmount            float64
	MaxInstallments      float64
	AmountVsAvgRatio     float64
	MaxMinutes           float64
	MaxKm                float64
	MaxTxCount24h        float64
	MaxMerchantAvgAmount float64
}{
	MaxAmount:            10000,
	MaxInstallments:      12,
	AmountVsAvgRatio:     10,
	MaxMinutes:           1440,
	MaxKm:                1000,
	MaxTxCount24h:        20,
	MaxMerchantAvgAmount: 10000,
}
var mccRisk = map[string]float32{
	"5411": 0.15,
	"5812": 0.30,
	"5912": 0.20,
	"5944": 0.45,
	"7801": 0.80,
	"7802": 0.75,
	"7995": 0.85,
	"4511": 0.35,
	"5311": 0.25,
	"5999": 0.50,
}

func BuildVector(req FraudRequest) Vector {

	vec := make(Vector, 14)

	//
	// 0
	// amount
	//

	vec[0] = clamp01(
		req.Transaction.Amount /
			normalization.MaxAmount,
	)

	//
	// 1
	// installments
	//

	vec[1] = clamp01(
		float64(req.Transaction.Installments) /
			normalization.MaxInstallments,
	)

	//
	// 2
	// amount vs customer avg
	//

	if req.Customer.AvgAmount > 0 {

		ratio :=
			req.Transaction.Amount /
				req.Customer.AvgAmount

		vec[2] = clamp01(
			ratio /
				normalization.AmountVsAvgRatio,
		)
	}

	//
	// 3
	// merchant avg amount
	//

	vec[3] = clamp01(
		req.Merchant.AvgAmount /
			normalization.MaxMerchantAvgAmount,
	)

	//
	// 4
	// tx count 24h
	//

	vec[4] = clamp01(
		float64(req.Customer.TxCount24h) /
			normalization.MaxTxCount24h,
	)

	//
	// 5
	// minutes since last tx
	//

	//
	// 6
	// km from last tx
	//

	if req.LastTransaction == nil {

		vec[5] = -1
		vec[6] = -1

	} else {

		now, _ := time.Parse(
			time.RFC3339,
			req.Transaction.RequestedAt,
		)

		prev, _ := time.Parse(
			time.RFC3339,
			req.LastTransaction.Timestamp,
		)

		minutes :=
			now.Sub(prev).Minutes()

		vec[5] = clamp01(
			minutes /
				normalization.MaxMinutes,
		)

		vec[6] = clamp01(
			req.LastTransaction.KmFromCurrent /
				normalization.MaxKm,
		)
	}

	//
	// 7
	// km from home
	//

	vec[7] = clamp01(
		req.Terminal.KmFromHome /
			normalization.MaxKm,
	)

	//
	// 8
	// merchant avg
	//

	vec[8] = clamp01(
		req.Merchant.AvgAmount /
			normalization.MaxMerchantAvgAmount,
	)

	//
	// 9
	// is online
	//

	if req.Terminal.IsOnline {
		vec[9] = 1
	}

	//
	// 10
	// card present
	//

	if req.Terminal.CardPresent {
		vec[10] = 1
	}

	//
	// 11
	// known merchant
	//

	known := false

	for _, m := range req.Customer.KnownMerchants {
		if m == req.Merchant.ID {
			known = true
			break
		}
	}

	if known {
		vec[11] = 1
	}

	//
	// 12
	// MCC risk
	//

	risk, ok := mccRisk[req.Merchant.MCC]

	if !ok {
		risk = 0.5
	}

	vec[12] = risk

	//
	// 13
	// hour of day
	//

	t, _ := time.Parse(
		time.RFC3339,
		req.Transaction.RequestedAt,
	)

	vec[13] = float32(
		float64(t.Hour()) / 24.0,
	)

	return vec
}
func clamp01(v float64) float32 {

	if v < 0 {
		return 0
	}

	if v > 1 {
		return 1
	}

	return float32(v)
}
