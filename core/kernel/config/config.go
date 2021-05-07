package config

import (
	"github.com/kico0909/cgo/core/kernel/logger"
	"github.com/kico0909/cgo/core/plugins/iniHandler"
	"github.com/kico0909/cgo/core/redis"
	"time"
)

type ConfigData struct {
	App     int64                  `json:"app"`
	Server  ConfigServerOptions    `json:"server"`
	TLS     ConfigTLSOptions       `json:"tls"`
	Mysql   ConfigMysqlOptions     `json:"mysql"`
	Redis   ConfgigRedisOptions    `json:"redis"`
	Session ConfigSessionOptions   `json:"session"`
	Custom  map[string]interface{} `json:"custom"`
	Log     ConfigLoggerOptions    `json:"log"`
}

type ConfigServerOptions struct {
	Port                 int64         `json:"port"`
	StaticRouter         string        `json:"staticRouter"`
	StaticPath           string        `json:"staticPath"`
	TemplatePath         string        `json:"templatePath"`
	ReadTimeout          time.Duration `json:"readTimeout"`
	WriteTimeout         time.Duration `json:"writeTimeout"`
	MaxHeaderBytes       int           `json:"maxHeaderBytes"`
	AllowOtherAjaxOrigin bool          `json:"allowOtherAjaxOrigin"`
}

type letsEncrypt struct {
	Domain string `json:"domain"`
	Email  string `json:"email"`
}

type ConfigTLSOptions struct {
	Key            bool        `json:"key"`
	LetsEncrypt    bool        `json:"letsEncrypt"`
	LetsEncryptOpt letsEncrypt `json:"letsEncryptOpt"`
	KeyPath        string      `json:"keyPath"`
	CertPath       string      `json:"certPath"`
}

type MysqlSetOpt struct {
	Tag      string `json:"tag"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	Dbname   string `json:"dbname"`
	Socket   string `json:"socket"`
	Charset  string `json:"charset"`
}

type ConfigMysqlOptions struct {
	Key     bool                    `json:"key"`
	MaxOpen int                     `json:"maxOpen"`
	MaxIdle int                     `json:"maxIdle"`
	Default MysqlSetOpt             `json:"default"`
	Write   MysqlSetOpt             `json:"write"`
	Read    MysqlSetOpt             `json:"read"`
	Childs  map[string]*MysqlSetOpt `json:"childs"`
}

type ConfgigRedisOptions struct {
	Key   bool                 `json:"key"`
	Setup redis.RedisSetupInfo `json:"setup"`
}

type ConfigSessionOptions struct {
	Key             bool                 `json:"key"`
	SessionType     string               `json:"sessionType"`
	SessionName     string               `json:"sessionName"`
	SessionLifeTime int64                `json:"sessionLifeTime"`
	Redis           redis.RedisSetupInfo `json:"redis"`
}

type ConfigLoggerOptions struct {
	Key        bool   `json:"key"`
	Path       string `json:"path"`
	FileName   string `json:"fileName"`
	StopCutOff bool   `json:"stopCutOff"`
}

type ConfigModule struct {
	Conf ConfigData
}

func (_self *ConfigModule) Set(path string) bool {

	err := iniHandler.InitFile(path, &_self.Conf)

	if err != nil {
		log.Println("<功能初始化> 初始化配置文件失败 ", err)
		return false
	}

	return true
}
