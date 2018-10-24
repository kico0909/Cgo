/*
操作文件系统
*/
package funcs

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"html/template"
	"log"
)

/*
将模板文件缓存至内存, 自动检测模板目录内的文件
*/

var TemplateCached *template.Template

func CacheHtmlTemplate(templatePath string) {
	var err error
	TemplateCached, err  = template.ParseGlob( templatePath + "/*.html")
	if err != nil {
		log.Println("模板缓存异常: 没有模板文件可被缓存!")
	}

}

/*
渲染模板
渲染基于缓存的模板
*/
func RenderHtml(w http.ResponseWriter, templateName string, data interface{}) {
	err := TemplateCached.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Print("模板创建页面错误----->", templateName, data)
		log.Print(err)
	}
}

/*
基于文件的模板渲染
*/
func RenderFile(w http.ResponseWriter, prefectPath string){
	fileCont,_ := ioutil.ReadFile(prefectPath)
	fmt.Fprintf(w,string(fileCont))
}

