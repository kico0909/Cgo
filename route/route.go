package route

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/kernel/logger"
	"github.com/Cgo/kernel/session"
	"github.com/Cgo/route/defaultPages"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	Method_Error          = errors.New("路由访问方式错误!")
	RegExp_Url_Set_String = "{[a-z|A-Z|0-9|_]*}"
	RegExp_Url_Set, _     = regexp.Compile(RegExp_Url_Set_String)
	RegExp_Url_String     = "[a-z|A-Z|0-9|_|.]*"
	RegExp_Url, _         = regexp.Compile("[a-z|A-Z|0-9|_|.]*")
	defaultApiCode        = defaultApiCodeType{200, 400}
)

// 配置出session
var sess *session.CgoSession

func SetSession(s *session.CgoSession) {
	sess = s
}

// 获得全局的config
var conf config.ConfigData

func SetConfig(c config.ConfigData) {
	conf = c
}

// 接口实现方法
func (_self *RouterManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 禁止方法继续执行
	var stopKey bool

	routerHandlerValue := &RouterHandler{
		W:       w,
		R:       r,
		path:    "",
		Vars:    make(map[string]string),
		Session: nil,
		Values:  make(map[string]interface{})}

	// Session 当前全局
	if conf.Session.Key {
		routerHandlerValue.Session, _ = sess.SessionStart(w, r)
	}

	// 检测路由前的拦截器
	for _, v := range _self.filter.beforeRoute {
		if v.rule.MatchString(r.URL.Path) { // 正则匹配
			tmp := (*(v.f))(routerHandlerValue)
			if !stopKey {
				stopKey = tmp
			}
			if v.BlockNext {
				break
			}
		}
	}

	// 拦截器是否阻塞执行
	if stopKey {
		return
	}

	for _, v := range _self.Routers {

		// 1. 匹配路由
		if checkRouter(r.URL.Path, v.regPath) {

			// 2. 匹配Method
			if v.Methods == nil || v.Methods[r.Method] {
				routerHandlerValue.path = v.path
				for k, v := range _self.getRouterValue(r.URL.Path, v) {
					routerHandlerValue.Vars[k] = v
				}
				v.H = routerHandlerValue

				// 3. 执行业务视图渲染
				v.viewRenderFunc(v.H)

				// 4. session 保存
				if v.H.Session != nil {
					v.H.Session.SessionRelease(v.H.W)
				}

				// 路由执行渲染后的拦截器
				for _, vv := range _self.filter.afterRender {
					if vv.rule.MatchString(r.URL.Path) { // 正则匹配
						(*(vv.f))(v.H)
						if vv.BlockNext {
							break
						}
					}
				}

			} else {

				// 访问模式拒绝页面
				switch strings.ToLower(r.Method) {

				case "get":
					defaultPages.Page_405_get(w)
					break

				case "post", "put", "delete":
					defaultPages.Page_405_post(w)
					break
				}

			}
			return
		}
	}
	// 没有路由情况下跳404
	// 访问模式拒绝页面
	switch strings.ToLower(r.Method) {
	case "get":
		defaultPages.Page_404_get(w)
		break
	case "post":
		defaultPages.Page_404_post(w)
		break
	}
}

// 注册一条新路由
func (_self *RouterManager) Register(path string, f routerHandlerFunc, childRouter ...routerGroup) *routerChip {
	if path == _self.staticRouter {
		log.Fatalln("路由地址与静态文件路由地址冲突!\n[", path, "==>", _self.staticPath, "]")
	}
	return _self.addRouter(path, f)
}

// 针对路由的拦截器
// 参数: 拦截器位置, 拦截的路由, 拦截器执行方法(需要返回Bool 是否拦截), 被拦截后是否继续执行拦截器
func (_self *RouterManager) InsertFilter(position string, pathRule string, f func(handler *RouterHandler) bool, BlockNext ...bool) {
	re, err := regexp.Compile(handlerPathString2regexp(pathRule))
	if err != nil {
		log.Println("过滤器:[", position, "]路径匹配错误[", err, "]!")
		re, _ = regexp.Compile(`[\D|\d]*`)
	}

	filterStruct := &filter{pathRule, re, &f, len(BlockNext) > 0 && BlockNext[0] == true}

	switch position {

	case "beforeRoute":
		_self.filter.beforeRoute = append(_self.filter.beforeRoute, filterStruct)
		break

	case "afterRender":
		_self.filter.afterRender = append(_self.filter.afterRender, filterStruct)
		break

	}

}

// 设置静态文件访问目录
func (_self *RouterManager) SetStaticPath(router string, path string) {

	// 检测是否有重复路由
	for _, v := range _self.Routers {
		if v.path == router {
			log.Fatalln("设置的静态文件地址路由路径冲突!")
		}
	}
	_self.staticRouter = router
	_self.staticPath = path

	// 注册一个静态文件的路由
	_self.addRouter(handlerPathString2regexp(router+"**"), _self.makeFileServe(http.StripPrefix(router, http.FileServer(http.Dir(_self.staticPath)))))

}

// 设置静态文件访问目录
func (_self *RouterManager) SetDefaultApiCode(success, fail int64) {
	defaultApiCode.Success = success
	defaultApiCode.Fail = fail
}

// 设置路由组 TODO 未完成的事业
func (_self *RouterManager) Group(path string, f routerHandlerFunc, childRouter ...routerGroup) *routerChip {

	return _self.addRouter(path, func(h *RouterHandler) {
		f(h)
		//for _,v := range childRouter{
		//
		//}
	})
}

// 注册一条新路由
func (_self *RouterManager) addRouter(path string, f routerHandlerFunc) *routerChip {
	rp, is := handlerPath(path)
	innerH := &RouterHandler{path: path}
	tmp := &routerChip{path: path, regPath: rp, viewRenderFunc: f, IsRouterValue: is, H: innerH}
	_self.Routers = append(_self.Routers, tmp)
	return tmp
}

// 获得路由解析传值
func (_self *RouterManager) getRouterValue(url string, rh *routerChip) map[string]string {

	res := make(map[string]string)
	reg, _ := regexp.Compile("[{|}]")

	// 不是路由传值
	if !rh.IsRouterValue {
		return res
	}

	routeSet := strings.Split(rh.path, "/")

	urlSet := strings.Split(url, "/")

	for k, v := range routeSet {
		// 确认是路由上设置的变量
		if RegExp_Url_Set.MatchString(v) {
			res[reg.ReplaceAllString(v, "")] = urlSet[k]
		}
	}

	return res
}

// 文件路由的处理方法
func (_self *RouterManager) makeFileServe(handler http.Handler) routerHandlerFunc {

	return func(h *RouterHandler) {

		handler.ServeHTTP(h.W, h.R)

	}

}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 设置路由的Method
func (_self *routerChip) Method(methods ...string) *routerChip {
	_self.Methods = make(map[string]bool)
	for _, v := range methods {
		_self.Methods[strings.ToUpper(v)] = true
	}
	return _self
}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 路由传值的原型链
func (r *RouterHandler) Show(str string) {
	fmt.Fprintf(r.W, str)
}

// 获得json类型的body传值
func (r *RouterHandler) GetBodyValueToJson(res interface{}) {
	defer r.R.Body.Close()
	b, err := ioutil.ReadAll(r.R.Body)
	if err != nil {
		return
	}
	json.Unmarshal(b, res)
}

// API形式的json数据渲染页面(用于API的返回)
type showForApiModeType struct {
	Code    int64       `json:"code"`
	Success bool        `json:"success"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data"`
}

func (r *RouterHandler) ShowForApiMode(success bool, err, data interface{}, code ...int64) {
	var result showForApiModeType
	result.Success = success
	if result.Success {
		result.Code = defaultApiCode.Success
	} else {
		result.Code = defaultApiCode.Fail
		if len(code) > 0 {
			result.Code = code[0]
		}
	}
	result.Message = err
	result.Data = data
	strByte, _ := json.Marshal(result)
	fmt.Fprintf(r.W, string(strByte))
}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----
// 创建新的路由
func NewRouter() *RouterManager {
	return &RouterManager{}
}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 把一个路由设置的URL转换成用于判断URL的正则
func handlerPath(path string) (string, bool) {
	// 创建一个正则
	return "^" + RegExp_Url_Set.ReplaceAllLiteralString(path, RegExp_Url_String) + "$", !(RegExp_Url_Set.FindIndex([]byte(path)) == nil)
}

// 路由匹配
func checkRouter(url, path string) bool {
	re, _ := regexp.Compile(path)
	return re.MatchString(url)
}

// 解析路由路径为正则字符串
func handlerPathString2regexp(path string) string {
	temp_time := strconv.FormatInt(time.Now().Unix(), 10)
	// 双星
	doubleStartReg, _ := regexp.Compile(`\*\*`)
	// 单星
	startReg, _ := regexp.Compile(`\*`)
	// 双星转换的中间量
	swapReg, _ := regexp.Compile(temp_time)

	// 替换双星
	path = doubleStartReg.ReplaceAllString(path, `\S`+temp_time)

	// 替换单星
	path = "^" + startReg.ReplaceAllString(path, `[a-z|A-Z|0-9|_|.]*`) + "$"

	// 替换中间变量
	path = swapReg.ReplaceAllString(path, `*`)

	return path

}
