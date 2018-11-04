package command

import (
	cgoApp "github.com/Cgo/kernel/app"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/route"
	"github.com/Cgo/funcs"
	"os/exec"
	"os"
	log "github.com/sirupsen/logrus"
	"strconv"
	)

const infoPath string = "./pid.txt"


// 服务器参数处理
func Run(comm *string, router *route.RouterManager, conf *config.ConfigData){


	switch *comm {

		case "start":
			saveStartInfos(strconv.FormatInt(int64(os.Getpid()), 10))
			serverStart(router,conf)
			break

		case "stop":
			serverStop(loadStartInfos())
			break

		default:

	}
}

// 服务器初始化与启动
func serverStart( router *route.RouterManager, conf *config.ConfigData){

	// 初始化模板
	//template.Init(conf)

	// 启动服务器
	cgoApp.ServerStart(router, conf)

}

// 服务器停止
func serverStop(pid string){

	var as []string = []string{"-9",pid}

	cmd := exec.Command("kill", as...)

	if cmd.Start() != nil {
		log.Fatal("关闭服务执行失败!")
	}else{
		log.Info("pid:[ " + pid + " ]进程被移除")
	}

	cmd = exec.Command("rm", "-rf", infoPath)
	cmd.Start()

}

// 记录启动信息PID文件
func saveStartInfos(pid string){
	f, err := os.OpenFile(infoPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, errs := f.WriteString(pid)
	if errs != nil {
		log.Fatal(errs)
	}
}

// 读取启动PID信息文件
func loadStartInfos()string{

	cont, err := funcs.ReadFile(infoPath)

	if err != nil {
		log.Fatalln("PID记录文件无法读取, 请手动结束应用!")
	}

	return string(cont)
}

func init () {



}