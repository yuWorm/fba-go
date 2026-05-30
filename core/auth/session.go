package auth

import "github.com/yuWorm/fba-go/core/redisx"

type SessionKeyBuilder struct {
	keys redisx.Keys
}

type SessionKeys struct {
	AccessToken  string
	RefreshToken string
	UserCache    string
}

func NewSessionKeys(keys redisx.Keys) SessionKeyBuilder {
	return SessionKeyBuilder{keys: keys}
}

func (b SessionKeyBuilder) ForSession(userID int64, sessionUUID string) SessionKeys {
	return SessionKeys{
		AccessToken:  b.keys.AccessToken(userID, sessionUUID),
		RefreshToken: b.keys.RefreshToken(userID, sessionUUID),
		UserCache:    b.keys.User(userID),
	}
}
