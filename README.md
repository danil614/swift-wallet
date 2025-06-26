# Swift Wallet API

Высокопроизводительный кошелёк-сервис с поддержкой конкурентных операций до 1000 RPS.

## Описание

REST API для управления балансами кошельков с гарантией обработки всех запросов без 50x ошибок даже при высокой
нагрузке.

## Endpoints

**Изменение баланса:**

```http
POST /api/v1/wallet
Content-Type: application/json

{
  "walletId": "550e8400-e29b-41d4-a716-446655440000",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

**Получение баланса:**

```http
GET /api/v1/wallets/{WALLET_UUID}
```

## Технологии

* Backend: Go 1.24
* Database: PostgreSQL 16
* Containerization: Docker + Docker Compose
* Testing: Testify + pgxmock

## Быстрый старт

1. Клонируйте репозиторий:

```bash
git clone https://github.com/danil614/swift-wallet.git
cd swiftwallet
```

2. Настройте окружение (config.env):

```
APP_PORT=8080
DB_HOST=db
DB_PORT=5432
DB_USER=wallet
DB_PASS=wallet
DB_NAME=wallet
```

3. Запустите систему:

```bash
docker-compose up --build
```

## Тестирование

```bash
go test ./...
```