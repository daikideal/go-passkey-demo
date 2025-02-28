package main

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

var (
	webAuthn *webauthn.WebAuthn
	err      error
)

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true, // Cookieを取り扱えるようにする
	}))

	wconfig := &webauthn.Config{
		RPDisplayName: "go-passkey-demo",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:5173"},
	}

	webAuthn, err = webauthn.New(wconfig)
	if err != nil {
		panic(err)
	}

	e.POST("/users", createUser())
	e.GET("/users", getUsers())
	e.GET("/users/:id", getUser())
	// パスキー管理
	e.GET("/users/:id/public_keys", listPublicKeysByUser())
	// 認証機の登録
	e.POST("/registration/options", beginRegistration(webAuthn))
	e.POST("/registration/verifications", finishRegistration(webAuthn))
	// 認証
	e.POST("/authentication/options", beginLogin(webAuthn))
	e.POST("/authentication/verifications", finishLogin(webAuthn))

	e.Logger.Fatal(e.Start(":8080"))
}
