package models

// Since many price periods share the same price bracket, we are going to re-use most
// of the information for a potential period when it is identical to anther potential
// period of a different week on the same pattern.
type potentialPhaseSubPeriod struct {
}

type PotentialPricePeriod struct {
	prices
	Spike

	// The price period
	PricePeriod PricePeriod

	// The pattern phase used to generate this period.
	PatternPhase PatternPhase
}

// Returns ``true`` if ``price`` falls within the price range of this potential period.
// Used by the predictor to remove phase permutations that do not match the current
// price values of a user.
func (potential *PotentialPricePeriod) IsValidPrice(price int) bool {
	// if the price is zero, it means the price is unknown, so we pass it.
	if price == 0 {
		return true
	}

	return price >= potential.prices.MinPrice() &&
		price <= potential.prices.MaxPrice()
}
