package config_test

import (
	"testing"

	"github.com/yuWorm/fba-go/core/config"
)

func TestRealtimeMultiInstanceDefaultsUseRedisKeyPrefix(t *testing.T) {
	opts := config.Options{
		Redis: config.RedisOptions{KeyPrefix: "acme"},
		Realtime: config.RealtimeOptions{
			MultiInstance: config.RealtimeMultiInstanceOptions{Enabled: true},
		},
	}.WithDefaults()

	if opts.Realtime.MultiInstance.Channel != "acme:realtime:broadcast" {
		t.Fatalf("Channel = %q, want acme:realtime:broadcast", opts.Realtime.MultiInstance.Channel)
	}
	if opts.Realtime.MultiInstance.NodeID != "" {
		t.Fatalf("NodeID default = %q, want empty so app can derive host/process id", opts.Realtime.MultiInstance.NodeID)
	}
}
