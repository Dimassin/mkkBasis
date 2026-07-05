# Task Management Service

Микросервисная система управления задачами в командах: регистрация/аутентификация, команды, задачи, отчёты, кеширование и метрики.

## Архитектура

| Сервис | Порт | Описание |
|--------|------|----------|
| **auth** | `8080` | Регистрация и логин (JWT) |
| **teams** | `8081` | Команды и приглашения |
| **tasks** | `8082` | Задачи, история, отчёты |
| **MySQL** | `3306` | Общая БД `auth` |
| **Redis** | `6379` | Кеш списка задач (tasks) |

## Требования

- Docker + Docker Compose
- Go 1.25+ (для локальной разработки и тестов)

## Запуск

```bash
docker compose up --build
```

Сервисы поднимутся на портах `8080`, `8081`, `8082`.

### Миграции (чистая БД)

Порядок применения:

1. `services/auth/migrations/000001` — таблица `users`
2. `services/teams/migrations/000001` — `teams`, `team_members`
3. `services/tasks/migrations/000001` — `tasks`, `task_history`, `task_comments`

Пример через migrate внутри контейнера:

```bash
migrate -path /migrations -database "mysql://myuser:mypassword@tcp(mysql:3306)/auth" up
```

### Локальный запуск (без Docker)

Нужны MySQL и Redis. В каждом сервисе:

```bash
cd services/auth && go run ./cmd/api
cd services/teams && go run ./cmd/api
cd services/tasks && go run ./cmd/api
```

Переменные окружения по умолчанию совпадают с `docker-compose.yml` (`DB_*`, `JWT_SECRET`, `REDIS_*`).

## Аутентификация

Для защищённых эндпоинтов:

```
Authorization: Bearer <access_token>
Content-Type: application/json
```

Токен получается через `POST /api/v1/login` или `POST /api/v1/register` на **auth** (`:8080`).

Общий `JWT_SECRET` должен совпадать во всех сервисах.

## Лимиты и метрики

- Rate limit: **100 запросов/мин** на пользователя (по JWT) или IP (для auth)
- Prometheus: `GET /metrics` на каждом сервисе
- Graceful shutdown по `SIGINT` / `SIGTERM`

---

## API

### Auth — `http://localhost:8080`

#### POST `/api/v1/register`

Регистрация пользователя.

**Body:**

```json
{
  "email": "owner@example.com",
  "password": "secret12",
  "username": "owner"
}
```

**Response `201`:**

```json
{
  "access_token": "eyJhbG...",
  "user_id": "1",
  "email": "owner@example.com",
  "username": "owner"
}
```

---

#### POST `/api/v1/login`

Аутентификация.

**Body:**

```json
{
  "email": "owner@example.com",
  "password": "secret12"
}
```

**Response `200`:**

```json
{
  "access_token": "eyJhbG...",
  "user_id": "1",
  "email": "owner@example.com",
  "username": "owner"
}
```

---

### Teams — `http://localhost:8081`

Все эндпоинты требуют `Authorization: Bearer <token>`.

#### POST `/api/v1/teams`

Создать команду (создатель становится `owner`).

**Body:**

```json
{
  "name": "Backend Team",
  "description": "Команда разработки"
}
```

**Response `201`:**

```json
{
  "id": 1,
  "name": "Backend Team",
  "description": "Команда разработки",
  "created_by": 1,
  "created_at": "2026-07-05T12:00:00Z"
}
```

---

#### GET `/api/v1/teams`

Список команд, в которых состоит текущий пользователь.

**Body:** нет

**Response `200`:**

```json
[
  {
    "id": 1,
    "name": "Backend Team",
    "description": "Команда разработки",
    "created_by": 1,
    "created_at": "2026-07-05T12:00:00Z"
  }
]
```

---

#### POST `/api/v1/teams/{id}/invite`

Пригласить пользователя в команду (только `owner` / `admin`). Перед добавлением вызывается mock email-сервис через circuit breaker.

**Body:**

```json
{
  "email": "member@example.com"
}
```

**Response `201`:**

```json
{
  "team_id": 1,
  "user_id": 2,
  "email": "member@example.com",
  "role": "member"
}
```

**Ошибки:** `403` (нет прав), `404` (команда/пользователь), `409` (уже в команде), `503` (email/circuit breaker)

Для теста падения email: `MOCK_EMAIL_FAIL=true` в env teams-сервиса.

---

### Tasks — `http://localhost:8082`

Все эндпоинты требуют `Authorization: Bearer <token>`.

#### POST `/api/v1/tasks`

Создать задачу (только член команды).

**Body:**

```json
{
  "title": "Настроить CI",
  "description": "Добавить pipeline",
  "team_id": 1,
  "assignee_id": 2,
  "priority": "high",
  "status": "new",
  "due_date": "2026-07-15"
}
```

Минимальный body:

```json
{
  "title": "Настроить CI",
  "team_id": 1
}
```

`priority` по умолчанию: `medium`. `status` по умолчанию: `new`.  
Допустимые `status`: `new`, `in_progress`, `done`, `cancelled`.  
Допустимые `priority`: `low`, `medium`, `high`, `critical`.  
`due_date` — формат `YYYY-MM-DD`.

**Response `201`:**

```json
{
  "id": 1,
  "title": "Настроить CI",
  "description": "Добавить pipeline",
  "status": "new",
  "priority": "high",
  "assignee_id": 2,
  "team_id": 1,
  "created_by": 1,
  "due_date": "2026-07-15T00:00:00Z",
  "created_at": "2026-07-05T12:00:00Z",
  "updated_at": "2026-07-05T12:00:00Z"
}
```

---

#### GET `/api/v1/tasks`

Список задач с фильтрацией и пагинацией. Результат кешируется в Redis (TTL 5 мин).

**Query-параметры:**

| Параметр | Обязательный | Описание |
|----------|--------------|----------|
| `team_id` | да | ID команды |
| `status` | нет | `new`, `in_progress`, `done`, `cancelled` (`todo` = `new`) |
| `assignee_id` | нет | ID исполнителя |
| `page` | нет | Страница (по умолчанию `1`) |
| `limit` | нет | Размер страницы (по умолчанию `20`, макс `100`) |
| `cursor` | нет | Cursor-based пагинация (альтернатива `page`) |

**Пример:**

```
GET /api/v1/tasks?team_id=1&status=new&assignee_id=2&page=1&limit=20
```

**Body:** нет

**Response `200`:**

```json
{
  "items": [
    {
      "id": 1,
      "title": "Настроить CI",
      "description": "Добавить pipeline",
      "status": "new",
      "priority": "high",
      "assignee_id": 2,
      "team_id": 1,
      "created_by": 1,
      "due_date": "2026-07-15T00:00:00Z",
      "created_at": "2026-07-05T12:00:00Z",
      "updated_at": "2026-07-05T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

С cursor:

```
GET /api/v1/tasks?team_id=1&limit=20&cursor=15
```

В ответе может быть `next_cursor` для следующей страницы.

---

#### PUT `/api/v1/tasks/{id}`

Обновить задачу. Изменения пишутся в `task_history`.

**Body** (все поля опциональны):

```json
{
  "title": "Настроить CI/CD",
  "description": "GitHub Actions + Docker",
  "status": "in_progress",
  "priority": "critical",
  "assignee_id": 2,
  "due_date": "2026-07-20"
}
```

Снять исполнителя:

```json
{
  "assignee_id": 0
}
```

Только статус:

```json
{
  "status": "in_progress"
}
```

**Response `200`:** объект задачи (как в create).

---

#### GET `/api/v1/tasks/{id}/history`

История изменений задачи.

**Body:** нет

**Response `200`:**

```json
[
  {
    "id": 1,
    "task_id": 1,
    "changed_by": 1,
    "field_name": "status",
    "old_value": "new",
    "new_value": "in_progress",
    "changed_at": "2026-07-05T12:30:00Z"
  }
]
```

---

### Reports — `http://localhost:8082`

Все эндпоинты требуют `Authorization: Bearer <token>`. **Body:** нет.

#### GET `/api/v1/reports/team-stats`

По каждой команде: название, число участников, задачи `done` за 7 дней.

**Response `200`:**

```json
[
  {
    "team_id": 1,
    "team_name": "Backend Team",
    "members_count": 2,
    "done_tasks_last_7_days": 3
  }
]
```

---

#### GET `/api/v1/reports/top-creators`

Топ-3 создателей задач в каждой команде за месяц.

**Response `200`:**

```json
[
  {
    "team_id": 1,
    "user_id": 1,
    "username": "owner",
    "tasks_created": 5,
    "rank": 1
  }
]
```

---

#### GET `/api/v1/reports/invalid-assignees`

Задачи, где assignee не является членом команды.

**Response `200`:**

```json
[
  {
    "task_id": 4,
    "title": "Битая задача",
    "team_id": 1,
    "assignee_id": 99
  }
]
```

---

## Пример полного flow

```text
1. POST :8080/api/v1/register          → токен owner
2. POST :8080/api/v1/register          → второй пользователь (member)
3. POST :8081/api/v1/teams             → создать команду
4. POST :8081/api/v1/teams/1/invite    → пригласить member
5. POST :8082/api/v1/tasks             → создать задачу
6. GET  :8082/api/v1/tasks?team_id=1   → список задач
7. PUT  :8082/api/v1/tasks/1           → обновить задачу
8. GET  :8082/api/v1/tasks/1/history   → история
9. GET  :8082/api/v1/reports/team-stats
```

## Тесты

Unit-тесты (без Docker):

```bash
cd services/auth  && go test ./internal/usecase/... -cover
cd services/teams && go test ./internal/usecase/... -cover
cd services/tasks && go test ./internal/usecase/... -cover
```

Интеграционные тесты (нужен Docker):

```bash
cd services/auth  && go test -tags=integration ./internal/adapter/repository/mysql/... -timeout 5m
cd services/teams && go test -tags=integration ./internal/adapter/repository/mysql/... -timeout 5m
cd services/tasks && go test -tags=integration ./internal/adapter/repository/mysql/... -timeout 5m
```

## Структура проекта

```text
services/
  auth/    — аутентификация
  teams/   — команды
  tasks/   — задачи и отчёты
docker-compose.yml
Dockerfile.auth
Dockerfile.teams
Dockerfile.tasks
```
