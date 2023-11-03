package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// TODO: 環境変数から取得する
var DSN = fmt.Sprintf(
	"host=%s port=%d dbname=%s user=%s password='%s' sslmode=disable search_path=%s",
	"localhost",
	15432,
	"mydb",
	"myuser",
	"mypassword",
	"myschema",
)

func main() {
	db, err := sql.Open("postgres", DSN)
	defer db.Close()

	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
	}

	fmt.Println("Connection established.")
}
