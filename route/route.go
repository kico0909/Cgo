package route

import (
	"net/http"
	"regexp"
	"strings"
	"errors"
	"github.com/Cgo/route/defaultPages"
	"github.com/Cgo/kernel/logger"
	"time"
	"strconv"
	)

var (
	Method_Error = errors.New("路由访问方式错误!")
	RegExp_Url_Set_String = "{[a-z|A-Z|0-9|_]*}"
	RegExp_Url_Set, _ = regexp.Compile(RegExp_Url_Set_String)
	RegExp_Url_String = "[a-z|A-Z|0-9|_|.]*"
	RegExp_Url,_ = regexp.Compile("[a-z|A-Z|0-9|_|.]*")
)

type filter struct {	// 拦截器结构
	path 			string							// 原始过滤器路由
	rule 			*regexp.Regexp					// 过滤规则
	f  				*func( *RouterHandler )	(bool)	// 符合规则 执行方法 ; 返回是否阻塞路由的执行
	BlockNext 		bool							// 过滤器被执行后是否阻塞过滤器判断, 默认false
}

// 路由管理员
type RouterManager struct {

	Routers []*routerChip			// 所有的路由
	filter struct {					// 过滤器
		beforeRoute []*filter		// 匹配路由之前进行拦截
		afterRender []*filter 		// 渲染页面之后执行的拦截
	}
	httpStatus struct{
		notFound func(http.ResponseWriter)
		notAllowedMethod func(http.ResponseWriter)
	}

	staticRouter string
	staticPath string
}

type routerHandlerFunc func(handler *RouterHandler)

// 一条路由的类型
type routerChip struct {
	H *RouterHandler							// 传入的原生的路由数据
	Vars map[string]string						// 路由的传值
	Methods map[string]bool						// 路由可被访问的模式
	FilterFunc func(*RouterHandler)				// 当前路由的拦截器
	IsRouterValue bool							// 是否是通过路由传值

	path string									// 原始路由
	regPath string								// 正则后的路由
	viewRenderFunc func(*RouterHandler)			// 路由执行的视图
}

type RouterHandler struct {
	W http.ResponseWriter
	R *http.Request
	path string
	Vars map[string]string
}

// 接口实现方法
func (_self *RouterManager) ServeHTTP(w http.ResponseWriter, r *http.Request){

	var stopKey bool	// 禁止方法继续执行
	// 检测路由前的拦截器
	for _, v := range _self.filter.beforeRoute {
		if v.rule.MatchString(r.URL.Path){	// 正则匹配
			tmp := (*(v.f))( &RouterHandler{w,r,"", nil} )
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

	for _,v := range _self.Routers {

		// 1. 匹配路由
		if checkRouter(r.URL.Path, v.regPath) {

			// 2. 匹配Method
			if v.Methods==nil || v.Methods[r.Method] {
				v.H = &RouterHandler{ w, r, v.path, _self.getRouterValue(r.URL.Path, v)}

				// 3. 执行业务视图渲染
				v.viewRenderFunc(v.H)

				// 路由执行渲染后的拦截器
				for _, vv := range _self.filter.afterRender {
					if vv.rule.MatchString(r.URL.Path){	// 正则匹配
						(*(vv.f))( v.H )
						if vv.BlockNext {
							break
						}
					}
				}

			}else{

				// 访问模式拒绝页面
				switch strings.ToLower(r.Method ){

					case "get":
							defaultPages.Page_405_get(w)
						break

					case "post","put","delete":
							defaultPages.Page_405_post(w)
						break
				}


			}
			return
		}
	}

	// 没有路由情况下跳404
	// 访问模式拒绝页面
	switch strings.ToLower(r.Method ){

		case "get":
			defaultPages.Page_404_get(w)
			break

		case "post":
			defaultPages.Page_404_post(w)
			break
	}
}

// 注册一条新路由
func (_self *RouterManager) Register(path string, f routerHandlerFunc ) *routerChip{
	if path == _self.staticRouter {
		log.Fatalln("路由地址与静态文件路由地址冲突!\n[",path, "==>",_self.staticPath,"]")
	}
	return _self.addRouter(path, f)
}

// 针对路由的拦截器
// 参数: 拦截器位置, 拦截器执行方法
func (_self *RouterManager) InsertFilter(position string, pathRule string, f func(handler *RouterHandler)(bool), BlockNext ...bool){
	re, err := regexp.Compile(handlerPathString2regexp(pathRule))
	if err != nil {
		log.Println("过滤器:[",position,"]路径匹配错误[",err,"]!")
		re ,_ = regexp.Compile(`[\D|\d]*`)
	}

	filterStruct := &filter{ pathRule,re, &f ,len(BlockNext)>0 && BlockNext[0]==true}

	switch position {

		case "beforeRoute":
			_self.filter.beforeRoute = append( _self.filter.beforeRoute, filterStruct)
			break

		case "afterRender":
			_self.filter.afterRender = append(_self.filter.afterRender, filterStruct)
			break

	}

}

// 设置静态文件访问目录
func (_self *RouterManager) SetStaticPath(router string, path string){

	// 检测是否有重复路由
	for _, v := range _self.Routers {
		if v.path == router {
			log.Fatalln("设置的静态文件地址路由路径冲突!")
		}
	}
	_self.staticRouter = router
	_self.staticPath = path

	// 注册一个静态文件的路由
	_self.addRouter(handlerPathString2regexp(router+"**"), _self.makeFileServe(http.StripPrefix(router,http.FileServer(http.Dir(_self.staticPath)))))

}

// 注册一条新路由
func (_self *RouterManager) addRouter(path string, f routerHandlerFunc ) *routerChip{
	rp, is := handlerPath(path)
	innerH := &RouterHandler{ path: path }
	tmp := &routerChip{ path: path, regPath: rp, viewRenderFunc: f, IsRouterValue: is, H: innerH }
	_self.Routers = append( _self.Routers,  tmp )
	return tmp
}

// 获得路由解析传值
func (_self *RouterManager) getRouterValue(url string, rh *routerChip)map[string]string{

	res := make(map[string]string)
	reg,_ := regexp.Compile("[{|}]")

	// 不是路由传值
	if !rh.IsRouterValue {
		return res
	}

	routeSet:= strings.Split(rh.path, "/")

	urlSet := strings.Split(url, "/")

	for k,v := range routeSet {
		// 确认是路由上设置的变量
		if RegExp_Url_Set.MatchString(v) {
			res[reg.ReplaceAllString(v,"")] = urlSet[k]
		}
	}

	return res
}

// 文件路由的处理方法
func (_self *RouterManager) makeFileServe(handler http.Handler) routerHandlerFunc{

	return func(h *RouterHandler){

		handler.ServeHTTP(h.W, h.R)

	}

}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 设置路由的Method
func (_self *routerChip) Method(methods... string) *routerChip{
	_self.Methods = make(map[string]bool)
	for _,v := range methods {
		_self.Methods[strings.ToUpper(v)] = true
	}
	return _self
}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 创建新的路由
func NewRouter() *RouterManager{
	return &RouterManager{}
}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 把一个路由设置的URL转换成用于判断URL的正则
func handlerPath(path string)(string, bool){
	// 创建一个正则
	return "^" + RegExp_Url_Set.ReplaceAllLiteralString(path, RegExp_Url_String)+"$",!(RegExp_Url_Set.FindIndex([]byte(path))==nil)
}

// 路由匹配
func checkRouter(url, path string)bool{
	re, _ := regexp.Compile (path)
	return re.MatchString(url)
}

// 解析路由路径为正则字符串
func handlerPathString2regexp (path string)(string){
	temp_time := strconv.FormatInt(time.Now().Unix(), 10)
	// 双星
	doubleStartReg, _ := regexp.Compile(`\*\*`)
	// 单星
	startReg, _ := regexp.Compile(`\*`)
	// 双星转换的中间量
	swapReg, _ := regexp.Compile(temp_time)

	// 替换双星
	path = doubleStartReg.ReplaceAllString(path, `\S`+temp_time, )

	// 替换单星
	path = "^" + startReg.ReplaceAllString(path, `[a-z|A-Z|0-9|_|.]*`, ) + "$"

	// 替换中间变量
	path = swapReg.ReplaceAllString(path, `*`)

	return path

}
