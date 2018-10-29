package config

import (
	"io/ioutil"
	"encoding/json"
	"github.com/Cgo/redis"
	"time"
	)
type letsEncrypt struct {
	Domain string `json:"domain"`
	Email string `json:"email"`
}
type tlsData struct {
	Key bool `json:"open"`
	LetsEncrypt bool `json:"letsEncrypt"`
	LetsEncryptOpt letsEncrypt `json:"letsEncryptOpt"`
	KeyPath string `json:"keyPath"`
	CertPath string `json:"certPath"`

}
type mysqlSetOpt struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host string `json:"host"`
	Port string `json:"port"`
	Dbname string `json:"dbname"`
	Socket string `json:"socket"`
}
type mysqlData struct {
	Key bool `json:"key"`
	Default mysqlSetOpt `json:"default"`
}
type redisData struct {
	Key bool `json:"key"`
	Setup reids.RedisSetupInfo
}

type sessionOpt struct {
	Key bool `json:"key"`
	SessionType string `json:"sessionType"`
	SessionName string `json:"sessionName"`
	SessionLifeTime int64 `json:"sessionLifeTime"`
	SessionRedis reids.RedisSetupInfo `json:"sessionRedis"`
}
type serverOption struct {
	Port int64	`json:"port"`
	IsStatic bool	`json:"is_static"`
	StaticPath string	`json:"staticPath"`
	TemplatePath string	`json:"templatePath"`
	ReadTimeout	time.Duration	`json:"readTimeout"`
	WriteTimeout time.Duration	`json:"writeTimeout"`
	MaxHeaderBytes int	`json:"max_header_bytes"`
}

type ConfigData struct {
	Server serverOption	`json:"server"`
	TLS tlsData `json:"tls"`
	Mysql mysqlData `json:"mysql"`
	Redis redisData `json:"redis"`
	Session sessionOpt `json:"session"`
	Custom map[string]interface{} `json:"custom"`
}

type ConfigModule struct {
	Conf ConfigData
}

func (_self *ConfigModule) Set(path string)bool{


	cont, err := ioutil.ReadFile(path)

	if err!=nil {
		return false
	}

	if err := json.Unmarshal(cont, &_self.Conf); err != nil {
		return false
	}


	return true

}