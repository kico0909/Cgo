package redis

import (
	"log"
	"encoding/json"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/redis"
)

func New (conf *config.ConfigData) *reids.DatabaseRedis {

		log.Println("初始化 [ redis ] ...")
		var cgoRedis reids.DatabaseRedis

		strByte, _ := json.Marshal(conf.Redis.Setup)
		var redisSetupInfo reids.RedisSetupInfo
		json.Unmarshal(strByte, &redisSetupInfo)
		cgoRedis.Init(&redisSetupInfo)
		return &cgoRedis
}