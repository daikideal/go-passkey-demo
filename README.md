# go-passkey-demo

## 環境構築

### 前提

- docker がインストールされていること

### コンテナをビルド

```bash
docker compose up -d --build
```

### bun × postgres のセットアップ

```bash
go run ./migration db init
```

```bash
go run ./migration db migrate
```

## ローカルホスト起動

docker-compose で以下のコンテナを起動:

- postgres
- redis

```bash
docker compose up -d
```

echo サーバーを起動:

```bash
go run .
```

以下の URL にアクセス:

http://localhost:8080/
