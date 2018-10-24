package template

import (
	"log"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/funcs"
)

func Init(conf *config.ConfigData){
	// 缓存模板 - 启动立即进行缓存
	if !conf.Server.IsStatic {
		log.Println("初始化 [ 模板缓存 ] ...")

		//basePath,err := funcs.GetMyPath()

		//if err != nil {
		//	log.Fatal(err)
		//}

		log.Println(conf.Server)

		funcs.CacheHtmlTemplate(conf.Server.TemplatePath)
	}
}