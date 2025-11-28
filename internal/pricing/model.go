package pricing

import "time"

type PriceTick struct {
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
	Volume float64   `json:"volume"`
	Ts     time.Time `json:"ts"`
}

type EnrichedTick struct {
	Symbol    string
	Price     float64
	Volume    float64
	Ts        time.Time
	IsAnomaly bool
}

// Enkel anomaly: prisändring > 2% jämfört med föregående pris
func NewEnrichedTick(prevPrice *float64, t PriceTick) EnrichedTick {
	isAnomaly := false
	if prevPrice != nil && *prevPrice > 0 {
		diff := (t.Price - *prevPrice) / *prevPrice
		if diff > 0.02 || diff < -0.02 { // > 2% upp eller ner
			isAnomaly = true
		}
	}

	return EnrichedTick{
		Symbol:    t.Symbol,
		Price:     t.Price,
		Volume:    t.Volume,
		Ts:        t.Ts,
		IsAnomaly: isAnomaly,
	}
}
