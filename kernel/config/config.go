package config

import (
	"encoding/json"
	"github.com/Cgo/redis"
	//"github.com/Cgo/route"
	"io/ioutil"
	"time"
)

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

type mysqlSetOpt struct {
	Tag      string `json:"tag"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Dbname   string `json:"dbname"`
	Socket   string `json:"socket"`
}
type ConfigMysqlOptions struct {
	Key            bool          `json:"key"`
	ConnectionInfo []mysqlSetOpt `json:"connectionInfo"`
}
type ConfgigRedisOptions struct {
	Key   bool                 `json:"key"`
	Setup reids.RedisSetupInfo `json:"setup"`
}
type ConfigSessionOptions struct {
	Key             bool                 `json:"key"`
	SessionType     string               `json:"sessionType"`
	SessionName     string               `json:"sessionName"`
	SessionLifeTime int64                `json:"sessionLifeTime"`
	SessionRedis    reids.RedisSetupInfo `json:"sessionRedis"`
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

//type ConfigCasOptions struct {
//	Key                 bool     `json:"key"`
//	Url                 string   `json:"url"`
//	WhiteList           []string `json:"whiteList"`
//	APIPath             string   `json:"apiPath"`
//	SessionName         string   `json:"sessionName"`
//	LogoutRouter        string   `json:"logoutRouter"`
//	LogoutRequestMethod string   `json:"logoutRequestMethod"`
//	LogoutReUrl         string   `json:"logoutReUrl"`
//	LogoutValueName     string   `json:"logoutValueName"`
//	APIErrCode          string   `json:"apiErrCode"`
//}

type ConfigLoggerOptions struct {
	Key        bool   `json:"key"`
	Path       string `json:"path"`
	FileName   string `json:"fileName"`
	AutoCutOff bool   `json:"autoCutOff"`
}

type ConfigData struct {
	Server  ConfigServerOptions    `json:"server"`
	TLS     ConfigTLSOptions       `json:"tls"`
	Mysql   ConfigMysqlOptions     `json:"mysql"`
	Redis   ConfgigRedisOptions    `json:"redis"`
	Session ConfigSessionOptions   `json:"session"`
	Custom  map[string]interface{} `json:"custom"`
	//Cas     ConfigCasOptions       `json:"cas"`
	Log ConfigLoggerOptions `json:"log"`
}

type ConfigModule struct {
	Conf ConfigData
}

func (_self *ConfigModule) Set(path string) bool {

	cont, err := ioutil.ReadFile(path)

	if err != nil {
		return false
	}

	// 解析json文件
	if err := json.Unmarshal(cont, &_self.Conf); err != nil {
		return false
	}

	return true
}
