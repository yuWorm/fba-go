package realtime

import (
	"encoding/json"
	"sort"
	"sync"
)

const (
	EventTaskNotification = "task_notification"
	EventTaskWorkerStatus = "task_worker_status"
)

type TaskNotification struct {
	Msg string `json:"msg"`
}

type Event struct {
	SocketID string
	Event    string
	Data     any
}

type EventPayload struct {
	SocketID string
	Data     []byte
	Args     [][]byte
	Ack      func(args ...any) error
}

type EventHandler func(EventPayload)

type Hub interface {
	Emit(event string, data any) error
	EmitTo(socketID string, event string, data any) error
	On(event string, handler EventHandler)
	OnlineStore() OnlineStore
}

type OnlineStore interface {
	Connect(sid string, sessionUUID string)
	Disconnect(sid string)
	SessionForSID(sid string) string
	Sessions() []string
	SIDs(sessionUUID string) []string
}

type MemoryOnlineStore struct {
	mu        sync.RWMutex
	sidToSess map[string]string
	sessToSID map[string]map[string]struct{}
}

func NewMemoryOnlineStore() *MemoryOnlineStore {
	return &MemoryOnlineStore{
		sidToSess: make(map[string]string),
		sessToSID: make(map[string]map[string]struct{}),
	}
}

func (s *MemoryOnlineStore) Connect(sid string, sessionUUID string) {
	if sid == "" || sessionUUID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sidToSess[sid] = sessionUUID
	if s.sessToSID[sessionUUID] == nil {
		s.sessToSID[sessionUUID] = make(map[string]struct{})
	}
	s.sessToSID[sessionUUID][sid] = struct{}{}
}

func (s *MemoryOnlineStore) Disconnect(sid string) {
	if sid == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionUUID := s.sidToSess[sid]
	if sessionUUID == "" {
		return
	}
	delete(s.sidToSess, sid)
	delete(s.sessToSID[sessionUUID], sid)
	if len(s.sessToSID[sessionUUID]) == 0 {
		delete(s.sessToSID, sessionUUID)
	}
}

func (s *MemoryOnlineStore) SessionForSID(sid string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sidToSess[sid]
}

func (s *MemoryOnlineStore) Sessions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]string, 0, len(s.sessToSID))
	for sessionUUID := range s.sessToSID {
		items = append(items, sessionUUID)
	}
	sort.Strings(items)
	return items
}

func (s *MemoryOnlineStore) SIDs(sessionUUID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sids := s.sessToSID[sessionUUID]
	items := make([]string, 0, len(sids))
	for sid := range sids {
		items = append(items, sid)
	}
	sort.Strings(items)
	return items
}

type MemoryHub struct {
	mu       sync.RWMutex
	store    OnlineStore
	events   []Event
	handlers map[string][]EventHandler
}

func NewMemoryHub(store OnlineStore) *MemoryHub {
	if store == nil {
		store = NewMemoryOnlineStore()
	}
	return &MemoryHub{
		store:    store,
		handlers: make(map[string][]EventHandler),
	}
}

func (h *MemoryHub) Emit(event string, data any) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, Event{Event: event, Data: data})
	return nil
}

func (h *MemoryHub) EmitTo(socketID string, event string, data any) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, Event{SocketID: socketID, Event: event, Data: data})
	return nil
}

func (h *MemoryHub) On(event string, handler EventHandler) {
	if event == "" || handler == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers[event] = append(h.handlers[event], handler)
}

func (h *MemoryHub) OnlineStore() OnlineStore {
	return h.store
}

func (h *MemoryHub) Events() []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]Event(nil), h.events...)
}

func (h *MemoryHub) Dispatch(event string, payload EventPayload) {
	h.mu.RLock()
	handlers := append([]EventHandler(nil), h.handlers[event]...)
	h.mu.RUnlock()
	for _, handler := range handlers {
		handler(payload)
	}
}

func encodeEventData(data any) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	switch value := data.(type) {
	case []byte:
		return value, nil
	case string:
		return json.Marshal(value)
	default:
		return json.Marshal(value)
	}
}
