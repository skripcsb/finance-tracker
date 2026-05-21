# Finance Tracker

Трекер личных финансов.

## Стек

- Go 1.22+
- PostgreSQL
- Chi
- pgx
- golang-jwt (авторизация)
- goose (миграции)
- slog (логирование)
- yaml.v3 (конфиг)
- Docker Compose

## Функциональность

- Регистрация и JWT-авторизация
- CRUD по транзакциям (доходы/расходы)
- Категории трат (создание своих, дефолтные)
- Фильтрация транзакций по дате, категории, типу
- Отчёт за период — сумма по категориям (GET /api/reports?from=...&to=...)
- Экспорт транзакций в CSV (GET /api/export/csv)
- Пагинация списков

## Структура

```
finance-tracker/
├── cmd/tracker/         — точка входа
├── internal/
│   ├── config/          — загрузка конфига
│   ├── handler/         — HTTP-хендлеры
│   ├── service/         — бизнес-логика
│   ├── storage/         — работа с PostgreSQL
│   ├── auth/            — JWT, middleware авторизации
│   └── model/           — структуры данных
├── migrations/          — SQL-миграции (goose)
├── docker-compose.yml
├── Dockerfile
└── .env.example
```

## Запуск

```bash
docker-compose up --build
```
