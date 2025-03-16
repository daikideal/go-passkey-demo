package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/daikideal/go-passkey-demo/db"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/labstack/echo/v4"
)

type uuid = string

// この構造体を User に紐づけて保存するための構造体
// https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.8.6/webauthn#Credential
//
// 2025/02/28 追記:
// W3Cの仕様的にはCredential Recordを保存することが推奨されている。今定義しているものと合っているのかまだ確認していない。
// https://www.w3.org/TR/webauthn-3/#credential-record
type WebauthnCredentials struct {
	ID              uuid                              `json:"id" bun:"id,pk"`
	UserID          uuid                              `json:"user_id" bun:"user_id"`
	CredentialID    []byte                            `json:"credential_id" bun:"credential_id"`
	PublicKey       []byte                            `json:"public_key" bun:"public_key"`
	AttestationType string                            `json:"attestation_type" bun:"attestation_type"`
	Transport       []protocol.AuthenticatorTransport `json:"transport" bun:"transport,array"`
	Flags           webauthn.CredentialFlags          `json:"flags" bun:"flags"`
	Authenticator   webauthn.Authenticator            `json:"authenticator" bun:"authenticator"`
	CreatedAt       time.Time                         `json:"created_at" bun:"created_at"`
	UpdatedAt       time.Time                         `json:"updated_at" bun:"updated_at"`
}

type User struct {
	ID                  uuid                  `json:"id" bun:"id,pk"`
	Name                string                `json:"name" bun:"name"`
	WebauthnCredentials []WebauthnCredentials `json:"webauthn_credentials" bun:"rel:has-many,join:id=user_id"`
	CreatedAt           time.Time             `json:"created_at" bun:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at" bun:"updated_at"`
}

// ユーザーには表示しないが、WebAuthnでユーザーを識別するために使用するID。
// UUIDをバイト列に変換すると、結果のサイズは16バイトになるので、これを使用する。
//
// https://w3c.github.io/webauthn/#dom-publickeycredentialuserentity-id
func (user *User) WebAuthnID() []byte {
	return []byte(user.ID)
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

// このユーザーがすでに登録している認証器の情報を生成するためのメソッド。
// webauthn.BeginRegistration に、同じ認証器が複数登録されるのを防ぐオプションを設定するために使用する。
func (user *User) CredentialExcludeList() []protocol.CredentialDescriptor {

	credentialExcludeList := []protocol.CredentialDescriptor{}
	for _, cred := range user.WebauthnCredentials {
		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.CredentialID, // protocol.URLEncodedBase64 は []byte のアノテーションなのでこれで問題ない
		}
		credentialExcludeList = append(credentialExcludeList, descriptor)
	}

	return credentialExcludeList
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

func getUser() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Logger().Info("GET /user/:id")

		userID := ctx.Param("id")

		user, err := findUserByID(ctx.Request().Context(), userID)
		if err != nil {
			ctx.Logger().Errorf("Failed to find user: %v\n", err)
			return ctx.JSON(404, nil)
		}

		return ctx.JSON(http.StatusOK, user)
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

func findUserByName(ctx context.Context, name string) (*User, error) {
	db := db.GetDB()

	var user User
	err := db.NewSelect().
		Model(&user).
		Relation("WebauthnCredentials").
		Column("*").
		Where("name = ?", name).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func createNewUser(ctx context.Context, name string) (*User, error) {
	db := db.GetDB()

	user := &User{
		Name: name,
	}
	_, err := db.NewInsert().
		Model(user).
		Column("name").
		Returning("*").
		Exec(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

type beginRegistrationReqest struct {
	Username string `json:"username"`
}

func beginRegistration(w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var req beginRegistrationReqest
		if err := ctx.Bind(&req); err != nil {
			ctx.Logger().Errorf("Failed to process request: %v\n", err)
			return ctx.JSON(400, nil)
		}

		// 認証機を登録するユーザーを特定。存在しない場合は新規作成する。
		user, err := findUserByName(context.Background(), req.Username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				user, err = createNewUser(context.Background(), req.Username)
				if err != nil {
					ctx.Logger().Errorf("Failed to create user: %v\n", err)
					return ctx.JSON(500, nil)
				}
			} else {
				ctx.Logger().Errorf("Failed to find user: %v\n", err)
				return ctx.JSON(400, nil)
			}
		}

		options, session, err := w.BeginRegistration(
			user,
			// パスキー認証が試したいので、 Resident Key しかサポートしなくする。
			webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
			webauthn.WithExclusions(user.CredentialExcludeList()),
		)
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
			Name:     "registration",
			Value:    sessionId,
			Path:     "/",
			HttpOnly: true,
		})

		return ctx.JSON(200, options)
	}
}

type finishRegistrationReqest struct {
	protocol.CredentialCreationResponse
}

func finishRegistration(w *webauthn.WebAuthn) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		body, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			ctx.Logger().Errorf("Failed to read request: %v\n", err)
			return ctx.JSON(500, nil)
		}

		fmt.Printf("Request body: %s\n", body)

		req := &finishRegistrationReqest{}
		err = json.Unmarshal(body, req)
		if err != nil {
			ctx.Logger().Errorf("Failed to parse request: %v\n", err)
			return ctx.JSON(400, nil)
		}

		// リクエストボディを元に戻す。そうしないと、FinishRegistration に ctx.Request() を渡した後、
		// ParseCredentialCreationResponseBody でエラーになる。
		// ref. https://syossan.hateblo.jp/entry/2019/01/11/175932
		ctx.Request().Body = io.NopCloser(bytes.NewBuffer(body))

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

		// セッションからユーザーを特定
		user, err := findUserByID(ctx.Request().Context(), string(session.UserID))
		if err != nil {
			ctx.Logger().Errorf("User is not found: %v\n", err)
			return ctx.JSON(404, nil)
		}

		res := ctx.Request()
		credential, err := w.FinishRegistration(user, *session, res)
		if err != nil {
			ctx.Logger().Errorf("Failed to finish registration: %v\n", err)
			return ctx.JSON(500, nil)
		}

		fmt.Printf("AAGUID: %s\n", parseAaguidAsUuid(credential.Authenticator.AAGUID))

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

// `webauthn.Authenticator.AAGUID` をUUIDフォーマットに変換する。
//
// SEE:
//   - https://github.com/passkeydeveloper/passkey-authenticator-aaguids/blob/main/aaguid.json
//   - https://github.com/web-auth/webauthn-framework/pull/49
func parseAaguidAsUuid(aaguid []byte) string {
	hexString := hex.EncodeToString(aaguid)

	return fmt.Sprintf("%s-%s-%s-%s-%s", hexString[0:8], hexString[8:12], hexString[12:16], hexString[16:20], hexString[20:])
}

type listPublicKeysByUserResponse struct {
	ID              uuid                              `json:"id"`
	CredentialID    []byte                            `json:"credential_id"`
	PublicKey       []byte                            `json:"public_key"`
	AttestationType string                            `json:"attestation_type"`
	Transport       []protocol.AuthenticatorTransport `json:"transport"`
	Flags           webauthn.CredentialFlags          `json:"flags"`
	Authenticator   webauthn.Authenticator            `json:"authenticator"`
}

func listPublicKeysByUser() echo.HandlerFunc {
	db := db.GetDB()

	return func(ctx echo.Context) error {
		userID := ctx.Param("id")

		var credentials []*WebauthnCredentials
		if err := db.NewSelect().
			Model(&credentials).
			Column("*").
			Where("user_id = ?", userID).
			Scan(ctx.Request().Context()); err != nil {
			ctx.Logger().Errorf("Failed to select webauthn credentials: %v\n", err)
			return ctx.JSON(500, nil)
		}

		res := make([]listPublicKeysByUserResponse, len(credentials))
		for i, v := range credentials {
			res[i] = listPublicKeysByUserResponse{
				ID:              v.ID,
				CredentialID:    v.CredentialID,
				PublicKey:       v.PublicKey,
				AttestationType: v.AttestationType,
				Transport:       v.Transport,
				Flags:           v.Flags,
				Authenticator:   v.Authenticator,
			}
		}

		return ctx.JSON(http.StatusOK, res)
	}
}

func deletePublicKey() echo.HandlerFunc {
	db := db.GetDB()

	return func(ctx echo.Context) error {
		userID := ctx.Param("user_id")
		publicKeyID := ctx.Param("public_key_id")

		_, err := db.NewDelete().
			Model(&WebauthnCredentials{}).
			Where("user_id = ? AND id = ?", userID, publicKeyID).
			Exec(ctx.Request().Context())
		if err != nil {
			ctx.Logger().Errorf("Failed to delete webauthn credential: %v\n", err)
			return ctx.JSON(http.StatusInternalServerError, nil)
		}

		return ctx.JSON(http.StatusNoContent, nil)
	}
}
