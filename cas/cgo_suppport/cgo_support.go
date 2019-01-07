package cgo_suppport

import (
	"github.com/Cgo/cas"
	"github.com/Cgo/kernel/session"
	beegoSession "github.com/astaxie/beego/session"
	"github.com/Cgo/kernel/logger"
	"net/http"
)

// cas 的 Cgo 对接

type cgoSession4cas struct {	//
	sessionManager *session.CgoSession
	ss beegoSession.Store
}

// 启动本次session
func (_self *cgoSession4cas) Start (w http.ResponseWriter, r *http.Request) {
	tmp, err := _self.sessionManager.SessionStart(w,r)
	if err != nil {
		log.Println("cas -o session 的  session store 获得错误")
	}
	_self.ss = tmp
}

// 停止本次session
func (_self *cgoSession4cas) End (w http.ResponseWriter, r *http.Request) {
	_self.ss.SessionRelease(w)
}

// 设置sesion manager
func (_self *cgoSession4cas) Set (sess interface{}) {
	_self.sessionManager = sess.(*session.CgoSession)
}

// 获得指定session
func (_self *cgoSession4cas) GetSessionInfo (key string)interface{} {
	return _self.ss.Get(key)
}

// 设置指定session
func (_self *cgoSession4cas) SetSessionInfo (key string, value interface{}) error {
	return _self.ss.Set(key, value)
}

// 删除指定session
func (_self *cgoSession4cas) DeleteSession (sid string) error {
	tmp, err := _self.sessionManager.GetSessionStore(sid)
	if err != nil {
		log.Println("cas cgo session manager 查找指定session id 的 store 失败!")
		return err
	}
	log.Println(tmp)
	return tmp.Flush()
}

// 获得当前session 的sid
func (_self *cgoSession4cas) GetSid () string {
	return _self.ss.SessionID()
}

func init(){
	cas.InstallSessionManager(&cgoSession4cas{})
}