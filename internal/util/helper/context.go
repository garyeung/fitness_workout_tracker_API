package helper

import (
	"context"
	"time"
)

type ContextKey string

type UserInfo struct {
	Id    int
	Email string
	Name  string
}

type JTIInfo struct {
	Id             string
	ExpirationTime time.Time
}

const (
	UserContextKey   ContextKey = "user"
	JTIContextKey    ContextKey = "jti"
	ClaimsContextKey ContextKey = "claims"
)

func GetUserInfoFromContext(ctx context.Context) (*UserInfo, bool) {

	return GetFromContext[*UserInfo](ctx, UserContextKey)
}
func GetJTIFromContext(ctx context.Context) (*JTIInfo, bool) {
	return GetFromContext[*JTIInfo](ctx, UserContextKey)
}

func SetUserInfoToContext(ctx context.Context, user *UserInfo) context.Context {
	return SetToContext(ctx, UserContextKey, user)
}

func SetJTIToContext(ctx context.Context, jti *JTIInfo) context.Context {
	return SetToContext(ctx, JTIContextKey, jti)
}

func GetFromContext[T any](ctx context.Context, key ContextKey) (T, bool) {
	t, ok := ctx.Value(key).(T)
	return t, ok
}

func SetToContext[T any](ctx context.Context, key ContextKey, value T) context.Context {
	return context.WithValue(ctx, key, value)
}
