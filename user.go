package main

import (
	"context"
	"net/http"
	"time"

	"github.com/daikideal/go-passkey-demo/db"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"

	googleUuid "github.com/google/uuid"
)

var usersInMemory []*User

type uuid = string

type User struct {
	ID        uuid      `json:"id" bun:"id"`
	Name      string    `json:"name" bun:"name"`
	Email     string    `json:"email" bun:"email"`
	Password  string    `json:"password" bun:"password"`
	CreatedAt time.Time `json:"created_at" bun:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at"`

	Credentials []webauthn.Credential
}

// ユーザーには表示しないが、WebAuthnでユーザーを識別するために使用するID。
// 64バイトのランダムなバイト列である必要がある。
// 専用のカラムを追加するべきなのかもしれないが、一旦は User.ID(UUID) をエンコーディングしたものを使用する
// 今回、UUIDはpostgresが生成するものであるため、エラーハンドリングは考えない。
//
// https://w3c.github.io/webauthn/#dom-publickeycredentialuserentity-id
func (user *User) WebAuthnID() []byte {
	parsedUUID, _ := googleUuid.Parse(user.ID)
	uuidBytes := parsedUUID[:]

	return uuidBytes
}

func (user *User) WebAuthnName() string {
	return user.Name
}

// webauthnの想定的にはNameとdisplayNameは異なるらしいが、一旦同じものを返す
func (user *User) WebAuthnDisplayName() string {
	return user.Name
}

func (user *User) WebAuthnCredentials() []webauthn.Credential {
	return user.Credentials
}

// 非推奨らしいので空文字を返す
func (user *User) WebAuthnIcon() string {
	return ""
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

func findUserByID(ctx context.Context, id uuid) (*User, error) {
	db := db.GetDB()

	var user User
	err := db.NewSelect().
		Model(&user).
		Column("*").
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func beginRegistration(w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		user, err := findUserByID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(404, nil)
		}

		options, session, err := w.BeginRegistration(user)
		if err != nil {
			ctx.Logger().Errorf("Failed to begin registration: %v\n", err)
			return ctx.JSON(500, nil)
		}

		// 認証機登録セッションを開始
		// cookieを使用してはいけない場合、レスポンスで返してやるのがよいか。
		sessionId, err := CreateSession(ctx.Request().Context(), session)
		if err != nil {
			ctx.Logger().Errorf("Failed to start session: %v\n", err)
			return ctx.JSON(500, nil)
		}
		ctx.SetCookie(&http.Cookie{
			Name:  "registration",
			Value: sessionId,
			Path:  "/",
		})

		return ctx.JSON(200, options)
	}
}

func finishRegistration(w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		user, err := findUserByID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			ctx.Logger().Errorf("User is not found: %v\n", err)
			return ctx.JSON(404, nil)
		}

		// 認証機登録セッションを特定
		cookie, err := ctx.Cookie("registration")
		if err != nil {
			ctx.Logger().Errorf("Cookie is not set: %v\n", err)
			return ctx.JSON(400, nil)
		}
		session, err := GetSession(ctx.Request().Context(), cookie.Value)
		if err != nil {
			ctx.Logger().Errorf("Session is not found: %v\n", err)
			return ctx.JSON(400, nil)
		}

		res := ctx.Request()
		credential, err := w.FinishRegistration(user, *session, res)
		if err != nil {
			ctx.Logger().Errorf("Failed to finish registration: %v\n", err)
			return ctx.JSON(500, nil)
		}

		// 何を保存すればいいかわからない。。
		user.Credentials = append(user.Credentials, *credential)
		usersInMemory = append(usersInMemory, user)
		ctx.Logger().Infof("Users: %+v\n", usersInMemory)

		DeleteSession(ctx.Request().Context(), cookie.Value)

		return ctx.JSON(201, "Registration success!")
	}
}
