SET
  statement_timeout = 0;

--bun:split
-- 
-- go-webauthnのCredential(= 認証器)構造体を保存する。
-- 
-- NOTE: 半構造化データなので、無理せずNoSQLに保存したほうがいいかもしれない。
--       ユースケース的に認証器は一度登録すると設定変更は考えづらく、再登録してもらうことになりそうなので、
--       一旦arrayやjsonbを使ったリレーションとして保存してみる。
-- 
-- SEE: https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.8.6/webauthn#Credential
-- 
CREATE TABLE webauthn_credentials (
  id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  credential_id TEXT NOT NULL,
  public_key TEXT NOT NULL,
  attestation_type VARCHAR(255) NOT NULL,
  transports VARCHAR(255) [] NOT NULL,
  flags JSONB NOT NULL,
  authenticator JSONB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

--bun:split