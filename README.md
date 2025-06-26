# goRedis

**goRedis** — это демон на Go 1.24, предоставляющий кэш-/key-value-хранилище «товаров» (*Item*) с тремя уровнями устойчивости данных:

1. **In-Memory** — быстрая оперативная работа.
2. **Redis** — периодический дамп (TTL) для мгновенного восстановления после перезапуска.
3. **PostgreSQL / JSON-файл** — финальный долговременный слой при остановке сервиса.

Сервис открывает:

* **REST API** (`/items`, `/items/{id}`)
* **gRPC API** (`GoRedisService`)
* **CLI** (интерактивная оболочка в STDIN)

Готовый Docker-композ запустит всё сразу: приложение, Redis 7 и PostgreSQL 17.

---

## Features

| ✔︎                                                           |
|--------------------------------------------------------------|
| In-memory стор с потокобезопасной картой и RW-мьютексами     |
| Периодическое TTL-резервирование в Redis                     |
| При завершении — дамп в Redis + PostgreSQL + JSON-файл       |
| Горячий reload состояния при старте из любого из источников  |
| REST v1 (`/items`) + gRPC (protobuf v3)                      |
| Dockerfile (multi-stage build) и `docker-compose.yml`        |
| Интерактивный CLI: *set / get / del / update*                |

---

## Architecture

```text
                    +-----------+         +-----------+
   REST 8080 <----> |  REST API |         |  gRPC API | <----> gRPC 50051
                    +-----+-----+         +-----+-----+
                          \                     /
                           \                   /
                            v                 v
                        +-------------------------+
                        |     In-Memory store     |
                        |  (map[int64]Item, RW)   |
                        +-----------+-------------+
                                     |
     Periodic dump (TTL)             | Snapshot on shutdown
                                     v
            +------------+      +-----------+      +----------------+
            |   Redis    |      | PostgreSQL |     |   JSON-file    |
            |  (KEY)     |      |  table     |     |  backup.json   |
            +------------+      +-----------+      +----------------+
```

* **Manager** (`internal/persist/manager.go`) занимается восстановлением и резервным копированием.
* **Storager** (`internal/storage/storager.go`) описывает единственный CRUD-интерфейс, имплементация *memory.go*.
* **Logger**: вся логика прозрачна и покрыта логированием (`internal/logger`).

---

## Quick Start

### Docker Compose

```bash
# клонируем репо, если нужно
git clone https://github.com/folivorra/goRedis.git
cd goRedis

# старт всех сервисов
docker compose up --build
```

После сборки будут доступны:

| Сервис   | URL                                        |
| -------- | ------------------------------------------ |
| REST API | `http://localhost:8080`                    |
| gRPC API | `localhost:50051`                          |
| Redis    | `localhost:6379`                           |
| Postgres | `localhost:5432` (`myuser` / `mypassword`) |

> Конфигурация читается из `config/app_config.yaml` (см. ниже) — она смаунтирована read-only в контейнер `app`.

---

## Configuration

`config/app_config.yaml` (пример по умолчанию):

```yaml
server:
  http_port: ":8080"    # REST
  grpc_port: ":50051"   # gRPC

storage:
  ttl: "2m"             # как часто дампить в Redis
  redis_key: "myapp:items"
  dump_file: "/app/data/backup.json"
  postgres_dsn: "postgresql://myuser:mypassword@postgres:5432/"

logger:
  log_file: "/app/logs/app.log"
```

> Отдельная переменная окружения `APP_CONFIG` может указать иной путь к YAML-файлу.

---

## REST API

### Сущность `Item`

| Поле  | Тип    | Описание                       |
| ----- | ------ | ------------------------------ |
| id    | int64  | Уникальный идентификатор (> 0) |
| name  | string | Название                       |
| price | float  | Цена (≥ 0)                     |

### Эндпоинты

| Метод    | Путь          | Описание       |
| -------- | ------------- | -------------- |
| `POST`   | `/items`      | Создать item   |
| `GET`    | `/items`      | Получить все   |
| `GET`    | `/items/{id}` | Получить по id |
| `PUT`    | `/items/{id}` | Обновить       |
| `DELETE` | `/items/{id}` | Удалить        |

---

## gRPC API

Файл протокола: `proto/goredis/v1/goredis.proto`.

```
service GoRedisService {
  rpc GetItem    (GetItemRequest)    returns (GetItemResponse);
  rpc CreateItem (CreateItemRequest) returns (CreateItemResponse);
  rpc UpdateItem (UpdateItemRequest) returns (UpdateItemResponse);
  rpc DeleteItem (DeleteItemRequest) returns (DeleteItemResponse);
  rpc GetAllItems(GetAllItemsRequest) returns (GetAllItemsResponse);
}
```

---

## CLI

После старта приложения просто начните вводить команды в тот же терминал:

```text
set <id> <name> <price>     # добавить
get <id>                    # получить
get all                     # вывести всё
update <id> <name> <price>  # обновить
del <id>                    # удалить
```

Пример:

```bash
set 3 "Orange" 7.30
get 3
update 3 "Blood Orange" 8.10
del 3
```

---

## Persistence layers

| Слой           | Когда пишем          | Формат / Механизм                        |
| -------------- | -------------------- | ---------------------------------------- |
| **Redis**      | Каждые `storage.ttl` | JSON-строка по ключу `storage.redis_key` |
| **PostgreSQL** | `App.Stop()`         | Таблица `items(id,name,price)`           |
| **JSON-file**  | `App.Stop()`         | Читаемый бэкап `dump_file`               |

Порядок восстановления при запуске:

1. Redis (если ключ найден).
2. PostgreSQL (если есть записи).
3. JSON-файл.
4. Чистый старт (пустой in-memory store).

---

## Project Structure

```
goRedis/
│
├── cmd/                # точка входа (main.go)
├── config/             # пример конфигурации
├── internal/
│   ├── app/            # инициализация всех зависимостей
│   ├── cli/            # интерактивная CLI
│   ├── config/         # загрузка YAML
│   ├── logger/         # настроенный log.Logger
│   ├── model/          # структуры доменной области
│   ├── storage/        # memory / redis / postgres
│   ├── persist/        # дампы, загрузка, менеджер
│   └── transport/      # rest + grpc + общие интерфейсы
├── proto/              # protobuf схемы (v1)
├── Dockerfile          # multi-stage build
└── docker-compose.yml  # полный dev-стек
```