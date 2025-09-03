package storage

import (
	"L0WB/internal/domain"
	"github.com/google/uuid"
	"sync"
	"time"
)

type OrderCache struct {
	mu     sync.RWMutex
	ttl    time.Duration
	orders map[uuid.UUID]*domain.Order
}

func NewOrderCache(ttl time.Duration) *OrderCache {
	return &OrderCache{
		ttl:    ttl,
		orders: make(map[uuid.UUID]*domain.Order),
	}
}

func (c *OrderCache) Set(orderUID uuid.UUID, order *domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[orderUID] = order
}

func (c *OrderCache) Get(orderUID uuid.UUID) (*domain.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.orders[orderUID]
	return order, ok
}

func (c *OrderCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.orders)
}
