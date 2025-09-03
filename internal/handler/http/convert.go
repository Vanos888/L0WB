package http

import (
	"L0WB/internal/domain"
	og "L0WB/internal/generated/servers/http/ordergen"
)

func getOrderResponseFromDomain(order *domain.Order) *og.GetOrderResponse {
	res := &og.GetOrderResponse{
		Success: true,
		Data: og.Order{
			OrderUID:    order.ID.String(),
			TrackNumber: order.TrackNumber,
			Entry:       order.Entry,
			Delivery: og.Delivery{
				Name:    order.Delivery.Name,
				Phone:   order.Delivery.Phone,
				Zip:     order.Delivery.Zip,
				City:    order.Delivery.City,
				Address: order.Delivery.Address,
				Region:  order.Delivery.Region,
				Email:   order.Delivery.Email,
			},
			Payment: og.Payment{
				Transaction:  order.Payment.Transaction,
				RequestID:    order.Payment.RequestID,
				Currency:     order.Payment.Currency,
				Provider:     order.Payment.Provider,
				Amount:       order.Payment.Amount,
				PaymentDt:    order.Payment.PaymentDt,
				Bank:         order.Payment.Bank,
				DeliveryCost: order.Payment.DeliveryCost,
				GoodsTotal:   order.Payment.GoodsTotal,
				CustomFee:    order.Payment.CustomFee,
			},
			Items:             ConvertToOGItems(order.Items),
			Locale:            order.Locale,
			InternalSignature: order.InternalSignature,
			CustomerID:        order.CustumerID,
			DeliveryService:   order.DeliveryService,
			Shardkey:          order.ShardKey,
			SmID:              order.SmID,
			DateCreated:       order.DateCreated,
			OofShard:          order.OofShard,
		},
	}
	return res
}

func ConvertToOGItems(domainItems []domain.Item) []og.Item {
	if len(domainItems) == 0 {
		return []og.Item{}
	}

	ogItems := make([]og.Item, len(domainItems))

	for i, domainItem := range domainItems {
		ogItems[i] = og.Item{
			ChrtID:      domainItem.ChartID,
			TrackNumber: domainItem.TrackNumber,
			Price:       domainItem.Price,
			Rid:         domainItem.RID,
			Name:        domainItem.Name,
			Sale:        domainItem.Sale,
			Size:        domainItem.Size,
			TotalPrice:  domainItem.TotalPrice,
			NmID:        domainItem.NmID,
			Brand:       domainItem.Brand,
			Status:      domainItem.Status,
		}
	}
	return ogItems
}
