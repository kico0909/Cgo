package cas

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"net/url"
	"time"
	"strings"
	"regexp"
	"reflect"
	"encoding/xml"
	"github.com/Cgo/route"
	"strconv"
	"fmt"
	"github.com/Cgo/funcs"
	"github.com/Cgo/kernel/logger"
)

type CasSessionFunc interface {
	Start(http.ResponseWriter, *http.Request)	// 初始化 session store
	End(http.ResponseWriter, *http.Request)		// 结束 session store
	Set(interface{})							// 设置 session manager
	GetSessionInfo( string )interface{}			// 获得用户的 session
	SetSessionInfo( string, interface{}) error	// 为用户设置 session
	DeleteSession( string ) error				// 删除指定sid 对应的session
	GetSid() string								// 获得当前session 的sid
}

// CAS拦截器
type CasFilter struct {

	// CAS 用的一些基础信息
	casUrl 				string				// cas 服务器地址
	mappingTicketSess	map[string]string	// ticket 和 session id 的映射

	// Session 部分的设置
	SessionInfoName 	string				// 用于保存session的名称
	sess 				*CasSessionFunc		// session 管理器的接口
	sessionManager		interface{}

	// Api路由的相关设置
	apiRouter 			*regexp.Regexp		// 用于定义api走的路由
	apiErrCode			string				// 当时API的路由时,验证不通过返回的code码

	// 登出部分的设置
	logoutRequestRouter *regexp.Regexp		// 定义登出消息的接收路由
	logoutMethod		string				// cas 登出消息推送的method
	logoutReUrl			string				// 登出时跳转回的路由
	logoutValueName		string				// 登出时跳转回的路由

	// 检测的路由
	casCheckRouter 			*regexp.Regexp		// 除api 和 白名单外 还需要经由cas验证的路由正则(默认全部路由)
	whiteList 				[]*regexp.Regexp	// Cas的白名单

}

type  casLogoutXMLtype struct {
	NameID string `xml:"NameID"`
	SessionIndex string `xml:"SessionIndex"`
}

var sessionApi *CasSessionFunc

// Cas ticket 和 用户sid 之间的关系
var casTicket2Sessionid map[string]string

// 初始化cas
func NewCas(casUrl, sessionName, apiRouter,logoutRouter, logoutMethod, logoutReUrl, logoutValueName, apiErrCode string, wlist []string, sm interface{})*CasFilter{

	log.Println("功能初始化: CAS验证 --- [ ok ]")

	return &CasFilter{	SessionInfoName: sessionName,
						apiRouter: string4regexp(apiRouter),
						apiErrCode: apiErrCode,
						sessionManager: sm,
						logoutReUrl: logoutReUrl,
						logoutValueName: logoutValueName,

						logoutRequestRouter: string4regexp(logoutRouter),
						logoutMethod: strings.ToLower(logoutMethod),
						sess: sessionApi,
						casUrl: casUrl,
						casCheckRouter: string4regexp("/**"),
						whiteList: handlerWhiteList(wlist),
						mappingTicketSess: make(map[string]string)}
}

// 设置session 管理器
func InstallSessionManager (sessionManager CasSessionFunc){
	sessionApi = &sessionManager
}

//  创建cas拦截器方法
func (_self *CasFilter) NewCasFilter ()(func( *route.RouterHandler)(bool)) {
	(*_self.sess).Set(_self.sessionManager)
	return func(h *route.RouterHandler)(bool){
		return _self.serveHTTP(h.W, h.R)
	}
}

// 接口的实现
func (_self *CasFilter) serveHTTP (w http.ResponseWriter, r *http.Request)(bool){

	// 判断是否是cas服务器的logout 的请求 router
	if _self.logoutRequestRouter.MatchString(r.URL.Path) {
		if r.Method == "POST" {
			r.ParseForm()
			casLogout := r.FormValue(_self.logoutValueName)
			if len(casLogout)>0{
				var tmp casLogoutXMLtype
				xml.Unmarshal([]byte(casLogout), &tmp)
				if casLogout != "" {

					// 清除session
					_self.clearUserSession(w,r,tmp.SessionIndex)

					// 不继续执行路由
					return true
				}
			}
		}
	}

	// 符合白名单 直接执行
	for _, reg:= range _self.whiteList {
		if reg.MatchString(r.URL.Path) {
			return false
		}
	}

	// 验证 用户是否在cas 已经登录
	if ok := _self.verifyLogin(w,r); ok {
		return false
	}

	// 后边均是未登录的处理

	// 判断是否是api路由
	if _self.apiRouter.MatchString(r.URL.Path) {
		code,_ := strconv.ParseInt(_self.apiErrCode, 10, 64)
		// 返回未登录的结果
		fmt.Fprintf( w, funcs.CreateHttpRequestResult(false, nil, nil ,  int(code)))
		return true
	}else{
		// 重定向到CAS 页面
		_self.ToLogin(w,r)
		return true
	}
}

// 验证是否登录 1. 判断session 是否有用户信息, 且用户信息未过期, 2. 是否有TICKET
func (_self *CasFilter) verifyLogin (w http.ResponseWriter, r *http.Request)bool{

	r.ParseForm()
	ticket := r.FormValue("ticket")

	// 启动session
	(*_self.sess).Start(w,r)

	// 获得用户在session 中的cas 信息
	userInfo := (*_self.sess).GetSessionInfo(_self.SessionInfoName)

	// 没有session
	if userInfo == nil {
		// ticket 存在 ==> 去验证获得用户信息
		if len(ticket) >0 {
			uf, err :=  _self.getUserInfo4CasServer(w, r, ticket)
			if err != nil {
				log.Println("通过Ticket获得用户信息失败!==>", err)
			}

			// 判断 ticket 请求成功
			if  len(uf.ServiceResponse.AuthenticationFailure.Code) < 1 {

				_self.mappingTicketSess[ticket] = (*_self.sess).GetSid()

				err = nil
				if err := (*_self.sess).SetSessionInfo(_self.SessionInfoName, uf); err != nil {
					log.Println("在session 设置用户登录信息失败!==>", err)
				}

				return true
			}

			return false
		}else{	// ticket 不存在 返回未登录
			return false
		}
	}else{	// 有session , 不存在过期问题, 验证通过, 不阻塞路由向下执行 所以返回false
		return true
	}
}

// 白名单路由字符串正则化
func handlerWhiteList(urls []string) (result []*regexp.Regexp){
	for _,v := range urls {
		result = append(result, string4regexp(v))
	}
	return result
}

// 路由字符串转正则
func string4regexp ( str string) *regexp.Regexp {
	temp_time := strconv.FormatInt(time.Now().Unix(), 10)
	// 双星
	doubleStartReg, _ := regexp.Compile(`\*\*`)
	// 单星
	startReg, _ := regexp.Compile(`\*`)
	// 双星转换的中间量
	swapReg, _ := regexp.Compile(temp_time)

	// 替换双星
	str = doubleStartReg.ReplaceAllString(str, `\S`+temp_time, )

	// 替换单星
	str = "^" + startReg.ReplaceAllString(str, `[a-z|A-Z|0-9|_|.]*`, ) + "$"

	// 替换中间变量
	str = swapReg.ReplaceAllString(str, `*`)

	reg, err := regexp.Compile(str)
	if err != nil {
		log.Println("Router字符串转正则错误,返回默认转换正则", str, err)
		reg, _ := regexp.Compile("^$")
		return reg
	}
	return reg
}

// 通过TICKET获得用户验证信息
func (_self *CasFilter) getUserInfo4CasServer(w http.ResponseWriter, r *http.Request, ticket string)(CasReqReturn, error){
	var ret CasReqReturn

	url := _self.casUrl + "/serviceValidate?ticket=" + r.FormValue("ticket") + "&service=" + getFullUrl(r,true) + "&format=JSON"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ret, err
	}
	reqRes, err := client.Do(req)
	if err != nil {
		return ret, err
	}
	err = nil

	defer reqRes.Body.Close()
	body, err := ioutil.ReadAll(reqRes.Body)
	err = json.Unmarshal(body, &ret)
	return ret, err

}

// 通过CAS 去登录, 302 跳转到cas 服务器
func (_self *CasFilter) ToLogin(w http.ResponseWriter, r *http.Request)  {
	http.Redirect(w,r,_self.casUrl+"?service=" + getFullUrl(r, true), http.StatusFound)
}

// 登出
func (_self *CasFilter) ToLogout(w http.ResponseWriter, r *http.Request){
	(*_self.sess).Start(w,r)
	uf := (*_self.sess).GetSessionInfo(_self.SessionInfoName)
	(*_self.sess).DeleteSession(uf.(CasReqReturn).ServiceResponse.AuthenticationSuccess.Ticket)
	(*_self.sess).End(w,r)
	http.Redirect(w,r,_self.casUrl+"/logout?service="+getFullUrl(r, _self.logoutReUrl), http.StatusFound)
}


// 根据通知解除登录
func (_self *CasFilter) clearUserSession (w http.ResponseWriter, r *http.Request, ticket string) bool {

	(*_self.sess).Start(w,r)

	sid := _self.mappingTicketSess[ticket]

	if sid == "" {
		log.Println("sessionID映射获得错误=>",ticket)
		return false
	}

	err := (*_self.sess).DeleteSession(sid)

	if err != nil {
		log.Println("session登出错误=>",err)
	}

	return true
}

// 获得当前页面的路径
func getFullUrl(r *http.Request, query interface{})string{
	var part1 string
	var host string

	proto := r.Header.Get("X-Forwarded-Proto")
	if len(proto) < 1 {
		if r.TLS != nil {
			part1 = "https://"
		}else{
			part1 = "http://"
		}
	}else{
		// 如果是nginx等代理工具转发的请求, 添加: [ proxy_set_header    X-Forwarded-Proto    $scheme; ] 即可获得URL的协议
		part1 = proto + "://"
	}

	if (reflect.TypeOf(query)).Name() == "string"{
		host = url.QueryEscape(part1 + r.Host + query.(string))
	}else{
		host = url.QueryEscape(part1 + r.Host + r.URL.Path)
	}

	if (reflect.TypeOf(query)).Name() == "bool"{
		if query.(bool) {
			if len(r.URL.RawQuery) >0 {
				vals := strings.Split(r.URL.RawQuery,"&")
				var query []string
				for _,v := range vals {
					if k,_ := regexp.MatchString(`^ticket=[\S|\D]*`, v); !k {
						query = append(query,v)
					}
				}
				if tmp := strings.Join(query,"&"); len(tmp)>0{
					host += "?" + tmp
				}

			}
		}
	}

	return host
}



