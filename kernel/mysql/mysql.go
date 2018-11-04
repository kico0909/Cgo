package mysql

import (
	"github.com/Cgo/mysql"
	"github.com/Cgo/kernel/config"
	"log"
)

func New (conf *config.ConfigMysqlOptions) *mysql.DatabaseMysql {
	log.Println("功能初始化: MYSQL(default) --- [ ok ]")
	return mysql.New(conf)
}