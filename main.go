package main

import (
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/", hello())

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
