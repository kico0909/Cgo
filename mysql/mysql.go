package mysql

import (
	"database/sql"
	_ "github.com/Cgo/go-sql-driver/mysql"
	"log"
	"errors"
	"os"

)

type DbQueryReturn []map[string]string

type dbinfoType struct {
	username string
	password string
	host string
	port string
	dbname string
	socket string
}

type DatabaseMysql struct {
	dbInfo map[string]dbinfoType
	conn map[string]*sql.DB
	connectionName string
	nowDBType string
}

/*
私有方法, 用于连接数据库
*/
func (_self *DatabaseMysql) connectionDB() bool {

	username := _self.dbInfo[_self.connectionName].username
	password := _self.dbInfo[_self.connectionName].password
	host := _self.dbInfo[_self.connectionName].host
	port := _self.dbInfo[_self.connectionName].port
	dbname := _self.dbInfo[_self.connectionName].dbname
	socket := _self.dbInfo[_self.connectionName].socket

	var dataSourceName string

	_, err := os.Stat(socket)
	// 存在套字链接的路径, 优先使用套子链接
	if len(socket)>0 && err==nil {
		dataSourceName = username + `:` + password + `@unix(` + socket + `)/` + dbname
	} else {
		if (host == "localhost" || host == "127.0.0.1") && port=="3306" {

			dataSourceName = username + `:` + password + `@` + `/` + dbname

		}else{

			dataSourceName = username + `:` + password + `@tcp(` + host + `:` + port + `)/` + dbname

		}
	}

	_db, _err := sql.Open("mysql", dataSourceName)


	if _err != nil {
		log.Println(_err)
		return false;
	}

	// 最大连接
	_db.SetMaxOpenConns(200)

	// 保持连接
	_db.SetMaxIdleConns(50)

	dbPing := _db.Ping()
	if dbPing!=nil {
		log.Println(dbPing)
		return false
	}

	if _err == nil {
		if _self.conn == nil {
			_self.conn = make(map[string]*sql.DB )
		}
		_self.conn[ _self.connectionName ] = _db
		return true
	}
	//*sql.DB
	return false
}

/*
私有方法, 用于关闭数据库
*/
func (_self *DatabaseMysql) closeDB(){
	_self.conn[_self.connectionName].Close()
}

/*
初始化,获得数据库配置信息
*/
func (_self *DatabaseMysql) Init(connectionName, usn, pwd, host, port, dbname,socket string)bool{
	_self.connectionName = connectionName;
	if _self.dbInfo == nil {
		_self.dbInfo = make(map[string]dbinfoType)
	}
	_self.dbInfo[connectionName] = dbinfoType{usn, pwd, host, port, dbname, socket}
	return _self.connectionDB()
}

/*
链接数据库, 根据配置文件中的配置信息去对数据库进行连接
*/
func (_self *DatabaseMysql) Connection(connectionName string) {
	log.Println(connectionName)
	_self.connectionName = connectionName
	_self.connectionDB()
}

/*
数据库查询操作
*/
func (_self *DatabaseMysql) Query (sql string)  (results DbQueryReturn, err error) {

	//var results DbQueryReturn   // 返回的类型
	conn, ok := _self.conn[_self.connectionName]

	if !ok {
		return nil, errors.New("get mysql connection error!")
	}

	rows, err := conn.Query(sql)
	if err != nil {
		return nil, errors.New("sql query error["+err.Error()+"]")
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
		for k, v := range values { //每行数据是放在values里面，现在把它挪到row里
			key := cols[k]
			row[key] = string(v)
		}

		results = append(results, row)
	}
	return results, err
}

/*
非查询类数据库操作
*/
func (_self *DatabaseMysql) Exec(query string, args ...interface{}) (sql.Result, error ){
	return _self.conn[_self.connectionName].Exec(query)
}