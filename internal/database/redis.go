package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fasmail/panel/internal/config"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SessionData struct {
	UserID    uuid.UUID  `json:"user_id"`
	Role      string     `json:"role"`
	CompanyID *uuid.UUID `json:"company_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func PingRedis(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return fmt.Errorf("redis client is nil")
	}
	return client.Ping(ctx).Err()
}

func StoreSession(ctx context.Context, client *redis.Client, sessionID string, userID uuid.UUID, role string, companyID *uuid.UUID, ttl time.Duration) error {
	data := SessionData{
		UserID:    userID,
		Role:      role,
		CompanyID: companyID,
		CreatedAt: time.Now(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s", sessionID)
	if err := client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("store session: %w", err)
	}

	userKey := fmt.Sprintf("user_sessions:%s", userID.String())
	if err := client.SAdd(ctx, userKey, sessionID).Err(); err != nil {
		return fmt.Errorf("add to user sessions: %w", err)
	}

	return nil
}

func GetSession(ctx context.Context, client *redis.Client, sessionID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &session, nil
}

func DeleteSession(ctx context.Context, client *redis.Client, sessionID string, userID uuid.UUID) error {
	key := fmt.Sprintf("session:%s", sessionID)
	if err := client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	userKey := fmt.Sprintf("user_sessions:%s", userID.String())
	client.SRem(ctx, userKey, sessionID)

	return nil
}

func DeleteAllUserSessions(ctx context.Context, client *redis.Client, userID uuid.UUID) error {
	userKey := fmt.Sprintf("user_sessions:%s", userID.String())

	sessionIDs, err := client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("get user sessions: %w", err)
	}

	for _, sid := range sessionIDs {
		key := fmt.Sprintf("session:%s", sid)
		client.Del(ctx, key)
	}

	client.Del(ctx, userKey)

	return nil
}
