
REVOKE ALL ON SCHEMA public FROM PUBLIC;

CREATE TABLE IF NOT EXISTS reports (
    id BIGSERIAL PRIMARY KEY,
    tracking_code TEXT NOT NULL,
    tracking_code_hash VARCHAR(64) DEFAULT NULL,
    category VARCHAR(64) NOT NULL,
    report_type VARCHAR(32) NOT NULL,
    priority VARCHAR(32) NOT NULL DEFAULT 'normal',
    reporter_name TEXT DEFAULT NULL,
    reporter_phone TEXT DEFAULT NULL,
    description TEXT NOT NULL,
    admin_response TEXT DEFAULT NULL,
    responded_by VARCHAR(64) DEFAULT NULL,
    responded_at TIMESTAMPTZ DEFAULT NULL,
    internal_notes TEXT DEFAULT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ DEFAULT NULL,
    read_by VARCHAR(64) DEFAULT NULL,
    has_new_message BOOLEAN NOT NULL DEFAULT FALSE,
    classification VARCHAR(32) DEFAULT NULL,
    classified_by VARCHAR(64) DEFAULT NULL,
    classified_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reports_tracking_code ON reports(tracking_code);
CREATE INDEX IF NOT EXISTS idx_reports_tracking_hash ON reports(tracking_code_hash);
CREATE INDEX IF NOT EXISTS idx_reports_is_read ON reports(is_read);
CREATE INDEX IF NOT EXISTS idx_reports_priority ON reports(priority);
CREATE INDEX IF NOT EXISTS idx_reports_has_new_message ON reports(has_new_message);
CREATE INDEX IF NOT EXISTS idx_reports_classification ON reports(classification);
CREATE INDEX IF NOT EXISTS idx_reports_updated_at ON reports(updated_at DESC);

CREATE TABLE IF NOT EXISTS evidence_files (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    mime_type VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS admin_users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(128) NOT NULL DEFAULT '',
    role_title VARCHAR(128) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'edare_kol',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_superuser BOOLEAN NOT NULL DEFAULT FALSE,
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
    totp_secret TEXT DEFAULT NULL,
    totp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    totp_enforced BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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

CREATE INDEX IF NOT EXISTS idx_report_forwardings_report ON report_forwardings(report_id);
CREATE INDEX IF NOT EXISTS idx_report_forwardings_recipient ON report_forwardings(recipient_username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_report_forwardings_unique ON report_forwardings(report_id, recipient_username);

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

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    admin_username VARCHAR(64) NOT NULL,
    admin_title VARCHAR(128) NOT NULL,
    action VARCHAR(64) NOT NULL,
    target_report_id BIGINT DEFAULT NULL,
    ip_address VARCHAR(45) NOT NULL,
    detail TEXT DEFAULT '',
    prev_hash VARCHAR(64) NOT NULL DEFAULT '0000000000000000000000000000000000000000000000000000000000000000',
    entry_hash VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_admin ON audit_logs(admin_username);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entry_hash ON audit_logs(entry_hash);

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

CREATE INDEX IF NOT EXISTS idx_admin_notifs_is_read ON admin_notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_admin_notifs_created_at ON admin_notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_admin_notifs_recipient ON admin_notifications(recipient_username);

CREATE TABLE IF NOT EXISTS admin_sessions (
    token_hash CHAR(64) PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    full_name VARCHAR(128) NOT NULL,
    role_title VARCHAR(128) NOT NULL,
    role VARCHAR(32) NOT NULL,
    is_superuser BOOLEAN NOT NULL DEFAULT FALSE,
    ip VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    requires_2fa_setup BOOLEAN NOT NULL DEFAULT FALSE,
    requires_password_change BOOLEAN NOT NULL DEFAULT FALSE,
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
    username VARCHAR(64) NOT NULL,
    ip VARCHAR(45) NOT NULL DEFAULT '',
    consecutive_failures INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (username, ip)
);

CREATE TABLE IF NOT EXISTS used_captcha_tokens (
    signature_b64 TEXT PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_used_captcha_expires ON used_captcha_tokens(expires_at);

CREATE TABLE IF NOT EXISTS used_mfa_nonces (
    nonce VARCHAR(64) PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_used_mfa_nonces_expiry ON used_mfa_nonces(expires_at);

CREATE TABLE IF NOT EXISTS pending_deletions (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    requested_by VARCHAR(64) NOT NULL,
    reason TEXT NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    approved_by VARCHAR(64) DEFAULT NULL,
    approved_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_pending_del_status ON pending_deletions(status);

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

CREATE OR REPLACE FUNCTION prevent_audit_log_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'AFTA SECURITY VIOLATION: Audit logs are immutable (WORM) and cannot be updated or deleted.';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_immutable_audit_logs ON audit_logs;
CREATE TRIGGER trg_immutable_audit_logs
BEFORE UPDATE OR DELETE ON audit_logs
FOR EACH ROW
EXECUTE FUNCTION prevent_audit_log_modification();

DROP TRIGGER IF EXISTS trg_immutable_audit_logs_truncate ON audit_logs;
CREATE TRIGGER trg_immutable_audit_logs_truncate
BEFORE TRUNCATE ON audit_logs
FOR EACH STATEMENT
EXECUTE FUNCTION prevent_audit_log_modification();

REVOKE UPDATE, DELETE, TRUNCATE ON audit_logs FROM PUBLIC;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'herasat_app') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO herasat_app;
        REVOKE UPDATE, DELETE, TRUNCATE ON audit_logs FROM herasat_app;
        GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO herasat_app;
    END IF;
END $$;
