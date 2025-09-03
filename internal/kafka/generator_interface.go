package kafka

import "L0WB/internal/domain"

type OrderGeneratorImpl struct{}

func (g *OrderGeneratorImpl) GenerateFakeOrders(count int) []*domain.CompleteFakeOrder {
	return GenerateFakeOrders(count)
}
