package models

type PotentialPricePeriod struct {
	prices
	PricePeriod PricePeriod
}

func (potential *PotentialPricePeriod) IsValidPrice(price int) bool {
	// if the price is zero, it means the price is unknown, so we pass it.
	if price == 0 {
		return true
	}

	return price >= potential.prices.MinPrice() &&
		price <= potential.prices.MaxPrice()
}
