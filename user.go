package main

import (
	"time"

	"github.com/daikideal/go-passkey-demo/db"
	"github.com/labstack/echo/v4"
)

type uuid = string

type User struct {
	ID        uuid      `json:"id" bun:"id"`
	Name      string    `json:"name" bun:"name"`
	Email     string    `json:"email" bun:"email"`
	Password  string    `json:"password" bun:"password"`
	CreatedAt time.Time `json:"created_at" bun:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at"`
}

func getUsers() echo.HandlerFunc {
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

		var user User
		err := ctx.Bind(&user)
		if err != nil {
			ctx.Logger().Errorf("Failed to insert user: %v\n", err)
			return ctx.JSON(400, err)
		}

		// TODO: バリデーション
		// TODO: パスワードのハッシュ化

		res, err := db.NewInsert().
			Model(&user).
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
