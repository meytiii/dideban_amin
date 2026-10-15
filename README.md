# Didebane Amin (دیده‌بان امین)

Whistleblower and citizen report submission portal for the Security Department (Herasat), General Directorate of Education, Hamadan Province.

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

#### Default Administrative Credentials:
- **Username**: `admin_it`
- **Password**: `Herasat@2026Secure`
- **Role**: `superuser`

### 5. Stopping the Database
To stop the PostgreSQL container:
```bash
docker compose down
```
To stop and wipe database data volume:
```bash
docker compose down -v
```

