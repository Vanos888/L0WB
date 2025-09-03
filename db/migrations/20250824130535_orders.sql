-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders(
    order_uid uuid PRIMARY KEY ,
    payment_id uuid NOT NULL,
    delivery_id uuid NOT NULL,
    item_ids uuid[] NOT NULL,
    track_number TEXT NOT NULL,
    entry TEXT NOT NULL,
    locate VARCHAR(5),
    internal_signature TEXT,
    customer_id TEXT NOT NULL,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INT NOT NULL,
    date_created TIMESTAMP default NOW(),
    oof_shard TEXT
    );

CREATE TABLE IF NOT EXISTS delivery (
    id uuid PRIMARY KEY,
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);

CREATE TABLE IF NOT EXISTS payments(
    id uuid PRIMARY KEY,
    transaction TEXT NOT NULL,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount INT,
    payment_dt BIGINT,
    bank TEXT,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

CREATE TABLE IF NOT EXISTS items(
    id uuid PRIMARY KEY,
    chart_id BIGINT,
    track_number TEXT,
    price INT,
    rid TEXT,
    name TEXT,
    sale INT,
    size TEXT,
    total_price INT,
    nm_id BIGINT,
    brand TEXT,
    status INT
);

CREATE INDEX idx_orders_payment_id ON orders USING btree (payment_id);
CREATE INDEX idx_orders_delivery_id ON orders USING btree (delivery_id);
CREATE INDEX idx_orders_item_ids ON orders USING gin (item_ids);



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS delivery;
DROP TABLE IF EXISTS orders;
DROP INDEX IF EXISTS idx_orders_payment_id;
DROP INDEX IF EXISTS idx_orders_delivery_id;
DROP INDEX IF EXISTS idx_orders_item_ids;
-- +goose StatementEnd
