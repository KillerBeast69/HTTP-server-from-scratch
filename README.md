# Chirpy 🐦

## What is Chirpy?
Chirpy is a lightweight, Twitter-like RESTful API built entirely in Go. It provides a robust backend for users to register, securely authenticate, and post short text messages known as "Chirps". The application includes features like an automated profanity filter, token-based authentication (JWTs and refresh tokens), and webhook integrations for premium user tiers (Chirpy Red).

## Why Should You Care?
Chirpy demonstrates clean, scalable backend architecture in Go without relying on heavy web frameworks. It showcases:
* **Standard Library Power**: Built using Go 1.22's enhanced `net/http` ServeMux.
* **Robust Security**: Utilizes Argon2 for password hashing, and implements short-lived JWTs alongside long-lived refresh tokens.
* **Type-Safe Database Interactions**: Uses `sqlc` to generate type-safe Go code from raw SQL queries, safely interacting with a PostgreSQL database.
* **Clean Data Flow**: Includes custom middleware for metrics tracking and a structured approach to stateful HTTP handlers.

It serves as an excellent reference codebase for anyone looking to understand how to build a production-ready API in Go from scratch.

## How to Install and Run

### Prerequisites
* [Go](https://golang.org/doc/install) (version 1.22.2 or higher)
* [PostgreSQL](https://www.postgresql.org/download/)
* [Goose](https://github.com/pressly/goose) (for database migrations)

### 1. Clone the repository
```bash
git clone https://github.com/KillerBeast69/HTTP-server-from-scratch
cd chirpy
```

### 2. Set up the Database
Create a new PostgreSQL database for the project. Then, apply the schema migrations using Goose:
```bash
cd sql/schema
goose postgres "postgres://username:password@localhost:5432/chirpy" up
cd ../..
```

### 3. Configure Environment Variables
Create a `.env` file in the root directory of the project and provide the following configuration:
```env
DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
PLATFORM="dev"
JWT_SECRET="your-super-secret-jwt-key"
POLKA_KEY="your-polka-webhook-secret"
```

### 4. Build and Run
Download the required dependencies and start the server:
```bash
go mod download
go build -o chirpy
./chirpy
```
The server will start on `http://localhost:8080`.

### API Endpoints
* **`GET /api/healthz`** - Check server health.
* **`POST /api/users`** - Create a new user account.
* **`POST /api/login`** - Authenticate and receive JWT/Refresh tokens.
* **`POST /api/chirps`** - Post a new chirp (max 140 characters, filters profanity like "kerfuffle", "sharbert", "fornax").
* **`GET /api/chirps`** - Retrieve all chirps (supports `?author_id=` and `?sort=` parameters).
* **`DELETE /api/chirps/{chirpID}`** - Delete your own chirp.
* **`POST /api/polka/webhooks`** - Webhook to upgrade a user to Chirpy Red.
