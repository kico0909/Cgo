package app

/*
服务器的内核 用于启动服务器
*/

import (
	"net/http"
	"time"
	"strconv"
	"strings"
	"crypto/tls"

	"github.com/Cgo/kernel/config"
	"github.com/Cgo/route"

	// 如果此处报错,请 go get golang.org/x/net/http2 等包
	"golang.org/x/net/http2"
	"golang.org/x/crypto/acme/autocert"
	"log"
)


type RouterType interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func ServerStart(router *route.RouterManager, conf *config.ConfigData){
	// 不启用HTTPS
	if !conf.TLS.Key{
		normalServerStart(router,conf)
	}
	// 启用 HTTPS 并且 自动申请使用和续期let's Encrypt证书
	if conf.TLS.LetsEncrypt {
		httpsLetsServerStart(router,conf)
	}
	httpsNormalServerStart(router,conf)
}


// 非HTTPS服务器
func normalServerStart (router *route.RouterManager, conf *config.ConfigData) {

	server := &http.Server{
		// 地址及端口号
		Addr: `:`+strconv.FormatInt(conf.Server.Port, 10),

		// 读取超时时间
		ReadTimeout: conf.Server.ReadTimeout * time.Second,

		// 写入超时时间
		WriteTimeout: conf.Server.WriteTimeout * time.Second,

		// 头字节限制
		MaxHeaderBytes: conf.Server.MaxHeaderBytes * 1024,

		// 路由
		Handler: router,

	}

	log.Println("服务器启动完成: (监听端口:"+strconv.FormatInt(conf.Server.Port, 10)+") --- [ ok ]\n\n")

	log.Fatalln(server.ListenAndServe())


}

// 启动自动申请let's encrypt 证书的服务器
func httpsLetsServerStart(router *route.RouterManager, conf *config.ConfigData){
	https_domain := strings.Split(conf.TLS.LetsEncryptOpt.Domain, ",")

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(https_domain...), //your domain here
		Cache:      autocert.DirCache("certs"),     //folder for storing certificates
		Email:      conf.TLS.LetsEncryptOpt.Email,
	}
	// 80端口 301 重定向
	go http.ListenAndServe(":http", certManager.HTTPHandler(nil)) // 支持 http-01

	// server 配置
	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			NextProtos:     []string{http2.NextProtoTLS, "http/1.1"},
			MinVersion:     tls.VersionTLS12,
		},
		MaxHeaderBytes: 32<<20,
		// 路由
		Handler: router,
	}

	log.Println("服务器启动完成:(TLS) --- [ ok ]")
	log.Fatalln(server.ListenAndServeTLS("", ""))
}

// 启动https服务器,需要填写证书路径
func httpsNormalServerStart(router *route.RouterManager, conf *config.ConfigData){
	// 启用 HTTPS 直接加载证书
	server := &http.Server{

		// 地址及端口号
		Addr: `:`+strconv.FormatInt(conf.Server.Port, 10),

		// 读取超时时间
		ReadTimeout: conf.Server.ReadTimeout * time.Second,

		// 写入超时时间
		WriteTimeout: conf.Server.WriteTimeout * time.Second,

		// 头字节限制
		MaxHeaderBytes: conf.Server.MaxHeaderBytes * 1024,
		// 路由
		Handler: router,
	}

	log.Println("服务器启动完成:(https) --- [ ok ]")

	log.Fatalln(server.ListenAndServeTLS(conf.TLS.CertPath, conf.TLS.KeyPath))
}

