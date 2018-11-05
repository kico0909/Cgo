package config

import (
	"io/ioutil"
	"encoding/json"
	"github.com/Cgo/redis"
	"time"
	"log"
)
type letsEncrypt struct {
	Domain 	string 	`json:"domain"`
	Email 	string 	`json:"email"`
}
type ConfigTLSOptions struct {
	Key 			bool 		`json:"open"`
	LetsEncrypt 	bool 		`json:"letsEncrypt"`
	LetsEncryptOpt 	letsEncrypt `json:"letsEncryptOpt"`
	KeyPath 		string 		`json:"keyPath"`
	CertPath 		string 		`json:"certPath"`

}
//type mysqlSetOpt struct {
//	Username 	string `json:"username"`
//	Password 	string `json:"password"`
//	Host 		string `json:"host"`
//	Port 		string `json:"port"`
//	Dbname 		string `json:"dbname"`
//	Socket 		string `json:"socket"`
//}
type mysqlSetOpt struct {
	Tag			string	`json:"tag"`
	Username 	string 	`json:"username"`
	Password 	string 	`json:"password"`
	Host 		string 	`json:"host"`
	Port 		string 	`json:"port"`
	Dbname 		string 	`json:"dbname"`
	Socket 		string 	`json:"socket"`
}
type ConfigMysqlOptions struct {
	Key 			bool	 		`json:"key"`
	ConnectionInfo 	[]mysqlSetOpt 	`json:"connectionInfo"`
}
type ConfgigRedisOptions struct {
	Key 	bool 					`json:"key"`
	Setup 	reids.RedisSetupInfo	`json:"setup"`
}
type ConfigSessionOptions struct {
	Key 			bool 					`json:"key"`
	SessionType 	string 					`json:"sessionType"`
	SessionName 	string 					`json:"sessionName"`
	SessionLifeTime int64 					`json:"sessionLifeTime"`
	SessionRedis 	reids.RedisSetupInfo 	`json:"sessionRedis"`
}
type ConfigServerOptions struct {
	Port 			int64			`json:"port"`
	StaticRouter 	string			`json:"staticRouter"`
	StaticPath 		string			`json:"staticPath"`
	TemplatePath 	string			`json:"templatePath"`
	ReadTimeout		time.Duration	`json:"readTimeout"`
	WriteTimeout 	time.Duration	`json:"writeTimeout"`
	MaxHeaderBytes 	int				`json:"maxHeaderBytes"`
}
type ConfigCasOptions struct {
	Key 				bool 		`json:"key"`
	Url 				string 		`json:"url"`
	WhiteList 			[]string 	`json:"whiteList"`
	APIPath 			string 		`json:"apiPath"`
	SessionName 		string 		`json:"sessionName"`
	LogoutRouter		string		`json:"logoutRouter"`
	LogoutRequestMethod	string		`json:"logoutRequestMethod"`
	LogoutReUrl			string		`json:"logoutReUrl"`
	LogoutValueName		string		`json:"logoutValueName"`
	APIErrCode			string		`json:"apiErrCode"`
}


type ConfigData struct {
	Server 	ConfigServerOptions		`json:"server"`
	TLS 	ConfigTLSOptions 		`json:"tls"`
	Mysql 	ConfigMysqlOptions 		`json:"mysql"`
	Redis 	ConfgigRedisOptions 	`json:"redis"`
	Session ConfigSessionOptions 	`json:"session"`
	Custom 	map[string]interface{} 	`json:"custom"`
	Cas 	ConfigCasOptions		`json:"cas"`
}

type ConfigModule struct {
	Conf ConfigData
}

func (_self *ConfigModule) Set(path string)bool{

	log.Println("功能初始化: Cgo配置文件("+path+") --- [ ok ]")

	cont, err := ioutil.ReadFile(path)

	if err!=nil {
		return false
	}

	if err := json.Unmarshal(cont, &_self.Conf); err != nil {
		return false
	}

	return true

}