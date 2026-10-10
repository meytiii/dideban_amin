package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(connStr string) {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(2 * time.Minute)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = DB.PingContext(pingCtx); err != nil {
		log.Fatalf("Failed to establish ping connection to PostgreSQL: %v", err)
	}

	log.Println("Successfully initialized native PostgreSQL connection pool.")

	runMigrations()

	if Sessions != nil {
		Sessions.init(DB)
	}
	if Lockouts != nil {
		Lockouts.init(DB)
	}
	if UsedCaptchas != nil {
		UsedCaptchas.init(DB)
	}

	enforceImmutableAuditLogs()
}

func runMigrations() {
	migrationSQL := `
	DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'priority'
		) THEN
			ALTER TABLE reports ADD COLUMN priority VARCHAR(32) NOT NULL DEFAULT 'normal';
			CREATE INDEX IF NOT EXISTS idx_reports_priority ON reports(priority);
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'internal_notes'
		) THEN
			ALTER TABLE reports ADD COLUMN internal_notes TEXT DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'admin_response'
		) THEN
			ALTER TABLE reports ADD COLUMN admin_response TEXT DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'responded_by'
		) THEN
			ALTER TABLE reports ADD COLUMN responded_by VARCHAR(64) DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'responded_at'
		) THEN
			ALTER TABLE reports ADD COLUMN responded_at TIMESTAMPTZ DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_users' AND column_name = 'password_changed_at'
		) THEN
			ALTER TABLE admin_users ADD COLUMN password_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_users' AND column_name = 'must_change_password'
		) THEN
			ALTER TABLE admin_users ADD COLUMN must_change_password BOOLEAN NOT NULL DEFAULT FALSE;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_users' AND column_name = 'role'
		) THEN
			ALTER TABLE admin_users ADD COLUMN role VARCHAR(32) NOT NULL DEFAULT 'edare_kol';
			UPDATE admin_users SET role = 'superuser' WHERE is_superuser = true;
			UPDATE admin_users SET role = 'edare_kol' WHERE username IN ('admin_herasat', 'admin_personeli');
			UPDATE admin_users SET role = 'navahi_manategh' WHERE username = 'nahie1';
			UPDATE admin_users SET role = 'edare_kol' WHERE role IS NULL OR role = '';
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'read_by'
		) THEN
			ALTER TABLE reports ADD COLUMN read_by VARCHAR(64) DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'has_new_message'
		) THEN
			ALTER TABLE reports ADD COLUMN has_new_message BOOLEAN NOT NULL DEFAULT FALSE;
			CREATE INDEX IF NOT EXISTS idx_reports_has_new_message ON reports(has_new_message);
			CREATE INDEX IF NOT EXISTS idx_reports_updated_at ON reports(updated_at DESC);
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'evidence_files' AND column_name = 'file_name'
		) THEN
			ALTER TABLE evidence_files ADD COLUMN file_name TEXT DEFAULT '';
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'evidence_files' AND column_name = 'file_size'
		) THEN
			ALTER TABLE evidence_files ADD COLUMN file_size BIGINT DEFAULT 0;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'classification'
		) THEN
			ALTER TABLE reports ADD COLUMN classification VARCHAR(32) DEFAULT NULL;
			CREATE INDEX IF NOT EXISTS idx_reports_classification ON reports(classification);
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'classified_by'
		) THEN
			ALTER TABLE reports ADD COLUMN classified_by VARCHAR(64) DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'reports' AND column_name = 'classified_at'
		) THEN
			ALTER TABLE reports ADD COLUMN classified_at TIMESTAMPTZ DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_users' AND column_name = 'totp_secret'
		) THEN
			ALTER TABLE admin_users ADD COLUMN totp_secret TEXT DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_users' AND column_name = 'totp_enabled'
		) THEN
			ALTER TABLE admin_users ADD COLUMN totp_enabled BOOLEAN NOT NULL DEFAULT FALSE;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_users' AND column_name = 'totp_enforced'
		) THEN
			ALTER TABLE admin_users ADD COLUMN totp_enforced BOOLEAN NOT NULL DEFAULT FALSE;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_users' AND column_name = 'last_login_at'
		) THEN
			ALTER TABLE admin_users ADD COLUMN last_login_at TIMESTAMPTZ DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'audit_logs' AND column_name = 'prev_hash'
		) THEN
			ALTER TABLE audit_logs ADD COLUMN prev_hash VARCHAR(64) NOT NULL DEFAULT '0000000000000000000000000000000000000000000000000000000000000000';
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'audit_logs' AND column_name = 'entry_hash'
		) THEN
			ALTER TABLE audit_logs ADD COLUMN entry_hash VARCHAR(64) NOT NULL DEFAULT '';
			CREATE INDEX IF NOT EXISTS idx_audit_logs_entry_hash ON audit_logs(entry_hash);
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'admin_password_history'
		) THEN
			CREATE TABLE admin_password_history (
				id BIGSERIAL PRIMARY KEY,
				admin_username VARCHAR(64) DEFAULT '',
				user_id BIGINT DEFAULT NULL,
				password_hash TEXT NOT NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_password_history' AND column_name = 'user_id'
		) THEN
			ALTER TABLE admin_password_history ADD COLUMN user_id BIGINT DEFAULT NULL;
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'admin_password_history' AND column_name = 'admin_username'
		) THEN
			ALTER TABLE admin_password_history ADD COLUMN admin_username VARCHAR(64) DEFAULT '';
		END IF;

		ALTER TABLE admin_password_history ALTER COLUMN admin_username DROP NOT NULL;

		UPDATE admin_password_history h
		SET user_id = u.id
		FROM admin_users u
		WHERE h.admin_username = u.username AND h.user_id IS NULL;

		UPDATE admin_password_history h
		SET admin_username = u.username
		FROM admin_users u
		WHERE h.user_id = u.id AND (h.admin_username IS NULL OR h.admin_username = '');
	END $$;

	CREATE TABLE IF NOT EXISTS admin_password_history (
		id BIGSERIAL PRIMARY KEY,
		admin_username VARCHAR(64) DEFAULT '',
		user_id BIGINT DEFAULT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_pwd_hist_user_created ON admin_password_history(admin_username, created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_pwd_hist_user_id ON admin_password_history(user_id, created_at DESC);

	CREATE TABLE IF NOT EXISTS report_forwardings (
		id BIGSERIAL PRIMARY KEY,
		report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
		forwarded_by VARCHAR(64) NOT NULL,
		recipient_username VARCHAR(64) NOT NULL,
		forwarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		is_viewed BOOLEAN NOT NULL DEFAULT FALSE,
		viewed_at TIMESTAMPTZ DEFAULT NULL,
		response_text TEXT DEFAULT NULL,
		responded_at TIMESTAMPTZ DEFAULT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	ALTER TABLE report_forwardings ADD COLUMN IF NOT EXISTS response_text TEXT DEFAULT NULL;
	ALTER TABLE report_forwardings ADD COLUMN IF NOT EXISTS responded_at TIMESTAMPTZ DEFAULT NULL;
	CREATE INDEX IF NOT EXISTS idx_report_forwardings_report ON report_forwardings(report_id);
	CREATE INDEX IF NOT EXISTS idx_report_forwardings_recipient ON report_forwardings(recipient_username);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_report_forwardings_unique ON report_forwardings(report_id, recipient_username);

	CREATE TABLE IF NOT EXISTS report_messages (
		id BIGSERIAL PRIMARY KEY,
		report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
		sender_type VARCHAR(16) NOT NULL,
		message_text TEXT NOT NULL,
		admin_username VARCHAR(64) DEFAULT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_report_messages_report ON report_messages(report_id);

	CREATE TABLE IF NOT EXISTS admin_notifications (
		id BIGSERIAL PRIMARY KEY,
		notif_type VARCHAR(50) NOT NULL,
		title VARCHAR(255) NOT NULL,
		message TEXT NOT NULL,
		target_report_id BIGINT REFERENCES reports(id) ON DELETE CASCADE,
		target_tracking_code VARCHAR(100),
		recipient_username VARCHAR(64) DEFAULT NULL,
		is_read BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	ALTER TABLE admin_notifications ADD COLUMN IF NOT EXISTS recipient_username VARCHAR(64) DEFAULT NULL;
	CREATE INDEX IF NOT EXISTS idx_admin_notifs_is_read ON admin_notifications(is_read);
	CREATE INDEX IF NOT EXISTS idx_admin_notifs_created_at ON admin_notifications(created_at);
	CREATE INDEX IF NOT EXISTS idx_admin_notifs_recipient ON admin_notifications(recipient_username);

	CREATE TABLE IF NOT EXISTS admin_report_views (
		id BIGSERIAL PRIMARY KEY,
		admin_username VARCHAR(64) NOT NULL,
		report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
		viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_viewed_message_id BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(admin_username, report_id)
	);
	CREATE INDEX IF NOT EXISTS idx_admin_report_views_user ON admin_report_views(admin_username);
	CREATE INDEX IF NOT EXISTS idx_admin_report_views_report ON admin_report_views(report_id);

	INSERT INTO admin_report_views (admin_username, report_id, viewed_at)
	SELECT read_by, id, COALESCE(read_at, NOW())
	FROM reports
	WHERE is_read = true AND read_by IS NOT NULL AND read_by != ''
	ON CONFLICT (admin_username, report_id) DO NOTHING;

	CREATE TABLE IF NOT EXISTS admin_sessions (
		token_hash CHAR(64) PRIMARY KEY,
		username VARCHAR(64) NOT NULL,
		full_name VARCHAR(128) NOT NULL,
		role_title VARCHAR(128) NOT NULL,
		role VARCHAR(32) NOT NULL,
		is_superuser BOOLEAN NOT NULL DEFAULT FALSE,
		ip VARCHAR(45) NOT NULL,
		user_agent TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_active TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		expiry TIMESTAMPTZ NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_admin_sessions_username ON admin_sessions(username);
	CREATE INDEX IF NOT EXISTS idx_admin_sessions_expiry ON admin_sessions(expiry);

	CREATE TABLE IF NOT EXISTS admin_session_kicks (
		token_hash CHAR(64) PRIMARY KEY,
		reason VARCHAR(64) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS admin_login_attempts (
		username VARCHAR(64) PRIMARY KEY,
		consecutive_failures INT NOT NULL DEFAULT 0,
		locked_until TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS used_captcha_tokens (
		signature_b64 TEXT PRIMARY KEY,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_used_captcha_expires ON used_captcha_tokens(expires_at);

	CREATE OR REPLACE FUNCTION touch_reports_updated_at()
	RETURNS TRIGGER AS $$
	BEGIN
		IF NEW.updated_at IS NOT DISTINCT FROM OLD.updated_at THEN
			NEW.updated_at := NOW();
		END IF;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS trg_reports_updated_at ON reports;
	CREATE TRIGGER trg_reports_updated_at
	BEFORE UPDATE ON reports
	FOR EACH ROW
	EXECUTE FUNCTION touch_reports_updated_at();

	DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'report_forwardings' AND column_name = 'updated_at'
		) THEN
			DROP TRIGGER IF EXISTS trg_report_forwardings_updated_at ON report_forwardings;
			CREATE TRIGGER trg_report_forwardings_updated_at
			BEFORE UPDATE ON report_forwardings
			FOR EACH ROW
			EXECUTE FUNCTION touch_reports_updated_at();
		END IF;
	END $$;

	ALTER TABLE report_messages ADD COLUMN IF NOT EXISTS admin_username VARCHAR(64) DEFAULT NULL;

	INSERT INTO report_messages (report_id, sender_type, message_text, admin_username, created_at)
	SELECT id, 'admin', admin_response, COALESCE(NULLIF(responded_by, ''), 'admin'), COALESCE(responded_at, created_at)
	FROM reports
	WHERE admin_response IS NOT NULL AND admin_response != ''
	  AND NOT EXISTS (
	      SELECT 1 FROM report_messages rm WHERE rm.report_id = reports.id AND rm.sender_type = 'admin'
	  );
	`
	if _, err := DB.Exec(migrationSQL); err != nil {
		log.Printf("Warning: Schema migration notice: %v", err)
	} else {
		log.Println("Database schema auto-migrations applied successfully.")
	}
}

func enforceImmutableAuditLogs() {
	query := `
	CREATE OR REPLACE FUNCTION prevent_audit_log_modification()
	RETURNS TRIGGER AS $$
	BEGIN
		RAISE EXCEPTION 'AFTA SECURITY VIOLATION: Audit logs are immutable (WORM) and cannot be updated or deleted.';
	END;
	$$ LANGUAGE plpgsql;

	DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM pg_trigger WHERE tgname = 'trg_immutable_audit_logs'
		) THEN
			CREATE TRIGGER trg_immutable_audit_logs
			BEFORE UPDATE OR DELETE ON audit_logs
			FOR EACH ROW
			EXECUTE FUNCTION prevent_audit_log_modification();
		END IF;

		IF NOT EXISTS (
			SELECT 1 FROM pg_trigger WHERE tgname = 'trg_immutable_audit_logs_truncate'
		) THEN
			CREATE TRIGGER trg_immutable_audit_logs_truncate
			BEFORE TRUNCATE ON audit_logs
			FOR EACH STATEMENT
			EXECUTE FUNCTION prevent_audit_log_modification();
		END IF;
	END $$;
	`
	if _, err := DB.Exec(query); err != nil {
		log.Printf("Notice: Immutable audit log trigger status: %v", err)
	} else {
		log.Println("Immutable audit log security constraints active (AFTA WORM Compliant).")
	}
}
