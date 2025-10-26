# user-aggregation

Сервис для агрегации пользовательских подписок/услуг с REST API: хранит записи о стоимости, сроках и названии сервиса, позволяет получать/изменять данные и считать итоговую сумму по фильтрам.

## Возможности

* Создание записи о подписке пользователя
* Получение всех записей и выборка по `user_id`
* Частичное обновление цены/даты окончания
* Удаление всех записей пользователя
* Подсчёт суммарной стоимости с фильтрами (`user_id`, `service_name`, `start_date`, `end_date`)
* Встроенная Swagger UI документация 

## Технологии

* **Go** 
* **PostgreSQL 16**
* **gorilla/mux** — роутер
* **pgx/v5** — драйвер PostgreSQL
* **golang-migrate** — миграции
* **zap / slog** — логирование в зависимости от окружения(local/prod) сервера
* **swaggo/http-swagger** — Swagger UI

## Структура

```
cmd/
  user-aggregation/     # запуск API-сервера
  migrator/             # утилита миграций (up|down|version)
internal/
  config/               # чтение и валидация конфигурации
  repo/                 # интерфейс и реализация хранилища (Postgres)
  server/               # http-сервер и хендлеры
  transport/http/respond# унифицированные ответы/ошибки
  models/               # доменные и ответные модели
migrations/             # SQL-миграции
config/config.yaml      # дефолтная конфигурация
docs/swagger.(json|yaml)# схемы Swagger
.env                    # пример переменных окружения
Docker-compose.yaml     # локальная инфраструктура
```

## Docker Compose

Контейнеры:

* `db` — PostgreSQL (порт по умолчанию `5432`, пробрасывается из `.env`)
* `migrator` — прогоняет миграции `migrations/0001_init.up.sql`
* `app` — API-сервер (порт `APP_PORT`, по умолчанию `8080`)



## Конфигурация

**config/config.yaml** (значения по умолчанию):

```yaml
app:
  name: user-aggregation
  env: local

http_server:
  address: ":8080"
  timeout: "4s"
  idle_timeout: "60s"
  shutdown_timeout: "10s"

storage:
  db_url: "postgres://postgres:postgres@db:5432/user-aggregation?sslmode=disable"
```

**.env** (используется docker-compose и для удобства локально):

```env
# ---------- Postgres ----------
POSTGRES_DB=user-aggregation
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_PORT=5432

# ---------- App ----------
APP_PORT=8080
CONFIG_PATH=./config/config.yaml

# ---------- Migrator ----------
MIGRATIONS_DB_URL=postgres://postgres:postgres@db:5432/user-aggregation?sslmode=disable
MIGRATIONS_PATH=./migrations
MIGRATIONS_COMMAND=up   # up | down | version
```

## API

База: `http://localhost:8080`

### Swagger

* UI: `GET /docs` (редирект на `/swagger/index.html`)
* Спецификация: `GET /swagger/doc.json`

### Модели

```json
// models.UserInfo
{
  "service_name": "string",
  "price": 123,                // integer
  "user_id": "uuid",          // генерируется / хранится на стороне сервиса
  "start_date": "2025-01-01T00:00:00Z",
  "end_date":   "2025-12-31T23:59:59Z"
}

// models.UpdateUserInfo (PATCH)
{
  "price": 123,                // optional
  "end_date": "2025-06-30T00:00:00Z" // optional
}

// response.ErrorPayload
{ "error": "string", "op": "string", "status": 400 }

// response.Summary
{ "total_cost": 456 }
```

### Эндпойнты

* `GET /users` — список всех записей (`[]UserInfo`)
* `POST /users` — создать запись (`201 Created`, body: `UserInfo`)
* `GET /users/{id}` — записи по `user_id` (`[]UserInfo`)
* `PATCH /users/{id}` — частичное обновление цены/даты окончания (body: `UpdateUserInfo`)
* `DELETE /users/{id}` — удалить все записи по `user_id` (возвращает количество удалённых записей)
* `GET /summary?user_id=&service_name=&start_date=&end_date=` — сумма `price` по фильтрам (`Summary`)

> Формат дат: ISO 8601 (RFC3339).

