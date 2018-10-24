package defaultPages

import (
	"net/http"
	)

// 404页面
func NotFound(w http.ResponseWriter) {
	page_404.Execute(w,nil)
}
