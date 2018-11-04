package redis

import (
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/redis"
)

func New (conf *config.ConfigData) *reids.DatabaseRedis {
	log.Info("功能初始化: Redis								[ ok ]")

	var cgoRedis reids.DatabaseRedis
	strByte, _ := json.Marshal(conf.Redis.Setup)
	var redisSetupInfo reids.RedisSetupInfo
	json.Unmarshal(strByte, &redisSetupInfo)
	cgoRedis.Init(&redisSetupInfo)
	return &cgoRedis
}