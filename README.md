# CLI Auth — Containerized Login System with Optional 2FA

A secure command-line login system built with Go, featuring user registration, authentication, optional TOTP-based two-factor authentication, account lockout, and session management. Runs in Docker containers with PostgreSQL for persistence.


## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose v2
- (Optional) [Go 1.24+](https://go.dev/dl/) — only needed for local development outside Docker

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/<your-username>/cli-auth.git
cd cli-auth
```

### 2. Create your environment file

```bash
cp .env.example .env
```

Edit `.env` and set a strong database password:

```dotenv
POSTGRES_USER=cliauth
POSTGRES_PASSWORD=your_strong_password_here
POSTGRES_DB=cliauth
SESSION_TIMEOUT=30m
```

### 3. Build and start the containers

```bash
docker compose up --build
```

This will:
- Pull the PostgreSQL 16 image
- Download Go dependencies and build the CLI binary (multi-stage Docker build)
- Start both containers (DB + CLI)
- Run database migrations automatically on first start

### 4. Attach to the CLI

In a **second terminal** (also recommended to use terminal in full screen for the QR to completely show up in case of enabling 2FA):

```bash
docker compose exec cli cli-auth
```

You'll see the interactive menu. Use arrow keys to navigate, Enter to select.

### 5. Stop the containers

```bash
docker compose down
```

Data persists in Docker volumes (`pg-data`, `cli-data`). To wipe everything:

```bash
docker compose down -v
```


## Architecture

```
┌─────────────────────────────────────────────┐
│  cli container                              │
│  ┌───────────────────────────────────────┐  │
│  │  Bubbletea TUI                        │  │
│  │  ┌─────────┐ ┌────────┐ ┌──────────┐ │  │
│  │  │  Auth   │ │Session │ │ Lockout  │ │  │
│  │  │ Service │ │ Store  │ │ Service  │ │  │
│  │  └────┬────┘ └───┬────┘ └────┬─────┘ │  │
│  │       └──────────┼───────────┘        │  │
│  │              ┌───┴────┐               │  │
│  │              │  User  │               │  │
│  │              │ Store  │               │  │
│  │              └───┬────┘               │  │
│  └──────────────────┼────────────────────┘  │
│                     │                        │
│  /data/.session     │  (session persistence) │
└─────────────────────┼────────────────────────┘
                      │ TCP :5432
┌─────────────────────┼────────────────────────┐
│  db container       │                        │
│  ┌──────────────────┴───────────────────┐   │
│  │         PostgreSQL 16                │   │
│  │  ┌──────────┐  ┌──────────────────┐  │   │
│  │  │  users   │  │    sessions      │  │   │
│  │  └──────────┘  └──────────────────┘  │   │
│  └──────────────────────────────────────┘   │
│  /var/lib/postgresql/data (volume)          │
└──────────────────────────────────────────────┘
```

## Database Schema

The application uses two tables in PostgreSQL: `users` and `sessions`.

### Users Table

| Column          | Type          | Constraints / Default              |
|-----------------|---------------|------------------------------------|
| `id`            | `SERIAL`      | Primary Key                        |
| `username`      | `VARCHAR(64)` | Unique, Not Null                   |
| `password_hash` | `TEXT`         | Not Null                           |
| `totp_secret`   | `TEXT`         | Nullable                           |
| `totp_enabled`  | `BOOLEAN`     | Not Null, Default `FALSE`          |
| `failed_attempts`| `INTEGER`    | Not Null, Default `0`              |
| `locked_until`  | `TIMESTAMPTZ` | Nullable                           |
| `created_at`    | `TIMESTAMPTZ` | Not Null, Default `NOW()`          |
| `last_login_at` | `TIMESTAMPTZ` | Nullable                           |

### Sessions Table

| Column       | Type           | Constraints / Default                              |
|--------------|----------------|----------------------------------------------------|
| `id`         | `UUID`         | Primary Key, Default `gen_random_uuid()`           |
| `user_id`    | `INTEGER`      | Not Null, Foreign Key → `users(id)` On Delete Cascade |
| `token`      | `VARCHAR(128)` | Unique, Not Null                                   |
| `created_at` | `TIMESTAMPTZ`  | Not Null, Default `NOW()`                          |
| `expires_at` | `TIMESTAMPTZ`  | Not Null                                           |

### Indexes

| Index Name              | Table      | Column     |
|-------------------------|------------|------------|
| `idx_sessions_token`    | `sessions` | `token`    |
| `idx_sessions_user_id`  | `sessions` | `user_id`  |
| `idx_users_username`    | `users`    | `username` |

## Commands

### Before Login

| Command    | Description              |
|------------|--------------------------|
| `register` | Create a new account     |
| `login`    | Login with credentials   |
| `help`     | Show available commands  |
| `exit`     | Quit the program         |

### After Login

| Command       | Description                    |
|---------------|--------------------------------|
| `whoami`      | Show your account details      |
| `enable-2fa`  | Enable TOTP-based 2FA         |
| `disable-2fa` | Disable 2FA                    |
| `logout`      | End your session               |
| `help`        | Show available commands        |
| `exit`        | Quit the program               |

## Enabling 2FA

1. Log in and type `enable-2fa`
2. Scan the QR code with your authenticator app (Google Authenticator, Authy, etc.)
3. Enter the 6-digit code from the app to verify
4. On next login, you'll be prompted for the TOTP code after your password

## Features

- **Secure passwords**: Hashed with bcrypt (cost 12)
- **TOTP 2FA**: Google Authenticator compatible; QR code rendered in terminal
- **Account lockout**: 5 failed attempts → 15 min lockout
- **Session management**: Configurable timeout (default 30 min), persists across container restarts
- **Interactive TUI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## Environment Variables

| Variable           | Default          | Description                    |
|--------------------|------------------|--------------------------------|
| `POSTGRES_USER`    | `cliauth`        | PostgreSQL username            |
| `POSTGRES_PASSWORD`| —                | PostgreSQL password (required) |
| `POSTGRES_DB`      | `cliauth`        | PostgreSQL database name       |
| `DB_HOST`          | `localhost`      | DB host (set to `db` in Docker)|
| `DB_PORT`          | `5432`           | PostgreSQL port                |
| `SESSION_TIMEOUT`  | `30m`            | Session expiration duration    |
| `SESSION_FILE_PATH`| `/data/.session` | Path to session token file     |

## Project Structure

```
cli-auth/
├── cmd/cli-auth/
│   └── main.go              # Entrypoint, wires dependencies
├── internal/
│   ├── auth/
│   │   ├── auth.go           # Login, register orchestration
│   │   ├── password.go       # bcrypt wrapper + complexity validation
│   │   └── totp.go           # TOTP enrollment + validation
│   ├── session/
│   │   ├── session.go        # Session model
│   │   └── store.go          # DB + file operations for sessions
│   ├── user/
│   │   ├── user.go           # User model
│   │   └── store.go          # DB operations for users
│   ├── lockout/
│   │   └── lockout.go        # Failed attempt tracking, lock/unlock
│   ├── db/
│   │   ├── db.go             # Connection init, migrations
│   │   └── schema.sql        # Embedded database schema
│   └── tui/
│       ├── app.go            # Bubbletea model, Update/View state machine
│       ├── views.go          # Screen renderers
│       ├── commands.go       # Async tea.Cmd functions
│       └── styles.go         # Lip Gloss styles
├── migrations/
│   └── 001_init.sql          # Database schema (reference copy)
├── Dockerfile                # Multi-stage Go build
├── docker-compose.yml        # PostgreSQL + CLI orchestration
├── .env.example              # Template for environment variables
└── README.md
```

## Tech Stack

- **Go 1.24** — application language
- **PostgreSQL 16** — persistence
- **Bubbletea / Lip Gloss** — terminal UI
- **pgx/v5** — PostgreSQL driver (pure Go)
- **pquerna/otp** — TOTP generation & validation
- **mdp/qrterminal** — QR code rendering in terminal
- **golang.org/x/crypto/bcrypt** — password hashing
