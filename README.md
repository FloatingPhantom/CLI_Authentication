# CLI Auth — Containerized Login System with Optional 2FA

A secure command-line login system built with Go, featuring user registration, authentication, optional TOTP-based two-factor authentication, account lockout, and session management. Runs in Docker containers with PostgreSQL for persistence.

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

## Quick Start

```bash
# Build and start both containers
docker-compose up --build

# In another terminal, attach to the CLI
docker-compose exec cli cli-auth
```

Or run them separately:

```bash
# Start only the database
docker-compose up -d db

# Run the CLI locally (requires Go 1.23+)
export DB_HOST=localhost DB_PORT=5432 DB_USER=cliauth DB_PASSWORD=cliauth_secret DB_NAME=cliauth
export SESSION_FILE_PATH=./.session SESSION_TIMEOUT=30m
go run ./cmd/cli-auth/
```

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

## Features

- **Secure passwords**: Hashed with bcrypt (cost 12)
- **TOTP 2FA**: Google Authenticator compatible; QR code rendered in terminal
- **Account lockout**: 5 failed attempts → 15 min lockout
- **Session management**: Configurable timeout (default 30 min), persists across container restarts
- **Interactive TUI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)

## Environment Variables

| Variable           | Default          | Description                    |
|--------------------|------------------|--------------------------------|
| `DB_HOST`          | `localhost`      | PostgreSQL host                |
| `DB_PORT`          | `5432`           | PostgreSQL port                |
| `DB_USER`          | `cliauth`        | PostgreSQL user                |
| `DB_PASSWORD`      | `cliauth_secret` | PostgreSQL password            |
| `DB_NAME`          | `cliauth`        | PostgreSQL database name       |
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
│   │   ├── password.go       # bcrypt wrapper
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
│   │   └── db.go             # Connection init, migrations
│   └── tui/
│       ├── app.go            # Bubbletea program, model, Update/View
│       ├── views.go          # Screen renderers
│       ├── commands.go       # Async tea.Cmd functions
│       └── styles.go         # Lip Gloss styles
├── migrations/
│   └── 001_init.sql          # Database schema
├── Dockerfile                # Multi-stage Go build
├── docker-compose.yml        # PostgreSQL + CLI orchestration
├── .env                      # Default environment variables
└── README.md
```

## Development

```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build binary
go build -o cli-auth ./cmd/cli-auth/

# Format code
gofmt -w .
```
