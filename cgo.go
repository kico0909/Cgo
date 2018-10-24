package Cgo

import (
		"github.com/Cgo/mysql"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/route"
	"github.com/Cgo/kernel/session"
	"github.com/Cgo/kernel/command"
	cgoMysql "github.com/Cgo/kernel/mysql"
	cgoRedis "github.com/Cgo/kernel/redis"
	"html/template"
	"github.com/Cgo/kernel/module"
	"os"
	"os/exec"
	"github.com/Cgo/redis"
)


var Config config.ConfigModule	// 配置
var Router *route.RouterManager	// 路由
var Session *session.CgoSession	// session
var Redis *reids.DatabaseRedis	// redis
var Mysql *mysql.DatabaseMysql	// mysql
var TemplateCached *template.Template	// 模板缓存文件

type Module module.DBModel
type RouterHandler route.RouterHandler

var (
	comm string
	daemon bool
)

func Run(){
	// 1. 检测启动参数是否正确
	for _,v := range os.Args {
		if v == "start" && len(comm)<1 {
			comm = "start"
		}
		if v == "stop" && len(comm)<1 {
			comm = "stop"
		}
		if v == "restart" && len(comm)<1 {
			comm = "restart"
		}
		if v =="-d" || v == "-domain" {
			daemon = true
		}
	}

	if daemon {
		cmd := exec.Command(os.Args[0], comm)
		cmd.Start()
		daemon = false
		os.Exit(0)
	}

	if comm == "start" {

		// 2. 启动session 如果session 设置了
		if Config.Conf.Session.Key {
			Session = Session.New(&Config.Conf)
		}

		// 3. mysql 初始化
		if Config.Conf.Mysql.Key {
			Mysql = cgoMysql.New(&Config.Conf)
		}

		// 4. redis 初始化
		if Config.Conf.Redis.Key {
			Redis = cgoRedis.New(&Config.Conf)
		}
	}

	// 执行启动
	command.Run(&comm, Router, &Config.Conf)


}


func init(){
	// 初始化全局路由变量
	Router = route.NewRouter()
}

