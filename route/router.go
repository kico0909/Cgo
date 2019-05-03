package route

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

var (
	Method_Error          = errors.New("路由访问方式错误!")
	RegExp_Url_Set_String = "{[a-z|A-Z|0-9|_]*}"
	RegExp_Url_Set, _     = regexp.Compile(RegExp_Url_Set_String)
	RegExp_Url_String     = "[a-z|A-Z|0-9|_|.]*"
	RegExp_Url, _         = regexp.Compile("[a-z|A-Z|0-9|_|.]*")
	RegExp_GroupPath, _   = regexp.Compile("[a-z|A-Z|0-9|_|.]*//")
	defaultApiCode        = defaultApiCodeType{200, 400}
	default_methods       = []string{"POST", "GET", "PUT", "DELETE"}
	default_method_post   = "POST"
	default_method_get    = "GET"
	default_method_put    = "PUT"
	default_method_delete = "DELETE"
)

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
func (r *RouterHandler) GetBodyValueToJson(res interface{}) error {
	defer r.R.Body.Close()
	b, err := ioutil.ReadAll(r.R.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, res)
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
