package pricing

func EnrichTick(prevPrice *float64, tick PriceTick) EnrichedTick {
	isAnomaly := false

	if prevPrice != nil {
		change := (tick.Price - *prevPrice) / *prevPrice
		if change > 0.05 || change < -0.05 {
			isAnomaly = true
		}
	}

	return EnrichedTick{
		Symbol:    tick.Symbol,
		Price:     tick.Price,
		Volume:    tick.Volume,
		Ts:        tick.Ts,
		IsAnomaly: isAnomaly,
	}
}
