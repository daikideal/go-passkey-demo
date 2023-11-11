package main

import (
	"context"
	"net/http"
	"time"

	"github.com/daikideal/go-passkey-demo/db"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"

	googleUuid "github.com/google/uuid"
)

type uuid = string

// この構造体を User に紐づけて保存するための構造体
// https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.8.6/webauthn#Credential
type WebauthnCredentials struct {
	ID              uuid                              `json:"id" bun:"id,pk"`
	UserID          uuid                              `json:"user_id" bun:"user_id"`
	CredentialID    []byte                            `json:"credential_id" bun:"credential_id"`
	PublicKey       []byte                            `json:"public_key" bun:"public_key"`
	AttestationType string                            `json:"attestation_type" bun:"attestation_type"`
	Transport       []protocol.AuthenticatorTransport `json:"transport" bun:"transport,array"`
	Flags           webauthn.CredentialFlags          `json:"flags" bun:"flags"`
	Authenticator   webauthn.Authenticator            `json:"authenticator" bun:"authenticator"`
}

type User struct {
	ID                  uuid                  `json:"id" bun:"id,pk"`
	Name                string                `json:"name" bun:"name"`
	Email               string                `json:"email" bun:"email"`
	Password            string                `json:"password" bun:"password"`
	WebauthnCredentials []WebauthnCredentials `json:"webauthn_credentials" bun:"rel:has-many,join:id=user_id"`
	CreatedAt           time.Time             `json:"created_at" bun:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at" bun:"updated_at"`
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
	res := make([]webauthn.Credential, len(user.WebauthnCredentials))

	for i, v := range user.WebauthnCredentials {
		res[i] = webauthn.Credential{
			ID:              v.CredentialID,
			PublicKey:       v.PublicKey,
			AttestationType: v.AttestationType,
			Transport:       v.Transport,
			Flags:           v.Flags,
			Authenticator:   v.Authenticator,
		}
	}

	return res
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
			Relation("WebauthnCredentials").
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
		Relation("WebauthnCredentials").
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

		newWebautnCredential := &WebauthnCredentials{
			UserID:          user.ID,
			CredentialID:    credential.ID,
			PublicKey:       credential.PublicKey,
			AttestationType: credential.AttestationType,
			Transport:       credential.Transport,
			Flags:           credential.Flags,
			Authenticator:   credential.Authenticator,
		}

		db := db.GetDB()
		_, err = db.NewInsert().
			Model(newWebautnCredential).
			Column("user_id", "credential_id", "public_key", "attestation_type", "transport", "flags", "authenticator").
			Exec(ctx.Request().Context())
		if err != nil {
			ctx.Logger().Errorf("Failed to insert webauthn credential: %v\n", err)
			return ctx.JSON(500, nil)
		}

		return ctx.JSON(201, "Registration success!")
	}
}
