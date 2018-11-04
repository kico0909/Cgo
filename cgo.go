package Cgo

import (
	"github.com/Cgo/mysql"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/route"
	"github.com/Cgo/kernel/session"
	"github.com/Cgo/kernel/command"
	cgoMysql "github.com/Cgo/kernel/mysql"
	cgoRedis "github.com/Cgo/kernel/redis"
	"github.com/Cgo/kernel/module"
	"os"
	"os/exec"
	"github.com/Cgo/redis"
	"github.com/Cgo/cas"
	_ "github.com/Cgo/cas/cgo_suppport"
	"github.com/Cgo/kernel/template"
	)


var Config config.ConfigModule			// 配置
var Router *route.RouterManager			// 路由
var Session *session.CgoSession			// session
var Redis *reids.DatabaseRedis			// redis
var Mysql *mysql.DatabaseMysql			// mysql
var Template *template.CgoTemplateType	// 模板缓存文件
var Cas *cas.CasFilter					// cas 方法

type Module = module.DBModel
type RouterHandler = route.RouterHandler

var (
	comm string
	daemon bool
)

func Run(){

	// 1. 检测启动参数是否正确
	comm, daemon = argumentHandler()

	// 创建静默启动线程
	if daemon {
		createDaemon()
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

		// 5. 检测静态路径
		if len(Config.Conf.Server.StaticRouter) > 0 && len(Config.Conf.Server.StaticPath)> 0 {
			Router.SetStaticPath(Config.Conf.Server.StaticRouter, Config.Conf.Server.StaticPath)
		}

		// 6. 尝试初始化cas
		if Config.Conf.Cas.Key {
			Cas = cas.NewCas(	Config.Conf.Cas.Url,
								Config.Conf.Cas.CasSessionName,
								Config.Conf.Cas.APIPath,
								Config.Conf.Cas.LogoutRouter,
								Config.Conf.Cas.LogoutRequestMethod,
								Config.Conf.Cas.LogoutReUrl,
								Config.Conf.Cas.LogoutValueName,
								Config.Conf.Cas.APIErrCode,
								Config.Conf.Cas.WhiteList,
								Session)

			// cas 页面拦截器
			Router.InsertFilter("beforeRoute", "/**", Cas.NewCasFilter())

		}

		// 7. 初始化模板
		if len(Config.Conf.Server.TemplatePath)>0 {
			Template = template.New(&Config.Conf)
		}
	}

	// 执行启动
	command.Run(&comm, Router, &Config.Conf)

}

// 命令行参数处理
func argumentHandler()(comm string, daemon bool){

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

	return comm,daemon

}

// 静默启动
func createDaemon(){
	cmd := exec.Command(os.Args[0], comm)
	cmd.Start()
	daemon = false
	os.Exit(0)
}

// 初始化路由
func init(){
	// 初始化全局路由变量
	Router = route.NewRouter()
}

