package order

import (
	"L0WB/internal/domain"
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetOrder(ctx context.Context, orderUID uuid.UUID) (domain.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	order, err := getOrderByUIDWithTx(ctx, tx, orderUID.String())
	if err != nil {
		return domain.Order{}, fmt.Errorf("error getting order: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("error committing transaction: %v", err)
	}
	return order, nil
}

func (r *Repository) GetAllOrdersByUID(ctx context.Context) ([]uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	query := `SELECT order_uid FROM orders`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying: %v", err)
	}
	defer rows.Close()

	var orders []uuid.UUID
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("error scanning: %v", err)
		}

		uid, err := uuid.Parse(orderUID)
		if err != nil {
			continue //Пропускаю невалидные ID
		}

		orders = append(orders, uid)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}
	return orders, rows.Err()
}

func getOrderByUIDWithTx(ctx context.Context, tx pgx.Tx, orderUID string) (domain.Order, error) {
	q, args, err := squirrel.Select(
		"order_uid",
		"payment_id",
		"delivery_id",
		"item_ids",
		"track_number",
		"entry",
		"locate",
		"internal_signature",
		"customer_id",
		"delivery_service",
		"shardkey",
		"sm_id",
		"date_created",
		"oof_shard",
	).From("orders").
		Where(squirrel.Eq{"order_uid": orderUID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.Order{}, fmt.Errorf("error building query: %v", err)
	}

	var order Order

	err = tx.QueryRow(ctx, q, args...).Scan(
		&order.OrderUID,
		&order.PaymentID,
		&order.DeliveryID,
		&order.ItemIDs,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		return domain.Order{}, fmt.Errorf("error fetching order: %v", err)
	}

	q, args, err = squirrel.Select(
		"id",
		"name",
		"phone",
		"zip",
		"city",
		"address",
		"region",
		"email",
	).From("delivery").
		Where(squirrel.Eq{"id": order.DeliveryID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.Order{}, fmt.Errorf("error building query: %v", err)
	}

	var delivery Delivery

	err = tx.QueryRow(ctx, q, args...).Scan(
		&delivery.ID,
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)
	if err != nil {
		return domain.Order{}, fmt.Errorf("error fetching delivery: %v", err)
	}

	q, args, err = squirrel.Select(
		"id",
		"transaction",
		"request_id",
		"currency",
		"provider",
		"amount",
		"payment_dt",
		"bank",
		"delivery_cost",
		"goods_total",
		"custom_fee",
	).From("payments").
		Where(squirrel.Eq{"id": order.PaymentID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.Order{}, fmt.Errorf("error building query: %v", err)
	}

	var payments Payment

	err = tx.QueryRow(ctx, q, args...).Scan(
		&payments.ID,
		&payments.Transaction,
		&payments.RequestID,
		&payments.Currency,
		&payments.Provider,
		&payments.Amount,
		&payments.PaymentDt,
		&payments.Bank,
		&payments.DeliveryCost,
		&payments.GoodsTotal,
		&payments.CustomFee,
	)
	if err != nil {
		return domain.Order{}, fmt.Errorf("error fetching payments: %v", err)
	}

	q, args, err = squirrel.Select(
		"id",
		"chart_id",
		"track_number",
		"price",
		"rid",
		"name",
		"sale",
		"size",
		"total_price",
		"nm_id",
		"brand",
		"status",
	).From("items").
		Where(squirrel.Eq{"id": order.ItemIDs}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return domain.Order{}, fmt.Errorf("error building query: %v", err)
	}

	var items []Item

	rows, err := tx.Query(ctx, q, args...)
	if err != nil {
		return domain.Order{}, fmt.Errorf("error fetching items: %v", err)
	}
	for rows.Next() {
		var item Item
		if err = rows.Scan(
			&item.ID,
			&item.ChartID,
			&item.TrackNumber,
			&item.Price,
			&item.RID,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		); err != nil {
			return domain.Order{}, fmt.Errorf("error fetching item: %v", err)
		}
		items = append(items, item)
	}

	return toDomainOrder(order, delivery, payments, items), nil
}

func toDomainOrder(dbOrder Order, delivery Delivery, payment Payment, items []Item) domain.Order {
	return domain.Order{
		ID:                dbOrder.OrderUID,
		TrackNumber:       dbOrder.TrackNumber,
		Entry:             dbOrder.Entry,
		Locale:            dbOrder.Locale,
		InternalSignature: dbOrder.InternalSignature,
		CustumerID:        dbOrder.CustomerID,
		DeliveryService:   dbOrder.DeliveryService,
		ShardKey:          dbOrder.ShardKey,
		SmID:              dbOrder.SmID,
		DateCreated:       dbOrder.DateCreated,
		OofShard:          dbOrder.OofShard,
		Delivery: domain.Delivery{
			Name:    delivery.Name,
			Phone:   delivery.Phone,
			Zip:     delivery.Zip,
			City:    delivery.City,
			Address: delivery.Address,
			Region:  delivery.Region,
			Email:   delivery.Email,
		},
		Payment: domain.Payment{
			Transaction:  payment.Transaction,
			RequestID:    payment.RequestID,
			Currency:     payment.Currency,
			Provider:     payment.Provider,
			Amount:       payment.Amount,
			PaymentDt:    payment.PaymentDt,
			Bank:         payment.Bank,
			DeliveryCost: payment.DeliveryCost,
			GoodsTotal:   payment.GoodsTotal,
			CustomFee:    payment.CustomFee,
		},
		Items: toDomainItems(items),
	}
}

func toDomainItems(dbItems []Item) []domain.Item {
	domainItems := make([]domain.Item, len(dbItems))
	for i, dbItem := range dbItems {
		domainItems[i] = domain.Item{
			ChartID:     dbItem.ChartID,
			TrackNumber: dbItem.TrackNumber,
			Price:       dbItem.Price,
			RID:         dbItem.RID,
			Name:        dbItem.Name,
			Sale:        dbItem.Sale,
			Size:        dbItem.Size,
			TotalPrice:  dbItem.TotalPrice,
			NmID:        dbItem.NmID,
			Brand:       dbItem.Brand,
			Status:      dbItem.Status,
		}
	}
	return domainItems
}

func toDTOItems(domainItems []domain.Item) []Item {
	dtoItems := make([]Item, len(domainItems))
	for i, domainItem := range domainItems {
		dtoItems[i] = Item{
			ID:          uuid.New(),
			ChartID:     domainItem.ChartID,
			TrackNumber: domainItem.TrackNumber,
			Price:       domainItem.Price,
			RID:         domainItem.RID,
			Name:        domainItem.Name,
			Sale:        domainItem.Sale,
			Size:        domainItem.Size,
			TotalPrice:  domainItem.TotalPrice,
			NmID:        domainItem.NmID,
			Brand:       domainItem.Brand,
			Status:      domainItem.Status,
		}
	}
	return dtoItems
}

func (r *Repository) SaveOrder(ctx context.Context, order *domain.Order) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	paymentID := uuid.New()
	deliveryID := uuid.New()

	qdelivery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert("delivery").
		Columns("id", "name", "phone", "zip", "city", "address", "region", "email").
		Values(
			deliveryID,
			order.Delivery.Name,
			order.Delivery.Phone,
			order.Delivery.Zip,
			order.Delivery.City,
			order.Delivery.Address,
			order.Delivery.Region,
			order.Delivery.Email,
		)
	query, args, err := qdelivery.ToSql()
	if err != nil {
		return fmt.Errorf("error building query delivery: %v", err)
	}
	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error saving delivery: %v", err)
	}

	qpayment := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert("payments").
		Columns("id", "transaction", "request_id", "currency", "provider", "amount", "payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
		Values(
			paymentID,
			order.Payment.Transaction,
			order.Payment.RequestID,
			order.Payment.Currency,
			order.Payment.Provider,
			order.Payment.Amount,
			order.Payment.PaymentDt,
			order.Payment.Bank,
			order.Payment.DeliveryCost,
			order.Payment.GoodsTotal,
			order.Payment.CustomFee,
		)
	query, args, err = qpayment.ToSql()
	if err != nil {
		return fmt.Errorf("error building query payment: %v", err)
	}
	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error saving payment: %v", err)
	}

	items := toDTOItems(order.Items)
	var itemIDs []uuid.UUID
	for _, i := range items {
		qitem := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
			Insert("items").
			Columns("id", "chart_id", "track_number", "price", "rid", "name", "sale", "size", "total_price", "nm_id", "brand", "status").
			Values(
				i.ID,
				i.ChartID,
				i.TrackNumber,
				i.Price,
				i.RID,
				i.Name,
				i.Sale,
				i.Size,
				i.TotalPrice,
				i.NmID,
				i.Brand,
				i.Status,
			)
		itemIDs = append(itemIDs, i.ID)
		query, args, err = qitem.ToSql()
		if err != nil {
			return fmt.Errorf("error building query item: %v", err)
		}
		_, err = r.db.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("error saving item: %v", err)
		}
	}

	qorder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert("orders").
		Columns("order_uid", "payment_id", "delivery_id", "item_ids", "track_number", "entry", "locate", "internal_signature", "customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard").
		Values(
			order.ID,
			paymentID,
			deliveryID,
			itemIDs,
			order.TrackNumber,
			order.Entry,
			order.Locale,
			order.InternalSignature,
			order.CustumerID,
			order.DeliveryService,
			order.ShardKey,
			order.SmID,
			order.DateCreated,
			order.OofShard,
		)

	query, args, err = qorder.ToSql()
	if err != nil {
		return fmt.Errorf("error building query orders: %v", err)
	}
	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error saving orders: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}
