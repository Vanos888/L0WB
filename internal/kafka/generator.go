package kafka

import (
	"L0WB/internal/domain"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

func GenerateFakeOrder() *domain.CompleteFakeOrder {
	// Генерируем базовые данные
	order := &domain.CompleteFakeOrder{
		OrderUID:          uuid.New().String(),
		TrackNumber:       "WBIL" + generateRandomString(10),
		Entry:             "WBIL",
		Locale:            randomChoice([]string{"en", "ru", "kz"}),
		InternalSignature: "",
		CustomerID:        "customer_" + generateRandomString(6),
		DeliveryService:   randomChoice([]string{"postal", "courier", "pickup"}),
		ShardKey:          randomChoice([]string{"1", "2", "3", "4", "5"}),
		SmID:              rand.Intn(100),
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          randomChoice([]string{"1", "2", "3"}),
		Delivery: domain.FakeDelivery{
			Name:    "Test User " + generateRandomString(5),
			Phone:   "+7" + generateRandomNumbers(9),
			Zip:     generateRandomNumbers(6),
			City:    randomChoice([]string{"Moscow", "SPb", "Kazan", "Novosibirsk"}),
			Address: "Street " + generateRandomString(8) + " " + generateRandomNumbers(2),
			Region:  randomChoice([]string{"Moscow", "Leningrad", "Tatarstan", "Siberia"}),
			Email:   "test" + generateRandomString(5) + "@mail.com",
		},
		Payment: domain.FakePayment{
			Transaction:  "tran_" + generateRandomString(10),
			RequestID:    "req_" + generateRandomString(8),
			Currency:     randomChoice([]string{"USD", "RUB", "EUR"}),
			Provider:     randomChoice([]string{"wbpay", "paypal", "stripe"}),
			Amount:       rand.Intn(1000) + 100,
			PaymentDt:    int(time.Now().Unix()),
			Bank:         randomChoice([]string{"alpha", "sber", "tinkoff"}),
			DeliveryCost: rand.Intn(100) + 50,
			GoodsTotal:   rand.Intn(10) + 1,
			CustomFee:    rand.Intn(20),
		},
	}

	// Добавляем 1-3 items
	itemCount := rand.Intn(3) + 1
	for i := 0; i < itemCount; i++ {
		order.Items = append(order.Items, domain.FakeItem{
			ChrtID:      9000000 + rand.Intn(10000),
			TrackNumber: order.TrackNumber,
			Price:       rand.Intn(500) + 100,
			Rid:         "rid_" + generateRandomString(8),
			Name:        randomChoice([]string{"T-Shirt", "Jeans", "Shoes", "Jacket", "Hat"}),
			Sale:        rand.Intn(30),
			Size:        randomChoice([]string{"S", "M", "L", "XL"}),
			TotalPrice:  rand.Intn(400) + 50,
			NmID:        2000000 + rand.Intn(10000),
			Brand:       randomChoice([]string{"Nike", "Adidas", "Puma", "Reebok"}),
			Status:      202,
		})
	}

	return order
}

func randomChoice(choices []string) string {
	return choices[rand.Intn(len(choices))]
}

func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func generateRandomNumbers(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = digits[rand.Intn(len(digits))]
	}
	return string(result)
}

func GenerateFakeOrders(count int) []*domain.CompleteFakeOrder {
	var orders []*domain.CompleteFakeOrder
	for i := 0; i < count; i++ {
		order := GenerateFakeOrder()
		if order != nil {
			orders = append(orders, order)
		}
	}
	return orders
}
