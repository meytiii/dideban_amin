INSERT INTO admin_users (
    username,
    password_hash,
    full_name,
    role_title,
    role,
    is_active,
    is_superuser,
    password_changed_at,
    must_change_password,
    totp_secret,
    totp_enabled,
    totp_enforced,
    created_at
) VALUES (
    'admin_it',
    '$argon2id$v=19$m=65536,t=3,p=4$n/aMzaUvqvyoDxpdFiFWvw$QZrVlF1xpZj1SaHCAjC69r3qjM6ABKR09ifE1JdQkZs',
    'مدیر ارشد فناوری',
    'مدیریت فناوری و امنیت اطلاعات',
    'superuser',
    true,
    true,
    NOW(),
    true,
    NULL,
    false,
    true,
    NOW()
) ON CONFLICT (username) DO NOTHING;
