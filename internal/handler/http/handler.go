package http

import (
	"L0WB/internal/domain"
	og "L0WB/internal/generated/servers/http/ordergen"
	"context"
	"github.com/google/uuid"
)

type IService interface {
	GetOrder(ctx context.Context, orderUID uuid.UUID) (*domain.Order, error)
}

type Handler struct {
	Service IService
	og.UnimplementedHandler
}

func NewHandler(service IService) *Handler {
	return &Handler{
		Service: service,
	}
}
