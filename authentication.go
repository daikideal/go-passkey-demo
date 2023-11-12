package main

import (
	"fmt"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
)

func beginLogin(w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		options, session, err := w.BeginDiscoverableLogin()
		if err != nil {
			ctx.Logger().Errorf("Failed to begin login: %v\n", err)
			return ctx.JSON(http.StatusInternalServerError, "Failed to begin login")
		}

		sessionId, err := CreateSession(ctx.Request().Context(), session)
		if err != nil {
			ctx.Logger().Errorf("Failed to start session: %v\n", err)
			return ctx.JSON(http.StatusInternalServerError, nil)
		}
		ctx.SetCookie(&http.Cookie{
			Name:  "authentication",
			Value: sessionId,
			Path:  "/",
		})

		return ctx.JSON(http.StatusOK, options)
	}
}

func finishLogin(w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		cookie, err := ctx.Cookie("authentication")
		if err != nil {
			ctx.Logger().Errorf("Cookie is not set: %v\n", err)
			return ctx.JSON(http.StatusBadRequest, nil)
		}
		session, err := GetSession(ctx.Request().Context(), cookie.Value)

		res, err := protocol.ParseCredentialRequestResponse(ctx.Request())
		if err != nil {
			ctx.Logger().Errorf("Failed to parse credential request response: %v\n", err)
			return ctx.JSON(http.StatusBadRequest, nil)
		}
		credential, err := w.ValidateDiscoverableLogin(func(rawID, userHandle []byte) (webauthn.User, error) {
			// TODO: rawID or userHandle を使ってユーザーを特定する
			// 		 userHandle の方は、webauthn.User インターフェースが返すアレのことだと思われる。
			// 		 場合によっては、User.WebAuthnID() の実装を変更する必要がある。
			return nil, nil
		}, *session, res)
		if err != nil {
			ctx.Logger().Errorf("Failed to validate discoverable login: %v\n", err)
			return ctx.JSON(http.StatusBadRequest, nil)
		}
		fmt.Printf("credential: %+v\n", credential)

		// 未実装
		return ctx.JSON(http.StatusOK, "Login Success")
	}
}
