package main

import (
	"time"

	"github.com/daikideal/go-passkey-demo/db"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	e := echo.New()

	e.GET("/", hello())
	e.POST("/users", createUser())
	e.GET("/users", getUser())

	e.Logger.Fatal(e.Start(":8080"))
}

func hello() echo.HandlerFunc {
	res := struct {
		Message string `json:"message"`
	}{
		Message: "Hello, World!",
	}

	return func(ctx echo.Context) error {
		return ctx.JSON(200, res)
	}
}

type uuid = string

type User struct {
	ID        uuid      `json:"id" bun:"id"`
	Name      string    `json:"name" bun:"name"`
	Email     string    `json:"email" bun:"email"`
	Password  string    `json:"password" bun:"password"`
	CreatedAt time.Time `json:"created_at" bun:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at"`
}

func getUser() echo.HandlerFunc {
	db := db.GetDB()

	return func(ctx echo.Context) error {
		ctx.Logger().Info("GET /user")

		var res []*User
		if err := db.NewSelect().
			Model(&res).
			Column("*").
			Scan(ctx.Request().Context()); err != nil {
			ctx.Logger().Errorf("Failed to select user: %v\n", err)
		}

		return ctx.JSON(200, res)
	}
}

func createUser() echo.HandlerFunc {
	db := db.GetDB()

	return func(ctx echo.Context) error {
		ctx.Logger().Info("POST /user")

		mockedReqUser := &User{
			Name:     "test",
			Email:    "test@example.com",
			Password: "testpass",
		}

		res, err := db.NewInsert().
			Model(mockedReqUser).
			Column("name", "email", "password").
			Returning("*").
			Exec(ctx.Request().Context())
		if err != nil {
			ctx.Logger().Errorf("Failed to insert user: %v\n", err)
			return ctx.JSON(500, res)
		}

		ctx.Logger().Infof("Success to insert user: %+v\n", res)

		return ctx.JSON(201, res)
	}
}
