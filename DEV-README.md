# Development Guide

Quick guide for setting up a local Miniflux development environment.

## Prerequisites

- [Nix](https://nixos.org/download.html) with flakes enabled
- [Docker](https://docs.docker.com/get-docker/) (for the development database)
- Optional: [direnv](https://direnv.net/) for automatic environment loading

## Quick Start

### 1. Set up the development environment

**With direnv (automatic):**
```bash
direnv allow
# Environment loads automatically when you cd into the directory
```

**Without direnv (manual):**
```bash
nix develop
# Drops you into a shell with all dependencies available
```

### 2. Start the development database

```bash
./dev-db.sh start
```

This will:
- Create and start a PostgreSQL container
- Set up the database with default credentials
- Show connection information

### 3. Run Miniflux

```bash
# Set the database URL
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/miniflux2?sslmode=disable"

# Run with auto-migration and admin user creation
make run
```

Default admin credentials:
- Username: `admin`
- Password: `test123`

The application will be available at: http://localhost:8080

## Development Workflow

### Database Management

```bash
./dev-db.sh start      # Start the database
./dev-db.sh stop       # Stop the database
./dev-db.sh status     # Check status
./dev-db.sh logs       # View logs
./dev-db.sh psql       # Connect with psql
./dev-db.sh destroy    # Remove container (deletes all data!)
```

### Building

```bash
make miniflux          # Build for current platform
go build               # Alternative (no optimizations)
```

### Testing

```bash
make test              # Run unit tests
make lint              # Run linters
```

### Running with custom options

```bash
# Run with custom DATABASE_URL
DATABASE_URL="postgres://user:pass@host/db?sslmode=disable" go run main.go

# Run with debug logging
LOG_LEVEL=debug go run main.go

# Run migrations manually
RUN_MIGRATIONS=1 go run main.go
```

## Environment Variables

Key environment variables for development:

- `DATABASE_URL` - PostgreSQL connection string (required)
- `LOG_LEVEL` - Logging level: `debug`, `info`, `warning`, `error` (default: `info`)
- `LOG_DATE_TIME` - Show timestamps in logs: `1` or `0` (default: `0`)
- `RUN_MIGRATIONS` - Run database migrations on startup: `1` or `0` (default: `0`)
- `CREATE_ADMIN` - Create admin user on startup: `1` or `0` (default: `0`)
- `ADMIN_USERNAME` - Admin username (default: `admin`)
- `ADMIN_PASSWORD` - Admin password (default: random)

## Project Structure

See [AGENTS.md](./AGENTS.md) for a comprehensive overview of the codebase architecture.

Key directories:
- `internal/ui/` - Web UI handlers (one file per view)
- `internal/template/` - HTML templates
- `internal/storage/` - Database queries
- `internal/model/` - Data structures
- `internal/locale/translations/` - i18n translations (20 languages)

## Tips

### Quick dev loop

```bash
# Terminal 1: Keep the DB running
./dev-db.sh start

# Terminal 2: Run with live reload (requires entr or similar)
ls **/*.go | entr -r make run

# Or just restart manually:
make run
```

### Database connection string

The format is:
```
postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=disable
```

Default for dev:
```
postgres://postgres:postgres@localhost:5432/miniflux2?sslmode=disable
```

### Resetting the database

```bash
./dev-db.sh destroy    # Remove everything
./dev-db.sh start      # Start fresh
make run               # Migrations run automatically with RUN_MIGRATIONS=1
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for contribution guidelines.

## Custom Features

This fork includes custom features maintained as clean commits:

- **Date-based entry view** - Browse entries grouped by publish date (Today, Yesterday, This Week, This Month, Earlier)
  - Route: `/entries/by-date`
  - Handler: `internal/ui/date_entries.go`
  - Template: `internal/template/templates/views/date_entries.html`

Custom features are maintained on feature branches and can be rebased over upstream updates.
