package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yuWorm/fba-go/core/config"
)

const defaultAccessTokenTTL = 2 * time.Hour

var (
	ErrAccessTokenExpired = errors.New("access token expired")
	ErrJWTSecretRequired  = errors.New("JWT secret must contain at least 32 bytes")
	ErrJWTIssuerRequired  = errors.New("JWT issuer is required")
)

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
	configErr error
}

func NewJWTService(opts config.AuthOptions) *JWTService {
	ttl := opts.AccessTokenTTL
	if ttl <= 0 {
		ttl = defaultAccessTokenTTL
	}
	return &JWTService{
		secret:    []byte(opts.JWTSecret),
		issuer:    strings.TrimSpace(opts.JWTIssuer),
		ttl:       ttl,
		Now:       time.Now,
		configErr: ValidateJWTOptions(opts),
	}
}

func ValidateJWTOptions(opts config.AuthOptions) error {
	if len([]byte(opts.JWTSecret)) < 32 {
		return ErrJWTSecretRequired
	}
	if strings.TrimSpace(opts.JWTIssuer) == "" {
		return ErrJWTIssuerRequired
	}
	return nil
}

func (s *JWTService) CreateAccessToken(_ context.Context, userID int64, sessionUUID string, _ map[string]any) (*AccessToken, error) {
	if s.configErr != nil {
		return nil, s.configErr
	}
	if sessionUUID == "" {
		sessionUUID = uuid.NewString()
	}
	now := s.Now()
	expiresAt := now.Add(s.ttl)
	claims := Claims{
		SessionUUID: sessionUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    s.issuer,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
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
	if s.configErr != nil {
		return nil, s.configErr
	}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
			}
			return s.secret, nil
		},
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrAccessTokenExpired
		}
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
