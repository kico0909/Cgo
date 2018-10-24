package session

import (
	_ "github.com/astaxie/beego/session/redis"
	beegoSession "github.com/astaxie/beego/session"
	"log"
	"github.com/Cgo/kernel/config"
	"net/http"
	)

// Cgo的session 封装 TODO 把beego的session做了二次封装
type CgoSession struct {
	manager *beegoSession.Manager
}

var (
	sessionEndName = "_glsessn_"
	sessionSetup beegoSession.ManagerConfig
)

var sessionManager CgoSession

// 新建
func (_self *CgoSession) New(conf *config.ConfigData)*CgoSession{

	var result *CgoSession
	var err error

	if conf.Server.IsStatic && !conf.Session.Key {
		return result
	}

	// 配置信息检测容错设置默认值
	if conf.Session.SessionType == "" {
		conf.Session.SessionType = "memory"
	}

	if conf.Session.SessionName == "" {
		conf.Session.SessionName = "_CHUNK"
	}

	if conf.Session.SessionLifeTime == 0 {
		conf.Session.SessionLifeTime = 3600
	}

	sessionSetup.CookieName = conf.Session.SessionName + sessionEndName
	sessionSetup.Gclifetime = conf.Session.SessionLifeTime
	sessionSetup.EnableSetCookie = true

	// 初始化 session
	switch conf.Session.SessionType{

	case "redis":
		srHost := conf.Session.SessionRedis.Host
		srPort := conf.Session.SessionRedis.Port
		srNumber := conf.Session.SessionRedis.Dbname
		srPassword := conf.Session.SessionRedis.Password
		sessionSetup.ProviderConfig = srHost+`:`+srPort+`,`+srNumber+`,`+srPassword
		break

	default:

	}

	log.Println("初始化SESSION [",conf.Session.SessionType,"]类型")

	sessionManager.manager, err = beegoSession.NewManager( conf.Session.SessionType, &sessionSetup )

	if err != nil {
		log.Println(333,err)
	}

	go sessionManager.manager.GC()

	return &sessionManager
}

// 启动session
func (_self *CgoSession) SessionStart(w http.ResponseWriter, r *http.Request) (beegoSession.Store, error) {
	return _self.manager.SessionStart(w,r)
}

// 根据id 获得
func (_self *CgoSession) GetSessionStore(sid string) (beegoSession.Store, error) {
	return _self.GetSessionStore(sid)
}

// 销毁全部
func (_self *CgoSession) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	_self.manager.SessionDestroy(w,r)
}


func (_self *CgoSession) SessionRegenerateID(w http.ResponseWriter, r *http.Request) (beegoSession.Store) {
	return _self.manager.SessionRegenerateID(w,r)
}

func (_self *CgoSession) GetActiveSession() int {
	return _self.manager.GetActiveSession()
}

func (_self *CgoSession) SetSecure(secure bool) {
	_self.manager.SetSecure(secure)
}


