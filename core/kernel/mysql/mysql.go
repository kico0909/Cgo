package mysql

import (
	config "cgo/core/kernel/config"
	"cgo/core/kernel/logger"
	mysql "cgo/core/mysql"
)

func New(conf *config.ConfigMysqlOptions) *mysql.DatabaseMysql {
	return mysql.New(conf, nil, func() {
		log.Println("功能初始化: MYSQL --- [ ok ]")
	})
}
