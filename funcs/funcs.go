package funcs

import "encoding/json"

type hrr struct {
	Success	bool `json:"success"`
	Error	error	`json:"error"`
	Data	map[string]interface{}	`json:"data"`
	Code	int	`json:"code"`
}
// 用于创建http请求的返回
func CreateHttpRequestResult (success bool, err error, data map[string]interface{}, code int)string{
	if success {
		code = 200
	}

	jsonObj := &hrr{ Success: success, Error: err, Data: data, Code: code}
	res,_ := json.Marshal(jsonObj)
	return string(res)
}