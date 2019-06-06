package mysql

/*
 Cgo的mysql方法实现, 可读写分离, 读库要使用Query方法,写库要使用Exec方法, 否则会导致读写错乱
*/

import (
	"database/sql"
	"errors"
	_ "github.com/Cgo/go-sql-driver/mysql"
	"github.com/Cgo/kernel/config"
	"github.com/Cgo/kernel/logger"
	"os"
	"strconv"
)

type DbQueryReturn []map[string]string

type DBQueryReturnNew []interface{}

// 数据库链接信息
type dbConnectionInfoType struct {
	username string
	password string
	host     string
	port     string
	dbname   string
	socket   string
	charset  string
}

// 数据库分类
type dbConnectionsInfoType struct {
	r dbConnectionInfoType
	w dbConnectionInfoType
}

// 链接分类
type connType struct {
	r *sql.DB
	w *sql.DB
}

type DatabaseMysql struct {
	sqlmode           string // 数据库模式: 单库, 读写分离
	dbConnectionsInfo dbConnectionsInfoType
	conn              connType
	//connectionName string
	//nowDBType string
}

func createConnectionInfo(conf dbConnectionInfoType) string {

	if len(conf.charset) < 1 {
		conf.charset = "utf8"
	}

	//  链接写库
	_, err := os.Stat(conf.socket)
	// 存在套字链接的路径, 优先使用套子链接
	if err == nil {
		return conf.username + `:` + conf.password + `@unix(` + conf.socket + `)/` + conf.dbname + "?charset=" + conf.charset
	} else {
		if (conf.host == "localhost" || conf.host == "127.0.0.1") && conf.port == "3306" {
			return conf.username + `:` + conf.password + `@/` + conf.dbname + "?charset=" + conf.charset
		} else {
			return conf.username + `:` + conf.password + `@tcp(` + conf.host + `:` + conf.port + `)/` + conf.dbname + "?charset=" + conf.charset
		}
	}
}

// 连接数据库
func (_self *DatabaseMysql) connectionDB() *DatabaseMysql {

	// 连接写库
	_dbw, _err := sql.Open("mysql", createConnectionInfo(_self.dbConnectionsInfo.w))
	if _err != nil {
		log.Fatalln("数据库连接[读]出现错误: ", _err)
	}

	// 最大连接
	_dbw.SetMaxOpenConns(200)

	// 保持连接
	_dbw.SetMaxIdleConns(50)

	dbPing_w := _dbw.Ping()
	if dbPing_w != nil {
		log.Fatalln("数据库连接[读]无法Ping通: ")
	}

	_self.conn.w = _dbw

	if _self.sqlmode == "default" {
		_self.conn.r = _dbw
		return _self
	}

	// 连接读库
	_dbr, _err := sql.Open("mysql", createConnectionInfo(_self.dbConnectionsInfo.r))

	if _err != nil {
		log.Fatalln("数据库连接[写]出现错误: ", _err)
	}

	// 最大连接
	_dbr.SetMaxOpenConns(200)

	// 保持连接
	_dbr.SetMaxIdleConns(50)

	dbPing := _dbr.Ping()
	if dbPing != nil {
		log.Fatalln("数据库连接[读]无法Ping通: ")
	}

	_self.conn.r = _dbr

	return _self

}

/*
私有方法, 用于关闭数据库
*/
func (_self *DatabaseMysql) closeDB() {
	_self.conn.w.Close()
	_self.conn.r.Close()
}

// 根据连接信息 初始化数据库
func New(conf *config.ConfigMysqlOptions) *DatabaseMysql {

	// 模式判断
	var sqlMode = ""
	if len(conf.Default.Host) < 1 {
		sqlMode = "rw"
	} else {
		sqlMode = "default"
	}

	// 连接信息生成
	var wDBinfo dbConnectionInfoType
	var rDBinfo dbConnectionInfoType

	if sqlMode == "default" {
		wDBinfo.username = conf.Default.Username
		wDBinfo.password = conf.Default.Password
		wDBinfo.host = conf.Default.Host
		wDBinfo.port = strconv.FormatInt(conf.Default.Port, 10)
		wDBinfo.dbname = conf.Default.Dbname
		wDBinfo.socket = conf.Default.Socket
		wDBinfo.charset = conf.Default.Charset
		rDBinfo = wDBinfo
	} else {
		wDBinfo.username = conf.Write.Username
		wDBinfo.password = conf.Write.Password
		wDBinfo.host = conf.Write.Host
		wDBinfo.port = strconv.FormatInt(conf.Write.Port, 10)
		wDBinfo.dbname = conf.Write.Dbname
		wDBinfo.socket = conf.Write.Socket
		wDBinfo.charset = conf.Default.Charset

		rDBinfo.username = conf.Read.Username
		rDBinfo.password = conf.Read.Password
		rDBinfo.host = conf.Read.Host
		rDBinfo.port = strconv.FormatInt(conf.Read.Port, 10)
		rDBinfo.dbname = conf.Read.Dbname
		rDBinfo.socket = conf.Read.Socket
		rDBinfo.charset = conf.Read.Charset
	}

	// 创建实例
	tmp := &DatabaseMysql{sqlmode: sqlMode, dbConnectionsInfo: dbConnectionsInfoType{w: wDBinfo, r: rDBinfo}}

	return tmp.connectionDB()
}

// 数据库查询操作
func (_self *DatabaseMysql) Query(sql string) (results DbQueryReturn, err error) {

	//var results DbQueryReturn   // 返回的类型
	rows, err := _self.conn.r.Query(sql)
	if err != nil {
		return nil, errors.New("sql query error[" + err.Error() + "]")
	}

	defer rows.Close()

	//读出查询出的列字段名
	cols, _ := rows.Columns()

	//values是每个列的值，这里获取到byte里
	values := make([][]byte, len(cols))

	//query.Scan的参数，因为每次查询出来的列是不定长的，用len(cols)定住当次查询的长度
	scans := make([]interface{}, len(cols))

	//让每一行数据都填充到[][]byte里面
	for i := range values {
		scans[i] = &values[i]
	}

	for rows.Next() { //循环，让游标往下推
		if err := rows.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			log.Println(err)
		}
		row := make(map[string]string) //每行数据
		for k, v := range values {     //每行数据是放在values里面，现在把它挪到row里
			key := cols[k]
			row[key] = string(v)
		}

		results = append(results, row)
	}
	return results, err
}

// 非查询类数据库操作
func (_self *DatabaseMysql) Exec(query string, args ...interface{}) (sql.Result, error) {
	return _self.conn.w.Exec(query)
}
