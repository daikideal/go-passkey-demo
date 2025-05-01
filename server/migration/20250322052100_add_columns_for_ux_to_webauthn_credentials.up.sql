SET
    statement_timeout = 0;

--bun:split
ALTER TABLE webauthn_credentials
ADD COLUMN IF NOT EXISTS used_count INTEGER NOT NULL DEFAULT 0,
-- NOTE: NULLの場合は認証に使用した実績がないとみなせるので、NULL許容にする。
ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP;

--bun:split