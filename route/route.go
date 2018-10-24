package route

import (
	"net/http"
	"log"
	"regexp"
	"strings"
	"fmt"
	"errors"
	"html/template"
	"github.com/Cgo/defaultPages"
)

var (
	Method_Error = errors.New("路由访问方式错误!")
	RegExp_Url_Str = "{[a-z|A-Z|0-9|_]*}"
)

// 路由管理员
type RouterManager struct {

	Routers []*routerChip	// 所有的路由
	filter struct {
		beforeRoute func( *RouterHandler )		// 匹配路由之前进行拦截
		afterRoute func( *RouterHandler )		// 匹配路由之后进行拦截
		afterRender func( *RouterHandler )	// 渲染页面之后执行的拦截
	}

	notFound *template.Template

}

//
type routerChip struct {
	H *RouterHandler										// 传入的原生的路由数据
	Vars map[string]string									// 路由的传值
	Methods map[string]bool									// 路由可被访问的模式
	FilterFunc func(*RouterHandler)	// 当前路由的拦截器
	IsRouterValue bool										// 是否是通过路由传值

	path string												// 原始路由
	regPath string											// 正则后的路由
	viewRenderFunc func(*RouterHandler)	// 路由执行的视图
}

type RouterHandler struct {
	W http.ResponseWriter
	R *http.Request
	path string
}

func (_self *RouterManager) ServeHTTP(w http.ResponseWriter, r *http.Request){

	filter := RouterHandler{w,r,""}

	// 检测路由前的拦截器
	if _self.filter.beforeRoute != nil{
		_self.filter.beforeRoute(&filter)
	}

	// 1. 匹配路由
	for _,v := range _self.Routers {

		// 1. 匹配路由
		if checkRouter(r.URL.Path, v.regPath) {
			// 2. 匹配method
			if v.Methods[r.Method] {

				v.H = &RouterHandler{ w, r, v.path }

				// 路由执行前的拦截器
				if _self.filter.afterRoute != nil{
					_self.filter.afterRoute(v.H)
				}

				// 4. 执行业务视图渲染

				v.viewRenderFunc(v.H)

				// 路由执行渲染后的拦截器
				if _self.filter.afterRender != nil{
					_self.filter.afterRender(v.H)
				}
			}else{
				fmt.Fprintf(w,Method_Error.Error())
			}
			return
		}
	}
	// 没有路由情况下跳404
	defaultPages.NotFound(w)
}


// 注册一条新路由
func (_self *RouterManager) Register(path string, f func(handler *RouterHandler) ) *routerChip{
	rp, is := handlerPath(path)
	innerH := &RouterHandler{ path: path }
	tmp := &routerChip{ path: path, regPath: rp, viewRenderFunc: f, IsRouterValue: is, H: innerH }
	_self.Routers = append( _self.Routers,  tmp )
	return tmp
}

// 针对路由的拦截器
// 参数: 拦截器位置, 拦截器执行方法
func (_self *RouterManager) Filter(position string, f func(handler *RouterHandler)){

	switch position {
		case "beforeRoute":
			_self.filter.beforeRoute = f
			break
		case "afterRoute":
			_self.filter.afterRoute = f
			break
		case "afterRender":
			_self.filter.afterRender = f
			break
	}

}




// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 设置路由的Method
func (_self *routerChip) Mehtods(methods... string) *routerChip{
	_self.Methods = make(map[string]bool)
	for _,v := range methods {
		_self.Methods[strings.ToUpper(v)] = true
	}
	return _self
}

// 针对当前路由的拦截器
func (_self *routerChip) Filter(f func(*RouterHandler)) *routerChip{
	_self.FilterFunc = f
	return _self
}


// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 获得路由解析传值
func (_self *RouterHandler)GetRouterValue(key string){
	log.Println("key===>", key)
}



// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 创建新的路由
func NewRouter() *RouterManager{
	return &RouterManager{}
}


// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

func handlerPath(path string)(string, bool){
	// 创建一个正则
	regexp, _ := regexp.Compile(RegExp_Url_Str)
	return regexp.ReplaceAllLiteralString(path, "[a-z|A-Z|0-9|_|.]*"),!(regexp.FindIndex([]byte(path))==nil)
}

// 路由匹配
func checkRouter(url, path string)bool{
	re,_ := regexp.Compile("^" + path + "$")
	return re.MatchString(url)
}
