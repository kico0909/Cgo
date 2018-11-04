package session

import (
	_ "github.com/astaxie/beego/session/redis"
	beegoSession "github.com/astaxie/beego/session"
	"github.com/Cgo/kernel/config"
	"net/http"
	"log"
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
func (_self *CgoSession) New(conf *config.ConfigSessionOptions)*CgoSession{

	var err error

	// 配置信息检测容错设置默认值
	if conf.SessionType == "" {
		conf.SessionType = "memory"
	}

	if conf.SessionName == "" {
		conf.SessionName = "_CHUNK"
	}

	if conf.SessionLifeTime == 0 {
		conf.SessionLifeTime = 3600
	}

	sessionSetup.CookieName = conf.SessionName + sessionEndName
	sessionSetup.Gclifetime = conf.SessionLifeTime
	sessionSetup.EnableSetCookie = true

	// 初始化 session
	switch conf.SessionType{

	case "redis":
		srHost := conf.SessionRedis.Host
		srPort := conf.SessionRedis.Port
		srNumber := conf.SessionRedis.Dbname
		srPassword := conf.SessionRedis.Password
		sessionSetup.ProviderConfig = srHost+`:`+srPort+`,`+srNumber+`,`+srPassword
		break

	default:

	}

	log.Println("功能初始化: SESSION("+conf.SessionType+")	 --- [ ok ]")

	sessionManager.manager, err = beegoSession.NewManager( conf.SessionType, &sessionSetup )

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
	return _self.manager.GetSessionStore(sid)
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


