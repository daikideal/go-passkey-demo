package main

import (
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/", hello())
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

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func getUser() echo.HandlerFunc {
	res := []User{
		{
			ID:       1,
			Name:     "test",
			Password: "test",
		},
		{
			ID:       2,
			Name:     "test_2",
			Password: "test_2",
		},
	}

	return func(ctx echo.Context) error {
		return ctx.JSON(200, res)
	}
}
