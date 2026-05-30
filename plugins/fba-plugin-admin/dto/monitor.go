package dto

import "github.com/yuWorm/fba-plugin-admin/model"

type CPUInfo struct {
	PhysicalNum int     `json:"physical_num"`
	LogicalNum  int     `json:"logical_num"`
	MaxFreq     float64 `json:"max_freq"`
	MinFreq     float64 `json:"min_freq"`
	CurrentFreq float64 `json:"current_freq"`
	Usage       float64 `json:"usage"`
}

type MemInfo struct {
	Total float64 `json:"total"`
	Used  float64 `json:"used"`
	Free  float64 `json:"free"`
	Usage float64 `json:"usage"`
}

type SysInfo struct {
	Name string `json:"name"`
	OS   string `json:"os"`
	IP   string `json:"ip"`
	Arch string `json:"arch"`
}

type DiskInfo struct {
	Dir    string `json:"dir"`
	Device string `json:"device"`
	Type   string `json:"type"`
	Total  string `json:"total"`
	Used   string `json:"used"`
	Free   string `json:"free"`
	Usage  string `json:"usage"`
}

type ServiceInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Home     string `json:"home"`
	Startup  string `json:"startup"`
	Elapsed  string `json:"elapsed"`
	CPUUsage string `json:"cpu_usage"`
	MemVMS   string `json:"mem_vms"`
	MemRSS   string `json:"mem_rss"`
	MemFree  string `json:"mem_free"`
}

type ServerMonitorInfo struct {
	CPU     CPUInfo     `json:"cpu"`
	Mem     MemInfo     `json:"mem"`
	Sys     SysInfo     `json:"sys"`
	Disk    []DiskInfo  `json:"disk"`
	Service ServiceInfo `json:"service"`
}

type RedisServerInfo struct {
	RedisVersion          string `json:"redis_version"`
	RedisMode             string `json:"redis_mode"`
	Role                  string `json:"role"`
	TCPPort               string `json:"tcp_port"`
	Uptime                string `json:"uptime"`
	ConnectedClients      string `json:"connected_clients"`
	BlockedClients        string `json:"blocked_clients"`
	UsedMemoryHuman       string `json:"used_memory_human"`
	UsedMemoryRSSHuman    string `json:"used_memory_rss_human"`
	MaxMemoryHuman        string `json:"maxmemory_human"`
	MemFragmentationRatio string `json:"mem_fragmentation_ratio"`
	InstantaneousOps      string `json:"instantaneous_ops_per_sec"`
	TotalCommands         string `json:"total_commands_processed"`
	RejectedConnections   string `json:"rejected_connections"`
	KeysNum               string `json:"keys_num"`
}

type RedisCommandStat struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RedisMonitorInfo struct {
	Info  RedisServerInfo    `json:"info"`
	Stats []RedisCommandStat `json:"stats"`
}

type SessionDetail struct {
	ID            int    `json:"id"`
	SessionUUID   string `json:"session_uuid"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	IP            string `json:"ip"`
	OS            string `json:"os"`
	Browser       string `json:"browser"`
	Device        string `json:"device"`
	Status        int    `json:"status"`
	LastLoginTime string `json:"last_login_time"`
	ExpireTime    string `json:"expire_time"`
}

func SessionFromModel(item model.Session) SessionDetail {
	return SessionDetail{
		ID:            item.ID,
		SessionUUID:   item.SessionUUID,
		Username:      item.Username,
		Nickname:      item.Nickname,
		IP:            item.IP,
		OS:            item.OS,
		Browser:       item.Browser,
		Device:        item.Device,
		Status:        item.Status,
		LastLoginTime: item.LastLoginTime,
		ExpireTime:    formatTime(item.ExpireTime),
	}
}

func SessionsFromModel(items []model.Session) []SessionDetail {
	result := make([]SessionDetail, 0, len(items))
	for _, item := range items {
		result = append(result, SessionFromModel(item))
	}
	return result
}
