package redis

import (
	"github.com/Cgo/kernel/logger"
	"encoding/json"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/redis"
)

func New (conf *config.ConfgigRedisOptions) *reids.DatabaseRedis {
	log.Println("功能初始化: Redis	 --- [ ok ]")

	var cgoRedis reids.DatabaseRedis
	strByte, _ := json.Marshal(conf.Setup)
	var redisSetupInfo reids.RedisSetupInfo
	json.Unmarshal(strByte, &redisSetupInfo)
	cgoRedis.Init(&redisSetupInfo)
	return &cgoRedis
}