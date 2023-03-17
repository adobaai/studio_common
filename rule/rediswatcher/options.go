package rediswatcher

import (
	"errors"

	"github.com/casbin/casbin/v2"
	rds "github.com/go-redis/redis"
	"github.com/google/uuid"
)

type Log interface {
	Infof(format string, a ...any)
	Error(a ...any)
	Errorf(format string, args ...any)
}

type WatcherOptions struct {
	Rds                    *rds.Client
	E                      casbin.IEnforcer
	Channel                string
	LocalID                string
	IgnoreSelf             bool
	NoSubscribe            bool
	Log                    Log
	OptionalUpdateCallback func(string)
}

func initConfig(option *WatcherOptions) error {
	if option.E == nil {
		return errors.New("invalid enforcer entity")
	}
	if option.Log == nil {
		return errors.New("invalid log")
	}
	if option.LocalID == "" {
		option.LocalID = uuid.New().String()
	}
	if option.Channel == "" {
		option.Channel = "studio.policies"
	}
	return nil
}
