package realtime_test

import (
	"testing"

	"github.com/yuWorm/fba-go/core/realtime"
)

func TestMemoryOnlineStoreTracksSessionUntilLastSIDDisconnects(t *testing.T) {
	store := realtime.NewMemoryOnlineStore()

	store.Connect("sid-1", "session-1")
	store.Connect("sid-2", "session-1")

	if sessions := store.Sessions(); len(sessions) != 1 || sessions[0] != "session-1" {
		t.Fatalf("Sessions() = %v, want [session-1]", sessions)
	}
	if got := store.SessionForSID("sid-1"); got != "session-1" {
		t.Fatalf("SessionForSID() = %q, want session-1", got)
	}

	store.Disconnect("sid-1")
	if sessions := store.Sessions(); len(sessions) != 1 || sessions[0] != "session-1" {
		t.Fatalf("Sessions() after first disconnect = %v, want [session-1]", sessions)
	}

	store.Disconnect("sid-2")
	if sessions := store.Sessions(); len(sessions) != 0 {
		t.Fatalf("Sessions() after final disconnect = %v, want empty", sessions)
	}
}

func TestMemoryHubRecordsTaskNotificationPayload(t *testing.T) {
	hub := realtime.NewMemoryHub(realtime.NewMemoryOnlineStore())

	if err := hub.Emit("task_notification", realtime.TaskNotification{Msg: "任务 task_demo 执行成功"}); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	events := hub.Events()
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Event != "task_notification" {
		t.Fatalf("event = %q, want task_notification", events[0].Event)
	}
	payload, ok := events[0].Data.(realtime.TaskNotification)
	if !ok {
		t.Fatalf("payload type = %T, want realtime.TaskNotification", events[0].Data)
	}
	if payload.Msg != "任务 task_demo 执行成功" {
		t.Fatalf("payload msg = %q", payload.Msg)
	}
}
