package auth

import (
	"context"
	"fmt"
	"log"
	"time"
	"workout-tracker-api/internal/cache"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const cachePrefix = "jwtblacklist:"

type Payload struct {
	Id    *int   `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type Claims struct {
	Payload
	jwt.RegisteredClaims
}

type TokenInterface interface {
	GenerateToken(claims Claims) (string, error)
	ParseToken(ctx context.Context, tokenString string) (*Claims, error)
	BlacklistToken(ctx context.Context, jti string, expirationTIme time.Time) error
	CheckBlacklist(ctx context.Context, jti string) (bool, error)
}

type JWTService struct {
	signingMethod jwt.SigningMethod
	cache         cache.CacheInterface
	secretKey     string
}

func NewJWTService(sm jwt.SigningMethod, cache cache.CacheInterface, secretKey string) TokenInterface {
	return &JWTService{
		signingMethod: sm,
		cache:         cache,
		secretKey:     secretKey,
	}
}

func (js *JWTService) GenerateToken(claims Claims) (string, error) {

	expirationTime := time.Now().UTC().Add(24 * time.Hour)
	issuedAtTIme := time.Now().UTC()
	tokenID := uuid.New().String()

	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	claims.IssuedAt = jwt.NewNumericDate(issuedAtTIme)
	claims.ID = tokenID

	token := jwt.NewWithClaims(js.signingMethod, claims)

	signedToken, err := token.SignedString([]byte(js.secretKey))

	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil

}

// ParseToken
func (js *JWTService) ParseToken(ctx context.Context, tokenString string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != js.signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(js.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token or claims could not be extracted")
	}

}

func (js *JWTService) BlacklistToken(ctx context.Context, jti string, expirationTime time.Time) error {

	key := cachePrefix + jti
	duration := time.Until(expirationTime)
	err := js.cache.SaveCache(ctx, key, jti, &duration)
	if err != nil {
		return fmt.Errorf("error saving cache: %w", err)
	}

	log.Printf("Token JTI '%s' blacklisted until %s\n", jti, expirationTime.String())
	return nil
}

func (js *JWTService) CheckBlacklist(ctx context.Context, jti string) (bool, error) {

	key := cachePrefix + jti
	found, err := js.cache.ExistCache(ctx, key)
	if err != nil {
		return false, fmt.Errorf("error checking cache: %w", err)
	}
	if !found {
		return false, nil
	}

	return true, nil
}
