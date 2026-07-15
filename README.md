# Chirpy

Chirpy is a lightweight microblogging HTTP API built with Go and PostgreSQL. It
started as a small HTTP server and grew into an authenticated backend where
users can publish and manage short messages called chirps.

## What the Project Does

Chirpy provides a JSON API for:

- Registering users and updating account credentials
- Hashing passwords with Argon2id
- Authenticating with JWT access tokens and persistent refresh tokens
- Creating, listing, filtering, sorting, and deleting chirps
- Restricting chirps to 140 characters and filtering selected words
- Ensuring users can delete only their own chirps
- Upgrading accounts through an API-key-authenticated Polka webhook
- Reporting service health and basic in-memory page-view metrics

Data is stored in PostgreSQL. Database migrations are managed with Goose, and
sqlc generates type-safe Go code from the SQL queries.

The main routes are:

| Method   | Route                   | Purpose                                      |
| -------- | ----------------------- | -------------------------------------------- |
| `POST`   | `/api/users`            | Register a user                              |
| `PUT`    | `/api/users`            | Update the authenticated user                |
| `POST`   | `/api/login`            | Log in and receive access and refresh tokens |
| `POST`   | `/api/refresh`          | Exchange a refresh token for an access token |
| `POST`   | `/api/revoke`           | Revoke a refresh token                       |
| `POST`   | `/api/chirps`           | Create a chirp                               |
| `GET`    | `/api/chirps`           | List, filter, and sort chirps                |
| `GET`    | `/api/chirps/{chirpID}` | Retrieve one chirp                           |
| `DELETE` | `/api/chirps/{chirpID}` | Delete an owned chirp                        |
| `POST`   | `/api/polka/webhooks`   | Process premium-account upgrades             |
| `GET`    | `/api/healthz`          | Check that the HTTP service is running       |

A minimal static page is available under `/app/`. Chirpy is primarily an API,
not a complete social-network frontend.

## Why You Might Be Interested

Chirpy is a compact example of backend development using Go's standard
`net/http` package rather than a web framework. It demonstrates how common
service concerns fit together in one codebase:

- Relational schema design and foreign-key relationships in PostgreSQL
- Versioned database migrations and generated, type-safe query code
- Secure password storage, JWT validation, and refresh-token revocation
- Bearer-token authorization and resource ownership checks
- Webhook integration with API-key authentication
- Middleware, atomic in-memory metrics, and JSON HTTP handlers

The project is useful as a learning reference. It is not a
production-ready social platform: the frontend is intentionally minimal, test
coverage is currently limited, and deployment infrastructure is not included.

## How to Install and Run

### Prerequisites

- [Git](https://git-scm.com/)
- [Go 1.26.4](https://go.dev/doc/install) or a compatible newer version
- [PostgreSQL](https://www.postgresql.org/download/)
- [Goose](https://github.com/pressly/goose) for database migrations

sqlc is optional for running the server because generated database code is
committed. Install it only if you plan to change the schema or SQL queries.

### 1. Clone the repository

```bash
git clone https://github.com/KayraBulbul/chirpy.git
cd chirpy
go mod download
```

### 2. Create a PostgreSQL database

Create an empty database named `chirpy`, or use another name and update the
connection string in the next step.

For example, with the PostgreSQL command-line tools:

```bash
createdb chirpy
```

### 3. Configure the environment

Create a `.env` file in the repository root:

```dotenv
DB_URL=postgres://USER:PASSWORD@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET=replace-with-a-long-random-jwt-secret
POLKA_KEY=replace-with-a-webhook-api-key
```

Replace the placeholder credentials and secrets. The server currently requires
this file at startup, and `.env` should never be committed.

`PLATFORM=dev` enables `POST /admin/reset`, which deletes all users and their
related chirps and refresh tokens. Do not enable it in a shared or production
environment.

### 4. Apply the migrations

Use the same connection string configured as `DB_URL`:

```bash
goose -dir sql/schema postgres "postgres://USER:PASSWORD@localhost:5432/chirpy?sslmode=disable" up
```

### 5. Start the server

Run the application from the repository root so it can load `.env` and serve
the static files correctly:

```bash
go run .
```

The server listens on `http://localhost:8080`. Verify it with:

```bash
curl http://localhost:8080/api/healthz
```

The response should be `OK`. The static page is available at
`http://localhost:8080/app/`.

### Tests

```bash
go test ./...
```

### Optional development commands

Regenerate the database package after changing files under `sql/`:

```bash
sqlc generate
```

Build a local executable from the current source:

```bash
go build -o chirpy .
```
