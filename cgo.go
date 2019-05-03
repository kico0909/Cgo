package Cgo

import (
	"github.com/Cgo/cas"
	_ "github.com/Cgo/cas/cgo_suppport"
	"github.com/Cgo/kernel/command"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/kernel/logger"
	cgologer "github.com/Cgo/kernel/logger"
	"github.com/Cgo/kernel/module"
	cgoMysql "github.com/Cgo/kernel/mysql"
	cgoRedis "github.com/Cgo/kernel/redis"
	"github.com/Cgo/kernel/session"
	"github.com/Cgo/kernel/template"
	"github.com/Cgo/mysql"
	"github.com/Cgo/redis"
	"github.com/Cgo/route"
	"os"
	"os/exec"
)

type RouterHandler = route.RouterHandler
type TableModule = module.TableModule
type CasUserinfoType = cas.CasReqReturn

var Config config.ConfigModule         // 配置
var Router *route.RouterManager        // 路由
var Session *session.CgoSession        // session
var Redis *reids.DatabaseRedis         // redis
var Mysql *mysql.DatabaseMysql         // mysql TODO 后期改成数据模型的封装
var Template *template.CgoTemplateType // 模板缓存文件
var Cas *cas.CasFilter                 // cas 方法
var Modules *module.DataModlues        // 数据模型
var Log *cgologer.Logger               // 可输出到文件的日志类

var RouterFilterKey = struct { // 拦截器的位置字段
	BeforeRouter string
	AfterRender  string
}{
	BeforeRouter: route.BEFORE_ROUTER,
	AfterRender:  route.AFTER_RENDER}

const (
	VERSION = "1.0"
)

var (
	comm   string
	daemon bool
)

func Run(confPath string, beforeStartEvents func()) {

	if len(confPath) < 1 {
		log.Fatalln("功能初始化: 需要指定配置文件的路径!")
	}

	if !Config.Set(confPath) {
		log.Fatalln("功能初始化: Cgo配置文件	---	[ fail ]")
	} else {
		log.Println("功能初始化: Cgo配置文件	---	[ ok ]")
	}

	route.SetConfig(Config.Conf)

	argumentHandler()

	// 创建静默启动线程
	if daemon {
		createDaemon()
	}

	// 启动执行
	if comm == "start" {

		// 0. 日志系统初始化
		if Config.Conf.Log.Key {
			Log = cgologer.New(Config.Conf.Log.Path, Config.Conf.Log.FileName, Config.Conf.Log.AutoCutOff)
		} else {
			Log = cgologer.New("", "", false)
		}

		// 2. 启动session 如果session 设置了
		if Config.Conf.Session.Key {
			route.SetSession(Session.New(&Config.Conf.Session))
		}

		// 3. mysql 初始化
		if Config.Conf.Mysql.Key {
			// 启动mysql
			Mysql = cgoMysql.New(&Config.Conf.Mysql)
			// 初始化数据模型
			Modules = module.New(Mysql)
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
