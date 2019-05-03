package route

import (
	beegoSession "github.com/astaxie/beego/session"
	"net/http"
	"regexp"
)

type filter struct { // 拦截器结构
	path      string                     // 原始过滤器路由
	rule      *regexp.Regexp             // 过滤规则
	f         *func(*RouterHandler) bool // 符合规则 执行方法 ; 返回是否阻塞路由的执行
	BlockNext bool                       // 过滤器被执行后是否阻塞过滤器判断, 默认false
}

// 路由管理员
type RouterManager struct {
	Routers []*routerChip // 所有的路由
	filter  struct {      // 过滤器
		beforeRoute []*filter // 匹配路由之前进行拦截
		afterRender []*filter // 渲染页面之后执行的拦截
	}
	httpStatus struct {
		notFound         func(http.ResponseWriter)
		notAllowedMethod func(http.ResponseWriter)
	}

	staticRouter string
	staticPath   string
}

type routerHandlerFunc func(handler *RouterHandler)

// 一条路由的类型
type routerChip struct {
	H             *RouterHandler       // 传入的原生的路由数据
	Vars          map[string]string    // 路由的传值
	Methods       map[string]bool      // 路由可被访问的模式
	FilterFunc    func(*RouterHandler) // 当前路由的拦截器
	IsRouterValue bool                 // 是否是通过路由传值

	path           string               // 原始路由
	regPath        string               // 正则后的路由
	viewRenderFunc func(*RouterHandler) // 路由执行的视图
}

type RouterHandler struct {
	W       http.ResponseWriter
	R       *http.Request
	path    string
	Vars    map[string]string
	Session beegoSession.Store
	Values  map[string]interface{}
}

type defaultApiCodeType struct {
	Success int64
	Fail    int64
}

// ------------------------------------------------------------------------

//
// 路由组 每一个Chip的返回值 方便上一级的处理
type routerGroupFinilReturn struct {
	Path        string
	HandlerFunc routerHandlerFunc
	Methods     string
}
