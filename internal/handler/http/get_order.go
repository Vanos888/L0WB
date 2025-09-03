package http

import (
	og "L0WB/internal/generated/servers/http/ordergen"
	"context"
	"fmt"
	"github.com/google/uuid"
)

func (h *Handler) GetOrder(ctx context.Context, req *og.GetOrderRequest) (*og.GetOrderResponse, error) {

	id, err := uuid.Parse(req.OrderUID)
	if err != nil {
		return nil, fmt.Errorf("invalid order uid")
	}

	order, err := h.Service.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := getOrderResponseFromDomain(order)
	return resp, nil
}
