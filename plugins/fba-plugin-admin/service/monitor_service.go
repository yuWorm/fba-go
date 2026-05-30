package service

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/yuWorm/fba-plugin-admin/dto"
	"github.com/yuWorm/fba-plugin-admin/repo"
)

type MonitorService struct {
	repo    repo.Repository
	started time.Time
}

func NewMonitorService(repository repo.Repository) *MonitorService {
	if repository == nil {
		repository = repo.NewMemoryRepository(repo.SeedData())
	}
	return &MonitorService{repo: repository, started: time.Now()}
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

func (s *MonitorService) Redis(context.Context) (dto.RedisMonitorInfo, error) {
	// Until the Go module wires a real Redis provider, expose a schema-compatible
	// local fallback instead of leaking empty fixture strings to the frontend.
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
	}, nil
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
