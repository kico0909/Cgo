package cas

import (
	"github.com/kico0909/cgo"
	"github.com/kico0909/cgo/core/kernel/logger"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"
)

type pubType struct {
	SessionName string `json:"sessionName"`
	CasReqReturn CasReqReturn `json:"casReqReturn"`
}

type casLogoutXMLtype struct {
	NameID       string `xml:"NameID"`
	SessionIndex string `xml:"SessionIndex"`
}

func New(sessionName string){
	pub.SessionName = sessionName
}

var pub pubType

// api 部分的cas验证器
func CasVerify_Api(h *cgo.RouterHandler) bool {

	casUrl := h.R.Header.Get("cas-url")

	// 未获得cas 的url
	if len(casUrl) < 1 {
		return true
	}
	// 检测session 验证登录 - 未登录去cas 登录
	s := h.Session.Get(pub.SessionName)

	// 没有session
	if s == nil {
		h.ShowForApiMode(false, "cas none login", nil)
		return true
	}

	return false
}

// 页面部分的cas验证器
func CasVerify_Page(h *cgo.RouterHandler) bool {
	casUrl := h.R.Header.Get("cas-url")

	// 未获得cas 的url
	if len(casUrl) < 1 {
		return true
	}

	// 检测session 验证登录 - 未登录去cas 登录
	s := h.Session.Get(pub.SessionName)

	// 没有session
	if s == nil {

		// 判断ticket
		vars := h.R.URL.Query()
		ticket, tok := vars["ticket"]

		// 无ticket 跳转cas登录
		if !tok {
			http.Redirect(h.W, h.R, casUrl+"?service="+getFullUrl(h.R), http.StatusFound)
			return true
		}

		if ticket[0] == "" {
			log.Error("未发现ticket变量长度")
			return true
		}

		uinfo, err := getUserBaseInfoForCas(casUrl, ticket[0], getFullUrl(h.R))

		if err != nil {
			log.Error("ticket 换取 用户信息失败[", err.Error(), "]")
			return true
		}

		// cas 请求成功, 返回失败登录信息
		if len(uinfo.ServiceResponse.AuthenticationFailure.Code) > 0 {

			// INVALID_TICKET 令牌失效的ticket,重新返回
			if uinfo.ServiceResponse.AuthenticationFailure.Code == "INVALID_TICKET" {
				http.Redirect(h.W, h.R, casUrl+"?service="+getFullUrl(h.R), http.StatusFound)
				return true
			}

			log.Error("[", uinfo.ServiceResponse.AuthenticationFailure.Description, " | ", uinfo.ServiceResponse.AuthenticationFailure.Code, "]")
			return true
		}

		// 赋值保存session
		var t pub.SessionType
		t.Ticket = ticket[0]
		t.User = uinfo.ServiceResponse.AuthenticationSuccess.Attributes
		if err := h.Session.Set(pub.SessionName, t); err != nil {
			log.Error("[session 保存失败]")
			return true
		}

		// 将session 保存到ticket sid 映射表中
		pub.UserSessionMap[ticket[0]] = h.Session.SessionID()
		return false
	}

	if len(s.(pub.SessionType).Ticket) < 1 {
		log.Error("cas验证已获得session,但session为空")
		return true
	}
	return false
}

// 用于cas通知登出消息的拦截器
func CasLogout(h *cgo.RouterHandler) bool {

	if strings.ToLower(h.R.Method) != "post" { // 不是POST 请求不执行拦截器
		return false
	}

	logoutValue := h.R.FormValue("logoutRequest")
	// 消息无返回值
	if len(logoutValue) < 1 {
		log.Println("<cas服务器通知登出> : [API接收API传值,传值无效或未获得传值, logoutRequest = '' ]")
		return false
	}

	var v casLogoutXMLtype
	xml.Unmarshal([]byte(logoutValue), &v)

	sid, ok := pub.UserSessionMap[v.SessionIndex]
	if !ok {
		log.Println("<cas服务器通知登出> : [ticket <-> sessionID 映射表获得sid 失败(", v.SessionIndex, ")]")
		return false
	}

	store, err := Cgo.Router.GetSessionManager().GetSessionStore(sid)
	if err != nil {
		log.Println("<cas服务器通知登出> : [获得session Store 失败(", sid, ")]")
		return false
	}

	if err := store.Flush(); err != nil {
		log.Println("<cas服务器通知登出> : [移除session失败(", sid, ")]")
		return false
	}

	return false
}

// ---------------------------------------------------------------------------------------------------------------------

// 获得当前页面的路径
func getFullUrl(r *http.Request) string {
	var part1 string
	var host string

	proto := r.Header.Get("X-Forwarded-Proto")
	if len(proto) < 1 {
		if r.TLS != nil {
			part1 = "https://"
		} else {
			part1 = "http://"
		}
	} else {
		// 如果是nginx等代理工具转发的请求, 添加: [ proxy_set_header    X-Forwarded-Proto    $scheme; ] 即可获得URL的协议
		part1 = proto + "://"
	}

	host = part1 + r.Host + r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		var q []string
		for k, v := range r.URL.Query() {
			if k != "ticket" {
				q = append(q, k+"="+v[0])
			}
		}
		if len(q) > 0 {
			host = host + "?" + strings.Join(q, "&")
		}
	}
	host = pub.EncodeURIComponent(host)

	return host
}

// 向cas服务器发送ticket, 获得数据
func getUserBaseInfoForCas(casUrl, ticket, service string) (CasReqReturn, error) {
	var v CasReqReturn
	var url = casUrl + "/serviceValidate?format=JSON&ticket=" + ticket + "&service=" + service

	res, err := http.Get(url)
	if err != nil {
		return v, err
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return v, err
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return v, err
	}

	return v, nil

}
