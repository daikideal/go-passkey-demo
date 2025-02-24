package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/redis/go-redis/v9"
)

const duration time.Duration = 5 * time.Minute

var sessionStore *redis.Client

func init() {
	sessionStore = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
}

func CreateSession(ctx context.Context, data *webauthn.SessionData) (string, error) {
	// REVEIW: user/:id 配下に作成した方がいいか？
	sessionId, _ := random(32)

	// redisに直接structを保存することはできない。
	// 試したところ、byte列にすれば保存できた。
	value, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Failed to encoding data to redis value: %w", err)
	}

	if err := sessionStore.Set(ctx, sessionId, value, duration).Err(); err != nil {
		return "", fmt.Errorf("Failed to create session: %w", err)
	}

	return sessionId, nil
}

func GetSession(ctx context.Context, sessionID string) (*webauthn.SessionData, error) {
	val, err := sessionStore.Get(ctx, sessionID).Bytes()
	if err != nil {
		return nil, fmt.Errorf("Failed to get session: %w", err)
	}

	var session *webauthn.SessionData
	if err = json.Unmarshal(val, &session); err != nil {
		return nil, fmt.Errorf("Failed to decode session: %w", err)
	}

	return session, nil
}

func DeleteSession(ctx context.Context, sessionID string) {
	sessionStore.Del(ctx, sessionID)
}

func random(length int) (string, error) {
	randomData := make([]byte, length)
	_, err := rand.Read(randomData)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomData), nil
}
