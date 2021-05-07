package cgo

import (
	"github.com/kico0909/cgo/core/kernel/command"
	"github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/kernel/dataModel"
	"github.com/kico0909/cgo/core/kernel/logger"
	cgologer "github.com/kico0909/cgo/core/kernel/logger"
	cgoMysql "github.com/kico0909/cgo/core/kernel/mysql"
	cgoRedis "github.com/kico0909/cgo/core/kernel/redis"
	"github.com/kico0909/cgo/core/kernel/session"
	"github.com/kico0909/cgo/core/kernel/template"
	"github.com/kico0909/cgo/core/mysql"
	"github.com/kico0909/cgo/core/plugins"
	reids "github.com/kico0909/cgo/core/redis"
	"github.com/kico0909/cgo/core/route"
	"os"
	"os/exec"
)

type RouterHandler = route.RouterHandler
type WSconn = route.WSConn
type DataModle = dataModel.DataModle

var Config config.ConfigModule                 // 配置
var Router *route.RouterManager                // 路由
var Session *session.CgoSession                // session
var Redis *reids.DatabaseRedis                 // redis
var MysqlDefault *mysql.DatabaseMysql          // mysql
var MysqlLinks map[string]*mysql.DatabaseMysql // mysql 子链接
var Template *template.CgoTemplateType         // 模板缓存文件
var Modules *dataModel.DM                      // 数据模型(主要)
var ModuleChilds map[string]*dataModel.DM      // 数据模型(子)
var Plugins = plugins.NewPlugins()             // 一些便捷插件

var RouterFilterKey = struct { // 拦截器的位置字段
	BeforeRouter string
	AfterRender  string
}{
	BeforeRouter: route.BEFORE_ROUTER,
	AfterRender:  route.AFTER_RENDER}

const (
	VERSION = "0.1.1"
)

var (
	comm   string
	daemon bool
)

func Run(confPath string, beforeStartEvents func()) {

	if len(confPath) < 1 {
		log.Println("功能初始化: 需要指定配置文件的路径!")
		os.Exit(0)
	}

	if !Config.Set(confPath) {
		log.Fatalln("功能初始化: Cgo配置文件	---	[ fail ]")
		os.Exit(0)
	} else {
		log.Println("功能初始化: Cgo配置文件	---	[ ok ]")
	}

	route.SetConfig(Config.Conf)

	argumentHandler()

	// 创建静默启动线程
	if daemon {
		createDaemon()
	}
	comm = "start"

	// 启动执行
	if comm == "start" {

		// 0. 日志系统初始化
		if Config.Conf.Log.Key {
			cgologer.New(Config.Conf.Log.Path, Config.Conf.Log.FileName, Config.Conf.Log.StopCutOff)
		} else {
			cgologer.New("", "", true)
		}

		// 2. 启动session 如果session 设置了
		if Config.Conf.Session.Key {
			route.SetSession(Session.New(&Config.Conf.Session))
		}

		// 3. mysql 初始化
		if Config.Conf.Mysql.Key {

			// 启动mysql
			MysqlDefault = cgoMysql.New(&Config.Conf.Mysql)

			// 初始化数据模型
			Modules = dataModel.New(MysqlDefault)
			for k, v := range MysqlDefault.Links {
				var tmp *dataModel.DM
				tmp = dataModel.New(v)
				if ModuleChilds == nil {
					ModuleChilds = make(map[string]*dataModel.DM)
				}
				ModuleChilds[k] = tmp
			}
		}

		// 4. redis 初始化
		if Config.Conf.Redis.Key {
			Redis = cgoRedis.New(&Config.Conf.Redis)
		}

		// 5. 检测静态路径
		if len(Config.Conf.Server.StaticRouter) > 0 && len(Config.Conf.Server.StaticPath) > 0 {
			Router.SetStaticPath(Config.Conf.Server.StaticRouter, Config.Conf.Server.StaticPath)
		}

		// 6. 初始化模板
		if len(Config.Conf.Server.TemplatePath) > 0 {
			Template = template.New(&Config.Conf)
		}

		// 7. 前置回调方法执行
		beforeStartEvents()
	}

	// 执行启动
	command.Run(&comm, Router, &Config.Conf)

}

// 命令行参数处理
func argumentHandler() {
	for _, v := range os.Args {
		if v == "start" && len(comm) < 1 {
			comm = "start"
		}
		if v == "stop" && len(comm) < 1 {
			comm = "stop"
		}
		if v == "restart" && len(comm) < 1 {
			comm = "restart"
		}
		if v == "-d" || v == "-domain" {
			daemon = true
		}
	}
}

// 静默启动
func createDaemon() {
	cmd := exec.Command(os.Args[0], comm)
	cmd.Start()
	daemon = false
	os.Exit(0)
}

// 初始化路由
func init() {
	// 初始化全局路由变量
	Router = route.NewRouter()

}
