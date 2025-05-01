SET
    statement_timeout = 0;

--bun:split
ALTER TABLE webauthn_credentials
DROP COLUMN IF EXISTS used_count,
-- NOTE: NULLの場合は認証に使用した実績がないとみなせるので、NULL許容にする。
DROP COLUMN IF EXISTS last_used_at;

--bun:split
