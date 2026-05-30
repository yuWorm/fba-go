package redisx

import "fmt"

const DefaultPrefix = "fba"

type Keys struct {
	prefix string
}

func NewKeys(prefix string) Keys {
	if prefix == "" {
		prefix = DefaultPrefix
	}
	return Keys{prefix: prefix}
}

func (k Keys) AccessToken(userID int64, sessionUUID string) string {
	return k.format("token:%d:%s", userID, sessionUUID)
}

func (k Keys) TokenExtra(userID int64, sessionUUID string) string {
	return k.format("token_extra_info:%d:%s", userID, sessionUUID)
}

func (k Keys) OnlineSet() string {
	return k.format("token_online")
}

func (k Keys) RefreshToken(userID int64, sessionUUID string) string {
	return k.format("refresh_token:%d:%s", userID, sessionUUID)
}

func (k Keys) User(userID int64) string {
	return k.format("user:%d", userID)
}

func (k Keys) LoginCaptcha(uuid string) string {
	return k.format("login:captcha:%s", uuid)
}

func (k Keys) LoginFailure(userID int64) string {
	return k.format("login:failure:%d", userID)
}

func (k Keys) UserLock(userID int64) string {
	return k.format("user:lock:%d", userID)
}

func (k Keys) PluginState(plugin string) string {
	return k.format("plugin:%s", plugin)
}

func (k Keys) PluginChanged() string {
	return k.format("plugin:changed")
}

func (k Keys) CacheInvalidateChannel() string {
	return k.format("cache:invalidate")
}

func (k Keys) SnowflakeNode(datacenter string, worker string) string {
	return k.format("snowflake:nodes:%s:%s", datacenter, worker)
}

func (k Keys) MigrationLock() string {
	return k.format("migration:lock")
}

func (k Keys) SchedulerLeader() string {
	return k.format("task:scheduler:leader")
}

func (k Keys) AsynqPrefix() string {
	return k.format("asynq")
}

func (k Keys) format(format string, args ...any) string {
	return k.prefix + ":" + fmt.Sprintf(format, args...)
}
