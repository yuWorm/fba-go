package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gofiber/contrib/v3/socketio"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/yuWorm/fba-go/core/config"
)

const socketIOServerIDAttribute = "fba.realtime.server_id"

type SocketIOHub struct {
	mu          sync.RWMutex
	nodeID      string
	store       OnlineStore
	broadcaster Broadcaster
	connections map[string]*socketio.Websocket
	handlers    map[string][]EventHandler
}

type SocketIOHubOption func(*SocketIOHub)

func WithNodeID(nodeID string) SocketIOHubOption {
	return func(h *SocketIOHub) {
		if nodeID != "" {
			h.nodeID = nodeID
		}
	}
}

func WithBroadcaster(broadcaster Broadcaster) SocketIOHubOption {
	return func(h *SocketIOHub) {
		h.broadcaster = broadcaster
	}
}

func NewSocketIOHub(store OnlineStore, opts ...SocketIOHubOption) *SocketIOHub {
	if store == nil {
		store = NewMemoryOnlineStore()
	}
	hub := &SocketIOHub{
		nodeID:      defaultNodeID(),
		store:       store,
		connections: make(map[string]*socketio.Websocket),
		handlers:    make(map[string][]EventHandler),
	}
	for _, opt := range opts {
		opt(hub)
	}
	return hub
}

func (h *SocketIOHub) Emit(event string, data any) error {
	payload, err := encodeEventData(data)
	if err != nil {
		return err
	}
	h.emitLocal(event, payload)
	if h.broadcaster != nil {
		if err := h.broadcaster.Publish(context.Background(), BroadcastMessage{
			Origin: h.nodeID,
			Event:  event,
			Data:   json.RawMessage(payload),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (h *SocketIOHub) emitLocal(event string, payload []byte) {
	h.mu.RLock()
	connections := make([]*socketio.Websocket, 0, len(h.connections))
	for _, conn := range h.connections {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()
	for _, conn := range connections {
		conn.EmitEvent(event, payload)
	}
}

func (h *SocketIOHub) EmitTo(socketID string, event string, data any) error {
	payload, err := encodeEventData(data)
	if err != nil {
		return err
	}
	h.mu.RLock()
	conn := h.connections[socketID]
	h.mu.RUnlock()
	if conn == nil {
		return socketio.ErrorInvalidConnection
	}
	conn.EmitEvent(event, payload)
	return nil
}

func (h *SocketIOHub) On(event string, handler EventHandler) {
	if event == "" || handler == nil {
		return
	}
	registerSocketIOEvent(event)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers[event] = append(h.handlers[event], handler)
}

func (h *SocketIOHub) OnlineStore() OnlineStore {
	return h.store
}

func (h *SocketIOHub) StartBroadcaster(ctx context.Context) error {
	if h.broadcaster == nil {
		return nil
	}
	return h.broadcaster.Start(ctx, func(message BroadcastMessage) {
		_ = h.ReceiveBroadcast(message)
	})
}

func (h *SocketIOHub) ShutdownBroadcaster(context.Context) error {
	if h.broadcaster == nil {
		return nil
	}
	return h.broadcaster.Close()
}

func (h *SocketIOHub) ReceiveBroadcast(message BroadcastMessage) bool {
	if message.Origin == "" || message.Origin == h.nodeID || message.Event == "" {
		return false
	}
	h.emitLocal(message.Event, message.Data)
	return true
}

func (h *SocketIOHub) connect(socketID string, sessionUUID string, conn *socketio.Websocket) {
	h.mu.Lock()
	h.connections[socketID] = conn
	h.mu.Unlock()
	h.store.Connect(socketID, sessionUUID)
}

func (h *SocketIOHub) disconnect(socketID string) {
	h.mu.Lock()
	delete(h.connections, socketID)
	h.mu.Unlock()
	h.store.Disconnect(socketID)
}

func (h *SocketIOHub) dispatch(event string, payload EventPayload) {
	h.mu.RLock()
	handlers := append([]EventHandler(nil), h.handlers[event]...)
	h.mu.RUnlock()
	for _, handler := range handlers {
		handler(payload)
	}
}

type SocketIOServer struct {
	id            string
	hub           *SocketIOHub
	authenticator Authenticator
	config        config.Options
}

type SocketIOServerOptions struct {
	Config        config.Options
	Authenticator Authenticator
}

func NewSocketIOServer(hub *SocketIOHub, opts SocketIOServerOptions) *SocketIOServer {
	cfg := opts.Config.WithDefaults()
	if hub == nil {
		hub = NewSocketIOHub(NewMemoryOnlineStore())
	}
	return &SocketIOServer{
		id:            uuid.NewString(),
		hub:           hub,
		authenticator: opts.Authenticator,
		config:        cfg,
	}
}

func (s *SocketIOServer) Mount(router fiber.Router) {
	if router == nil {
		return
	}
	registerSocketIOServer(s)
	registerSocketIOLifecycle()
	socketio.EnablePolling = s.config.Realtime.EnablePolling
	handler := socketio.New(func(kws *socketio.Websocket) {
		// gofiber/contrib/socketio keeps event listeners package-global and
		// offers no unregister hook. Stamping each connection with its server id
		// lets global listeners route payloads back to the correct host instance.
		kws.SetAttribute(socketIOServerIDAttribute, s.id)
		kws.SetAttribute("connected_at", time.Now().UTC().Format(time.RFC3339))
	})
	path := s.config.Realtime.Path
	router.Get(path, handler)
	if s.config.Realtime.EnablePolling {
		router.Post(path, handler)
		router.Options(path, handler)
	}
}

func (s *SocketIOServer) handleConnect(payload *socketio.EventPayload) {
	authPayload, err := parseHandshakeAuth(payload.HandshakeAuth)
	if err != nil {
		payload.Kws.Close()
		return
	}
	if s.authenticator != nil {
		if err := s.authenticator.Authenticate(context.Background(), authPayload); err != nil {
			payload.Kws.Close()
			return
		}
	}
	payload.Kws.SetAttribute("session_uuid", authPayload.SessionUUID)
	s.hub.connect(payload.SocketUUID, authPayload.SessionUUID, payload.Kws)
}

func (s *SocketIOServer) handleDisconnect(payload *socketio.EventPayload) {
	s.hub.disconnect(payload.SocketUUID)
}

func parseHandshakeAuth(raw json.RawMessage) (AuthPayload, error) {
	if len(raw) == 0 {
		return AuthPayload{}, ErrMissingAuth
	}
	var payload AuthPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return AuthPayload{}, fmt.Errorf("%w: %v", ErrInvalidAuth, err)
	}
	if payload.Token == "" || payload.SessionUUID == "" {
		return AuthPayload{}, ErrMissingAuth
	}
	return payload, nil
}

func Shutdown(ctx context.Context) error {
	return socketio.Shutdown(ctx)
}

var (
	socketIOServersMu       sync.RWMutex
	socketIOServers         = make(map[string]*SocketIOServer)
	socketIOLifecycleOnce   sync.Once
	socketIOEventMu         sync.Mutex
	socketIOEventRegistered = make(map[string]struct{})
)

func registerSocketIOServer(server *SocketIOServer) {
	socketIOServersMu.Lock()
	defer socketIOServersMu.Unlock()
	socketIOServers[server.id] = server
}

func socketIOServerFor(payload *socketio.EventPayload) *SocketIOServer {
	if payload == nil || payload.Kws == nil {
		return nil
	}
	id := payload.Kws.GetStringAttribute(socketIOServerIDAttribute)
	if id == "" {
		return nil
	}
	socketIOServersMu.RLock()
	defer socketIOServersMu.RUnlock()
	return socketIOServers[id]
}

func registerSocketIOLifecycle() {
	socketIOLifecycleOnce.Do(func() {
		socketio.On(socketio.EventConnect, func(payload *socketio.EventPayload) {
			if server := socketIOServerFor(payload); server != nil {
				server.handleConnect(payload)
			}
		})
		socketio.On(socketio.EventDisconnect, func(payload *socketio.EventPayload) {
			if server := socketIOServerFor(payload); server != nil {
				server.handleDisconnect(payload)
			}
		})
	})
}

func registerSocketIOEvent(event string) {
	socketIOEventMu.Lock()
	defer socketIOEventMu.Unlock()
	if _, ok := socketIOEventRegistered[event]; ok {
		return
	}
	socketIOEventRegistered[event] = struct{}{}
	socketio.On(event, func(payload *socketio.EventPayload) {
		server := socketIOServerFor(payload)
		if server == nil {
			return
		}
		server.hub.dispatch(event, EventPayload{
			SocketID: payload.SocketUUID,
			Data:     payload.Data,
			Args:     append([][]byte(nil), payload.Args...),
			Ack:      ackFunc(payload),
		})
	})
}

func ackFunc(payload *socketio.EventPayload) func(args ...any) error {
	return func(args ...any) error {
		encoded := make([][]byte, 0, len(args))
		for _, arg := range args {
			item, err := encodeEventData(arg)
			if err != nil {
				return err
			}
			if item == nil {
				item = []byte("null")
			}
			encoded = append(encoded, item)
		}
		return payload.Ack(encoded...)
	}
}

func defaultNodeID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "fba-go"
	}
	return hostname + "-" + uuid.NewString()
}
