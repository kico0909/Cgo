package mysql

import (
	"github.com/Cgo/mysql"
	"github.com/Cgo/kernel/config"
	"log"
)

func New (conf *config.ConfigData) *mysql.DatabaseMysql {
	log.Print("初始化MYSQL [ default ] 的链接 ... \n")
	var tmp = &mysql.DatabaseMysql{}
	tmp.Init("default",
		conf.Mysql.Default.Username,
		conf.Mysql.Default.Password,
		conf.Mysql.Default.Host,
		conf.Mysql.Default.Port,
		conf.Mysql.Default.Dbname,
		conf.Mysql.Default.Socket)

	return tmp

}