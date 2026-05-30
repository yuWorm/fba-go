package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type ContractSnapshot struct {
	RouteCount       int      `json:"route_count"`
	PriorityRoutes   []Route  `json:"priority_routes"`
	ResponseEnvelope bool     `json:"response_envelope"`
	RedisKeys        []string `json:"redis_keys"`
}

func Snapshot(contracts Contracts) (ContractSnapshot, error) {
	keys := make([]string, 0, len(contracts.Redis.Keys))
	for name := range contracts.Redis.Keys {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return ContractSnapshot{
		RouteCount:       len(contracts.API.Routes),
		PriorityRoutes:   contracts.API.PriorityRoutes,
		ResponseEnvelope: contracts.Response.Success.Envelope,
		RedisKeys:        keys,
	}, nil
}

func WriteSnapshot(path string, snapshot ContractSnapshot) error {
	content, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}
