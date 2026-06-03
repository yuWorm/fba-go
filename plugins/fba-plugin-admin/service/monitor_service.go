package service

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type MonitorService struct {
	repo    repo.Repository
	redis   RedisClient
	started time.Time
}

func NewMonitorService(repository repo.Repository) *MonitorService {
	return NewMonitorServiceWithRedis(repository, nil)
}

func NewMonitorServiceWithRedis(repository repo.Repository, redisClient RedisClient) *MonitorService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &MonitorService{repo: repository, redis: redisClient, started: time.Now()}
}

func (s *MonitorService) Server(context.Context) (dto.ServerMonitorInfo, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	total := float64(mem.Sys)
	used := float64(mem.Alloc)
	free := math.Max(total-used, 0)
	usage := 0.0
	if total > 0 {
		usage = round2(used / total * 100)
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "localhost"
	}
	executable, err := os.Executable()
	if err != nil {
		executable = ""
	}

	return dto.ServerMonitorInfo{
		CPU: dto.CPUInfo{
			PhysicalNum: runtime.NumCPU(),
			LogicalNum:  runtime.NumCPU(),
			MaxFreq:     0,
			MinFreq:     0,
			CurrentFreq: 0,
			Usage:       0,
		},
		Mem: dto.MemInfo{
			Total: round2(total / gb),
			Used:  round2(used / gb),
			Free:  round2(free / gb),
			Usage: usage,
		},
		Sys: dto.SysInfo{
			Name: hostname,
			OS:   runtime.GOOS,
			IP:   localIP(),
			Arch: runtime.GOARCH,
		},
		Disk: diskInfo(),
		Service: dto.ServiceInfo{
			Name:     "fba-go",
			Version:  runtime.Version(),
			Home:     executable,
			Startup:  s.started.Format(dto.TimeLayout),
			Elapsed:  formatDuration(time.Since(s.started)),
			CPUUsage: "0.00%",
			MemVMS:   formatBytes(mem.Sys),
			MemRSS:   formatBytes(mem.Alloc),
			MemFree:  formatBytes(uint64(free)),
		},
	}, nil
}

func (s *MonitorService) Redis(ctx context.Context) (dto.RedisMonitorInfo, error) {
	if s.redis == nil {
		return fallbackRedisMonitor(), nil
	}
	raw, err := s.redis.Info(ctx, "server", "clients", "memory", "stats", "commandstats", "keyspace").Result()
	if err != nil {
		return dto.RedisMonitorInfo{}, err
	}
	return redisMonitorFromInfo(raw), nil
}

func fallbackRedisMonitor() dto.RedisMonitorInfo {
	return dto.RedisMonitorInfo{
		Info: dto.RedisServerInfo{
			RedisVersion:          "unavailable",
			RedisMode:             "standalone",
			Role:                  "master",
			TCPPort:               "6379",
			Uptime:                "0s",
			ConnectedClients:      "0",
			BlockedClients:        "0",
			UsedMemoryHuman:       "0B",
			UsedMemoryRSSHuman:    "0B",
			MaxMemoryHuman:        "0B",
			MemFragmentationRatio: "0",
			InstantaneousOps:      "0",
			TotalCommands:         "0",
			RejectedConnections:   "0",
			KeysNum:               "0",
		},
		Stats: []dto.RedisCommandStat{
			{Name: "ping", Value: "0"},
		},
	}
}

func (s *MonitorService) Sessions(ctx context.Context, username string) ([]dto.SessionDetail, error) {
	items, err := s.repo.ListSessions(ctx, repo.SessionFilter{Username: username})
	if err != nil {
		return nil, err
	}
	return dto.SessionsFromModel(items), nil
}

func (s *MonitorService) DeleteSession(ctx context.Context, userID int, sessionUUID string) error {
	return s.repo.DeleteSession(ctx, userID, sessionUUID)
}

const gb = 1024 * 1024 * 1024

func diskInfo() []dto.DiskInfo {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil || stat.Blocks == 0 {
		return []dto.DiskInfo{}
	}
	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	free := uint64(stat.Bavail) * uint64(stat.Bsize)
	used := total - free
	usage := 0.0
	if total > 0 {
		usage = round2(float64(used) / float64(total) * 100)
	}
	return []dto.DiskInfo{
		{
			Dir:    "/",
			Device: "/",
			Type:   "local",
			Total:  formatBytes(total),
			Used:   formatBytes(used),
			Free:   formatBytes(free),
			Usage:  fmt.Sprintf("%.2f%%", usage),
		},
	}
}

func redisMonitorFromInfo(raw string) dto.RedisMonitorInfo {
	values := parseRedisInfo(raw)
	stats := redisCommandStats(values)
	if len(stats) == 0 {
		stats = []dto.RedisCommandStat{{Name: "ping", Value: "0"}}
	}
	return dto.RedisMonitorInfo{
		Info: dto.RedisServerInfo{
			RedisVersion:          values["redis_version"],
			RedisMode:             defaultString(values["redis_mode"], "standalone"),
			Role:                  defaultString(values["role"], "master"),
			TCPPort:               defaultString(values["tcp_port"], "6379"),
			Uptime:                formatRedisUptime(values["uptime_in_seconds"]),
			ConnectedClients:      defaultString(values["connected_clients"], "0"),
			BlockedClients:        defaultString(values["blocked_clients"], "0"),
			UsedMemoryHuman:       defaultString(values["used_memory_human"], "0B"),
			UsedMemoryRSSHuman:    defaultString(values["used_memory_rss_human"], "0B"),
			MaxMemoryHuman:        defaultString(values["maxmemory_human"], "0B"),
			MemFragmentationRatio: defaultString(values["mem_fragmentation_ratio"], "0"),
			InstantaneousOps:      defaultString(values["instantaneous_ops_per_sec"], "0"),
			TotalCommands:         defaultString(values["total_commands_processed"], "0"),
			RejectedConnections:   defaultString(values["rejected_connections"], "0"),
			KeysNum:               redisKeysNum(values),
		},
		Stats: stats,
	}
}

func parseRedisInfo(raw string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		values[key] = value
	}
	return values
}

func redisCommandStats(values map[string]string) []dto.RedisCommandStat {
	stats := make([]dto.RedisCommandStat, 0)
	for key, value := range values {
		if !strings.HasPrefix(key, "cmdstat_") {
			continue
		}
		command := strings.TrimPrefix(key, "cmdstat_")
		calls := "0"
		for _, part := range strings.Split(value, ",") {
			name, raw, ok := strings.Cut(part, "=")
			if ok && name == "calls" {
				calls = raw
				break
			}
		}
		stats = append(stats, dto.RedisCommandStat{Name: command, Value: calls})
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Name < stats[j].Name
	})
	return stats
}

func redisKeysNum(values map[string]string) string {
	total := 0
	for key, value := range values {
		if !strings.HasPrefix(key, "db") {
			continue
		}
		for _, part := range strings.Split(value, ",") {
			name, raw, ok := strings.Cut(part, "=")
			if !ok || name != "keys" {
				continue
			}
			count, err := strconv.Atoi(raw)
			if err == nil {
				total += count
			}
		}
	}
	return strconv.Itoa(total)
}

func formatRedisUptime(raw string) string {
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return "0s"
	}
	return formatDuration(time.Duration(seconds) * time.Second)
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func localIP() string {
	conn, err := net.DialTimeout("udp", "8.8.8.8:80", 100*time.Millisecond)
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok && addr.IP != nil {
		return addr.IP.String()
	}
	return "127.0.0.1"
}

func formatBytes(value uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(value)
	unit := 0
	for size >= 1024 && unit < len(units)-1 {
		size /= 1024
		unit++
	}
	if unit == 0 {
		return strconv.FormatUint(value, 10) + units[unit]
	}
	return fmt.Sprintf("%.2f%s", size, units[unit])
}

func formatDuration(value time.Duration) string {
	seconds := int64(value.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, seconds%60)
	}
	hours := minutes / 60
	return fmt.Sprintf("%dh%dm%ds", hours, minutes%60, seconds%60)
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
