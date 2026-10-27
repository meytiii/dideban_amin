DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'herasat_app') THEN
        CREATE ROLE herasat_app WITH LOGIN PASSWORD 'CHANGE_IN_PRODUCTION';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'herasat_migrator') THEN
        CREATE ROLE herasat_migrator WITH LOGIN PASSWORD 'CHANGE_IN_PRODUCTION';
    END IF;
END $$;

GRANT CONNECT ON DATABASE herasat_db TO herasat_app, herasat_migrator;
GRANT USAGE ON SCHEMA public TO herasat_app, herasat_migrator;
GRANT ALL PRIVILEGES ON SCHEMA public TO herasat_migrator;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO herasat_migrator;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO herasat_migrator;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO herasat_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO herasat_app;
REVOKE UPDATE, DELETE, TRUNCATE ON audit_logs FROM herasat_app;
REVOKE UPDATE, DELETE, TRUNCATE ON audit_logs FROM PUBLIC;
