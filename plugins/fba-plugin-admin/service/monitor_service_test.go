package service_test

import (
	"context"
	"testing"

	"github.com/yuWorm/fba-go/core/realtime"
	"github.com/yuWorm/fba-plugin-admin/repo"
	"github.com/yuWorm/fba-plugin-admin/service"
)

func TestMonitorServiceMarksRealtimeOnlineSessions(t *testing.T) {
	online := realtime.NewMemoryOnlineStore()
	online.Connect("sid-1", "fixture-session")
	svc := service.NewMonitorServiceWithRealtime(repo.NewMemoryRepository(repo.SeedData()), nil, online)

	sessions, err := svc.Sessions(context.Background(), "admin")
	if err != nil {
		t.Fatalf("Sessions() error = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("sessions = %d, want 1", len(sessions))
	}
	if sessions[0].Status != 1 {
		t.Fatalf("Status = %d, want 1 for realtime online session", sessions[0].Status)
	}

	online.Disconnect("sid-1")
	sessions, err = svc.Sessions(context.Background(), "admin")
	if err != nil {
		t.Fatalf("Sessions() after disconnect error = %v", err)
	}
	if sessions[0].Status != 0 {
		t.Fatalf("Status after disconnect = %d, want 0", sessions[0].Status)
	}
}
