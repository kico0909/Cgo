package Cgo

import (
	log2 "github.com/Cgo/kernel/logger"
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
	"github.com/Cgo/kernel/logger"
)


var Config config.ConfigModule			// 配置
var Router *route.RouterManager			// 路由
var Session *session.CgoSession			// session
var Redis *reids.DatabaseRedis			// redis
var Mysql *mysql.DatabaseMysql			// mysql TODO 后期改成数据模型的封装
var Template *template.CgoTemplateType	// 模板缓存文件
var Cas *cas.CasFilter					// cas 方法
var Modules *module.DataModlues				// 数据模型
var Log *log2.Logger

type RouterHandler = route.RouterHandler
type TableModule = module.TableModule
type CasUserinfoType = cas.CasReqReturn

var (
	comm string
	daemon bool
)

func Run(confPath string, beforeStartEvents func()){

	if len(confPath)<1 {
		log.Fatalln("功能初始化失败: 需要指定配置文件的路径!")
	}

	Config.Set(confPath)

	argumentHandler()

	// 创建静默启动线程
	if daemon {
		createDaemon()
	}

	// 启动执行
	if comm == "start" {

		// 0. 日志系统初始化
		if Config.Conf.Log.Key {
			Log = log2.New(Config.Conf.Log.Path, Config.Conf.Log.FileName, Config.Conf.Log.AutoCutOff)
		}else{
			Log = log2.New("", "", false)
		}

		// 2. 启动session 如果session 设置了
		if Config.Conf.Session.Key {
			Session = Session.New(&Config.Conf.Session)
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
		if len(Config.Conf.Server.StaticRouter) > 0 && len(Config.Conf.Server.StaticPath)> 0 {
			Router.SetStaticPath(Config.Conf.Server.StaticRouter, Config.Conf.Server.StaticPath)
		}

		// 6. 尝试初始化cas
		if Config.Conf.Cas.Key {
			Cas = cas.NewCas(	Config.Conf.Cas.Url,
								Config.Conf.Cas.SessionName,
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



		// 8. 前置回调方法执行
		beforeStartEvents()
	}

	// 执行启动
	command.Run(&comm, Router, &Config.Conf)

}

// 命令行参数处理
func argumentHandler(){
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
		if v == "-d" || v == "-domain" {
			daemon = true
		}
	}
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

