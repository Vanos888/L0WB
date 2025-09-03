package domain

import (
	"github.com/google/uuid"
	"time"
)

type Order struct {
	ID                uuid.UUID
	TrackNumber       string
	Entry             string
	Delivery          Delivery
	Payment           Payment
	Items             []Item
	Locale            string
	InternalSignature string
	CustumerID        string
	DeliveryService   string
	ShardKey          string
	SmID              int
	DateCreated       time.Time
	OofShard          string
}

type Delivery struct {
	Name    string
	Phone   string
	Zip     string
	City    string
	Address string
	Region  string
	Email   string
}

type Payment struct {
	Transaction  string
	RequestID    string
	Currency     string
	Provider     string
	Amount       int
	PaymentDt    int
	Bank         string
	DeliveryCost int
	GoodsTotal   int
	CustomFee    int
}

type Item struct {
	ChartID     int
	TrackNumber string
	Price       int
	RID         string
	Name        string
	Sale        int
	Size        string
	TotalPrice  int
	NmID        int
	Brand       string
	Status      int
}
