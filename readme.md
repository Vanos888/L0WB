# L0 WB

Сервис для поиска  ордера по ID 

## Оглавление

*   [Как Запустить](#Как запустить)
*   [Документация ручек](#Документация ручек)



## Как запустить 
1. Запусти через Docker Compose (docker-compose up -d)
2. Запусти миграции (goose -dir db/migrations postgres "postgresql://Ivan:1q2w3e4r@localhost:5432/orders?sslmode=disable" up)
## Использование
Перейди на http://localhost:8081
Введи ID заказа (id можно взять в таблице orders)
Нажми "Поиск"
## Комментарии
Топик Kafka доступен по url http://localhost:8080/
Генерация ордеров в кафку происходит автоматически при помощи метода генерации