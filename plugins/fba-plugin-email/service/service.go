package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultEmailCaptchaTTL    = 180 * time.Second
	defaultEmailCaptchaPrefix = "fba:email:captcha"
)

type RedisClient interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

type CaptchaSender interface {
	SendCaptcha(ctx context.Context, recipients []string, code string, expireMinutes int) error
}

type NoopCaptchaSender struct{}

func (NoopCaptchaSender) SendCaptcha(context.Context, []string, string, int) error {
	return nil
}

type Options struct {
	Redis         RedisClient
	Sender        CaptchaSender
	CaptchaTTL    time.Duration
	CaptchaPrefix string
}

type Service struct {
	redis  RedisClient
	sender CaptchaSender
	ttl    time.Duration
	prefix string
}

func New(opts Options) *Service {
	if opts.Sender == nil {
		opts.Sender = NoopCaptchaSender{}
	}
	if opts.CaptchaTTL <= 0 {
		opts.CaptchaTTL = defaultEmailCaptchaTTL
	}
	if opts.CaptchaPrefix == "" {
		opts.CaptchaPrefix = defaultEmailCaptchaPrefix
	}
	return &Service{
		redis:  opts.Redis,
		sender: opts.Sender,
		ttl:    opts.CaptchaTTL,
		prefix: strings.TrimRight(opts.CaptchaPrefix, ":"),
	}
}

func (s *Service) SendCaptcha(ctx context.Context, recipients []string, ip string) error {
	code, err := digitCode(6)
	if err != nil {
		return err
	}
	if s.redis != nil {
		if err := s.redis.Set(ctx, s.captchaKey(ip), code, s.ttl).Err(); err != nil {
			return err
		}
	}
	return s.sender.SendCaptcha(ctx, recipients, code, int(s.ttl.Minutes()))
}

func (s *Service) captchaKey(ip string) string {
	return s.prefix + ":" + ip
}

func digitCode(length int) (string, error) {
	var builder strings.Builder
	builder.Grow(length)
	for i := 0; i < length; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(9))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('1' + value.Int64()))
	}
	return builder.String(), nil
}
