package mysql

import (
	"github.com/Cgo/mysql"
	"github.com/Cgo/kernel/config"
	log "github.com/sirupsen/logrus"
)

func New (conf *config.ConfigData) *mysql.DatabaseMysql {
	log.Info("功能初始化: MYSQL(default)					[ ok ]")
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