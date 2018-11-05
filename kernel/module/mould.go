package module

import (
	"github.com/Cgo/mysql"
	"strconv"
	"strings"
	"database/sql"
	)

type DataModlues struct {
	conn *mysql.DatabaseMysql
}

func New (sqlSource  *mysql.DatabaseMysql)*DataModlues {
	return &DataModlues{conn: sqlSource}
}

func (dm *DataModlues) Add (tableName ...string)*TableModule{
	return &TableModule{origin: dm.conn, table: "`" + strings.Join(tableName,"`,`") + "`"}
}

type TableModule struct {
	origin 	*mysql.DatabaseMysql
	table, value, where, orderBy, insert, update, execHeader, sqlStr	string	// Sql语句使用到的字符串
}

func (_self *TableModule) reset(){
	_self.value = ""
	_self.where = ""
	_self.insert = ""
	_self.update = ""
	_self.execHeader = ""
}

// 获得数据表的全部信息
func (_self *TableModule) Get(num ...int64) (mysql.DbQueryReturn, error){
	if len(_self.value) == 0 {
		_self.value = "*"
	}
	if len(_self.where) <= 0 {
		_self.where = " "
	}
	_self.sqlStr = "select " + _self.value + " from " + _self.table +_self.where + " " + _self.orderBy
	if len(num) == 1 {
		_self.sqlStr += " limit 0," + strconv.FormatInt(num[0], 10)
	}
	if len(num) == 2 {
		_self.sqlStr += " limit "+strconv.FormatInt(num[0], 10)+"," + strconv.FormatInt(num[1], 10)
	}

	_self.reset()

	return _self.Query(_self.sqlStr)

}

// 无返回的执行
func (_self *TableModule) Save()(sql.Result, error){

	switch _self.execHeader {

	case "replace into":
		_self.sqlStr = _self.execHeader + " " + _self.table + _self.insert +  _self.where
		break

	case "insert into":
		_self.sqlStr = _self.execHeader + " " + _self.table +  _self.insert +  _self.where
		break

	case "update":
		_self.sqlStr = _self.execHeader + " " + _self.table + " set " + _self.update + _self.where + _self.orderBy
		break

	}

	_self.reset()

	return _self.Exec(_self.sqlStr)
	//return true
}

// 删除
func (_self *TableModule) Del()( sql.Result, error ) {
	_self.sqlStr = "delete from " + _self.table + _self.where
	_self.reset()
	return _self.Exec(_self.sqlStr)
}

// 按需求输出部分结果
func (_self *TableModule) Value (n ...string)*TableModule{
	_self.value = strings.Join(n, ",")
	return _self
}

// 条件
func (_self *TableModule) Where ( s ...string)*TableModule {
	if len(s) <1 {
		_self.where = " "
	}else{
		_self.where = " where " + strings.Join(s, " ") + " "
	}
	return _self
}

// 添加新数据
func (_self *TableModule) New(k2v map[string]interface{}, replace bool) *TableModule {
	var keys  []string
	var values []string
	if replace {
		_self.execHeader = "replace into"
	}else{
		_self.execHeader = "insert into"
	}

	for k,v := range k2v {

		keys = append(keys, k)

		switch v.(type) {
			case int:
				values = append(values,strconv.FormatInt(int64(v.(int)),10))
				break
			case string:
				values = append(values,"'"+v.(string)+"'")
				break;
		}

	}
	_self.insert = " (" +strings.Join(keys, ",") + ") Values (" +  strings.Join(values, ",") + ") "
	return _self
}

// 更新数据
func (_self *TableModule) Update(k2v map[string]interface{}) *TableModule {
	var tmpStr  []string
	_self.execHeader = "update"
	for k,v := range k2v {
		tmp := ""
		switch v.(type) {
		case int:
			tmp = strconv.FormatInt(int64(v.(int)),10)
			break
		case string:
		tmp = "'"+v.(string)+"'"
			break;
		}
		tmpStr = append(tmpStr, k+"="+tmp)
	}
	_self.update = strings.Join(tmpStr, ",")
	return _self
}

// 规则
func (_self *TableModule) OrderBy (order ...string) *TableModule {
	_self.orderBy = strings.Join(order, " ")
	return _self
}

// 查询操作
func (_self *TableModule) Query (sqlStr string)(mysql.DbQueryReturn, error){
	return  _self.origin.Query(_self.sqlStr)
}
// 执行操作
func (_self *TableModule) Exec (sqlStr string)(sql.Result, error ){
	return  _self.origin.Exec(_self.sqlStr)
}




