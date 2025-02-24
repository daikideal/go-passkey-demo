# go-passkey-demo

## 環境構築

### 前提

- docker がインストールされていること

### コンテナをビルド

```bash
docker compose build
```

### bun × postgres のセットアップ

```bash
docker compose exec server go run ./migration db init
```

```bash
docker compose exec server go run ./migration db migrate
```

## ローカルホスト起動

docker-compose でコンテナを起動:

```bash
docker compose up --watch
```

※Docker Compose v2.32.4 時点では、`--detach`オプションは`--watch`オプションと併用できない。

以下の URL にアクセス:

http://localhost:5173/

## postgres にログイン

起動した postgres コンテナで psql コマンドを実行し、db にログイン:

```bash
docker compose exec -it postgres psql -d mydb -U myuser
```

スキーマを選択:

```sql
SET search_path TO myschema;
```
