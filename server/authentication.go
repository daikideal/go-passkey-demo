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

type finishLoginResponse struct {
	UserID string `json:"user_id"`
}

func finishLogin(w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		cookie, err := ctx.Cookie("authentication")
		if err != nil {
			ctx.Logger().Errorf("Cookie is not set: %v\n", err)
			return ctx.JSON(http.StatusBadRequest, nil)
		}
		session, err := GetSession(ctx.Request().Context(), cookie.Value)

		var userID string
		// ValidateDiscoverableLogin にて、どのようにログインするユーザーを特定するかを定義する関数。
		//
		// userHandle は User インターフェース実装されている WebAuthnId() のこと。
		// 今回はプライマリIDであるUUIDをバイト列に変換したものを返しているので、 userHandle を string に変換して User をクエリすればユーザーを特定できる。
		// rawID が何なのかわかっておらず、いまいちどうやって使えばいいかわからない。
		handler := func(rawID, userHandle []byte) (webauthn.User, error) {
			userID = string(userHandle)
			user, err := findUserByID(ctx.Request().Context(), userID)
			if err != nil {
				ctx.Logger().Errorf("Failed to find user: %v\n", err)
				return nil, fmt.Errorf("Failed to find user")
			}

			return user, nil
		}

		res, err := protocol.ParseCredentialRequestResponse(ctx.Request())
		if err != nil {
			ctx.Logger().Errorf("Failed to parse credential request response: %v\n", err)
			return ctx.JSON(http.StatusBadRequest, "Failed to parse credential request response")
		}

		if _, err = w.ValidateDiscoverableLogin(handler, *session, res); err != nil {
			ctx.Logger().Errorf("Failed to validate discoverable login: %v\n", err)
			return ctx.JSON(http.StatusBadRequest, "Failed to validate discoverable login")
		}

		return ctx.JSON(http.StatusOK, finishLoginResponse{UserID: userID})
	}
}
