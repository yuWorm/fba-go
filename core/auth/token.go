package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yuWorm/fba-go/core/config"
)

const defaultAccessTokenTTL = 2 * time.Hour

type TokenService interface {
	CreateAccessToken(ctx context.Context, userID int64, sessionUUID string, extra map[string]any) (*AccessToken, error)
	ParseAccessToken(token string) (*Claims, error)
}

type AccessToken struct {
	Token       string
	SessionUUID string
	ExpiresAt   time.Time
}

type Claims struct {
	SessionUUID string `json:"session_uuid"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
	issuer string
	ttl    time.Duration
	Now    func() time.Time
}

func NewJWTService(opts config.AuthOptions) *JWTService {
	ttl := opts.AccessTokenTTL
	if ttl == 0 {
		ttl = defaultAccessTokenTTL
	}
	secret := opts.JWTSecret
	if secret == "" {
		secret = "change-me"
	}
	return &JWTService{
		secret: []byte(secret),
		issuer: opts.JWTIssuer,
		ttl:    ttl,
		Now:    time.Now,
	}
}

func (s *JWTService) CreateAccessToken(_ context.Context, userID int64, sessionUUID string, _ map[string]any) (*AccessToken, error) {
	if sessionUUID == "" {
		sessionUUID = uuid.NewString()
	}
	expiresAt := s.Now().Add(s.ttl)
	claims := Claims{
		SessionUUID: sessionUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    s.issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return nil, err
	}
	return &AccessToken{
		Token:       token,
		SessionUUID: sessionUUID,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *JWTService) ParseAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
