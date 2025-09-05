package config

type Config struct {
	PgDSN        string `envconfig:"PG_DSN"`
	KafkaBrokers string `envconfig:"KAFKA_BROKERS"`
	KafkaTopic   string `envconfig:"KAFKA_TOPIC"`
	ServerPort   string `envconfig:"SERVER_PORT"`
}
