# Didebane Amin (دیده‌بان امین)

Stack: Go + Native HTML/CSS/JS + PostgreSQL.

## Running Locally with Docker (No Local PostgreSQL Installation Required)

### 1. Prerequisites
- Docker Desktop installed and running
- Go 1.22+ installed

### 2. Start PostgreSQL via Docker Compose
Ensure Docker Desktop is open and running, then execute:
```bash
docker compose up -d
```
This automatically boots PostgreSQL on port `5432` and executes:
- `docker/init/01_schema.sql`: Full 12-table schema, WORM audit triggers, indexes.
- `docker/init/02_seed.sql`: Seeds default superuser (`admin_it`).

### 3. Start the Backend Application
Make sure `.env` is configured (or copy from `.env.example`):
```bash
go run .
```

### 4. Access the Services
- **Public Reporting Portal**: [http://localhost:8080/](http://localhost:8080/)
- **Admin Cockpit**: [http://localhost:8080/admin/](http://localhost:8080/admin/)

#### Initial Administrative Credentials:
On initial application bootstrap, if no superuser exists, the initial administrator (`admin_it`) is provisioned. Set `HERASAT_INITIAL_ADMIN_PASSWORD` in your environment, or the backend generates a random secure password and prints it once to stdout upon first boot. Password change is enforced immediately on initial login.

### 5. Stopping the Database
To stop the PostgreSQL container:
```bash
docker compose down
```
To stop and wipe database data volume:
```bash
docker compose down -v
```

## Testing

Run tests across the suite:
```bash
go test -v ./...
```
