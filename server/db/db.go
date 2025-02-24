package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var (
	// TODO: 環境変数から取得するように変更
	dsn = fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password='%s' sslmode=disable search_path=%s",
		"postgres",
		5432,
		"mydb",
		"myuser",
		"mypassword",
		"myschema",
	)

	db *bun.DB
)

func init() {
	initDB()
}

func GetDB() *bun.DB {
	return db
}

func initDB() {
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	db = bun.NewDB(sqldb, pgdialect.New())
}
