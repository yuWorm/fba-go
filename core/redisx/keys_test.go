package redisx_test

import (
	"testing"

	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/redisx"
)

func TestKeysUseCompatibleDefaultPrefix(t *testing.T) {
	keys := redisx.NewKeys("")

	cases := map[string]string{
		"access":           keys.AccessToken(10001, "session-1"),
		"online":           keys.OnlineSet(),
		"online_sid":       keys.OnlineSID("sid-1"),
		"online_session":   keys.OnlineSession("session-1"),
		"realtime_channel": keys.RealtimeBroadcastChannel(),
		"refresh":          keys.RefreshToken(10001, "session-1"),
		"user":             keys.User(10001),
		"captcha":          keys.LoginCaptcha("uuid-1"),
		"dict_cache":       keys.DictCache(),
		"scheduler_leader": keys.SchedulerLeader(),
		"migration_lock":   keys.MigrationLock(),
	}

	want := map[string]string{
		"access":           "fba:token:10001:session-1",
		"online":           "fba:token_online",
		"online_sid":       "fba:token_online:sid:sid-1",
		"online_session":   "fba:token_online:session:session-1",
		"realtime_channel": "fba:realtime:broadcast",
		"refresh":          "fba:refresh_token:10001:session-1",
		"user":             "fba:user:10001",
		"captcha":          "fba:login:captcha:uuid-1",
		"dict_cache":       "fba:cache:dict",
		"scheduler_leader": "fba:task:scheduler:leader",
		"migration_lock":   "fba:migration:lock",
	}

	for name, got := range cases {
		if got != want[name] {
			t.Fatalf("%s key = %q, want %q", name, got, want[name])
		}
	}
}

func TestUniversalOptionsMapsSentinel(t *testing.T) {
	opts := redisx.UniversalOptions(config.RedisOptions{
		Mode:       "sentinel",
		Addrs:      []string{"10.0.0.1:26379", "10.0.0.2:26379"},
		MasterName: "mymaster",
		Password:   "secret",
		DB:         2,
		PoolSize:   16,
	})

	if opts.MasterName != "mymaster" {
		t.Fatalf("MasterName = %q, want mymaster", opts.MasterName)
	}
	if len(opts.Addrs) != 2 {
		t.Fatalf("Addrs length = %d, want 2", len(opts.Addrs))
	}
	if opts.DB != 2 {
		t.Fatalf("DB = %d, want 2", opts.DB)
	}
	if opts.PoolSize != 16 {
		t.Fatalf("PoolSize = %d, want 16", opts.PoolSize)
	}
}
