package defaultPages

import (
	"html/template"
)

// 404 模板
var page_404_get = template.Must(template.New("").Parse(
	`<!DOCTYPE html>
			<html lang="hz-cn">
			<head>
				<meta charset="UTF-8">
				<title>CBIM - 404</title>
			</head>
			<body style="margin: 0; padding: 0; background: #f7f7f7;">
				<style>
					p{ width: 100%; min-height: 0; text-align: center; }
					.big{color:#006489; font-size: 50px; line-height: 50px; text-align: center; margin-top: 100px;}
					.small{font-size: 16px; color: #ccc; margin-top: 20px;}
					.small a{ }
				</style>
				<p class="big">404</p>
				<p class="big">Not Found</p>
				<p class="small"><a href="/">返回首页</a></p>
			</body>
			</html>`))

var page_404_post = template.Must(template.New("").Parse(`{code:404, success: false, error: "Not Found"}`))
