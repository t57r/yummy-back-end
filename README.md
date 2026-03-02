# Link to PR:

[Link](https://github.com/t57r/GolangForDevs/pull/15)

# Yummy Backend Docs

## What This App Is

Yummy is a Go backend API for recipe browsing with:

- JWT auth (`/auth/signup`, `/auth/signin`, `/auth/refresh`)
- Recipes listing and details (`/recipes`, `/recipes/{recipeId}`)
- Categories (`/categories`)
- User profile (`/me`)
- Favorites (`/me/favorites`, `/me/favorites/{recipeId}`)
- Static recipe images (`/img/{fileName}`)

Tech stack:

- Go + Fiber v3
- PostgreSQL
- `sqlc` for typed DB queries

---

## Main Dependencies

From `go.mod`:

- `github.com/gofiber/fiber/v3` - HTTP server framework
- `github.com/golang-jwt/jwt/v5` - JWT access token handling
- `github.com/jackc/pgx/v5` - PostgreSQL driver/pool
- `github.com/joho/godotenv` - load `.env` for local development
- `golang.org/x/crypto` - password hashing (`bcrypt`)

Dev tooling used in this repo:

- `sqlc` - generate Go code from SQL queries
- `@redocly/cli` / `redoc-cli` - OpenAPI lint and HTML docs generation
- `@mermaid-js/mermaid-cli` - ER diagram image rendering

---

## Project Structure (Important Parts)

- `cmd/server/main.go` - app bootstrap, routes, middleware, graceful shutdown
- `cmd/server/handlers/` - HTTP handlers
- `cmd/server/middlewares/` - auth middleware
- `internal/db/` - generated sqlc DB layer
- `sql/schema/` - database DDL
- `sql/queries/` - SQL query definitions used by sqlc
- `docs/open_api.yaml` - OpenAPI source
- `docs/open_api.html` - rendered API docs
- `docs/er-diagram.mmd` - Mermaid ER source
- `docs/er-diagram.png` - rendered ER image

---

## Environment Variables

Typical local `.env`:

```env
DATABASE_URL=postgres://yummy:yummy@127.0.0.1:5433/yummy?sslmode=disable
JWT_ACCESS_SECRET=dev_access_secret_change_me
JWT_REFRESH_SECRET=dev_refresh_secret_change_me
ACCESS_TTL_SECONDS=900
REFRESH_TTL_SECONDS=2592000
PORT=8081
MAX_DB_CONNECTIONS=10
```

---

## Run / Build

Run app:

```bash
go run ./cmd/server
```

Build binary:

```bash
go build -o bin/yummy-server ./cmd/server
```

Run tests/build checks:

```bash
go test ./...
```

---

## Database & SQL Generation Workflow

This project uses SQL-first development.

### 1) Update schema (DDL)

Edit files in:

- `sql/schema/001_core.sql`
- `sql/schema/002_users_favorites.sql`
- `sql/schema/003_refresh_tokens.sql`

Apply schema to DB (example):

```bash
psql "$DATABASE_URL" -f sql/schema/001_core.sql
psql "$DATABASE_URL" -f sql/schema/002_users_favorites.sql
psql "$DATABASE_URL" -f sql/schema/003_refresh_tokens.sql
```

### 2) Update queries

Edit SQL query files in `sql/queries/*.sql`.

### 3) Regenerate typed Go DB code

```bash
sqlc generate
```

`sqlc` config is in `sqlc.yaml`, output goes to `internal/db/`.

### 4) Verify

```bash
go test ./...
```

---

## API Docs (OpenAPI)

Source:

- `docs/open_api.yaml`

Lint:

```bash
npx -y @redocly/cli lint docs/open_api.yaml
```

Generate HTML:

```bash
npx -y redoc-cli bundle docs/open_api.yaml -o docs/open_api.html
```

---

## ER Diagram

Mermaid source is in `docs/er-diagram.mmd`.

```mermaid
erDiagram
    USERS {
        BIGSERIAL id PK
        TEXT name
        TEXT email UK
        TEXT password_hash
        TIMESTAMPTZ created_at
    }

    RECIPES {
        BIGINT id PK
        TEXT title
        TEXT image
        TEXT summary
        TEXT source_url
        INT prep_time_minutes
        INT cook_time_minutes
        INT serves
        TEXT[] ingredients
        TEXT[] preparation
        TEXT[] tags
    }

    CATEGORIES {
        SERIAL id PK
        TEXT name UK
    }

    FAVORITES {
        BIGINT user_id PK,FK
        BIGINT recipe_id PK,FK
        TIMESTAMPTZ created_at
    }

    RECIPE_CATEGORIES {
        BIGINT recipe_id PK,FK
        INT category_id PK,FK
    }

    REFRESH_TOKENS {
        BIGSERIAL id PK
        BIGINT user_id FK
        TEXT token_hash UK
        TIMESTAMPTZ expires_at
        TIMESTAMPTZ created_at
        TIMESTAMPTZ revoked_at
    }

    USERS ||--o{ FAVORITES : has
    RECIPES ||--o{ FAVORITES : is_favorited_in
    RECIPES ||--o{ RECIPE_CATEGORIES : has
    CATEGORIES ||--o{ RECIPE_CATEGORIES : groups
    USERS ||--o{ REFRESH_TOKENS : owns
```
