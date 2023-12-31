package main

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

var (
	webAuthn *webauthn.WebAuthn
	err      error
)

func main() {
	e := echo.New()
	// passkey認証ができるフォームがほしいだけなので、echoの静的ファイルハンドラを使用する。
	e.Static("/", "web")

	wconfig := &webauthn.Config{
		RPDisplayName: "go-passkey-demo",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	}
	webAuthn, err = webauthn.New(wconfig)

	e.GET("/", index())
	e.POST("/users", createUser())
	e.GET("/users", getUsers())
	// 認証機の登録
	e.POST("/registration/options", beginRegistration(webAuthn))
	e.POST("/registration/verifications", finishRegistration(webAuthn))
	// 認証
	e.POST("/authentication/options", beginLogin(webAuthn))
	e.POST("/authentication/verifications", finishLogin(webAuthn))

	e.Logger.Fatal(e.Start(":8080"))
}

func index() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return ctx.File("./web/index.html")
	}
}
