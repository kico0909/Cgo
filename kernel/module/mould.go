package module

import (
	"github.com/Cgo/mysql"
	"strconv"
	"strings"
	"database/sql"
	)

type DBModel struct {
	Table string	// 表名称
	origin *mysql.DatabaseMysql

	value string		// 需要返回的值部分
	where string	// Where 语句部分
	orderBy string	// orderBy
	insert string	// 插入时候的语句翻译
	update string	// UPDATE 的语句
	execHeader string	// 无返回的执行类型

	sqlStr string
}

func (_self *DBModel) reset(){
	_self.value = ""
	_self.where = ""
	_self.insert = ""
	_self.update = ""
	_self.execHeader = ""
}

// 获得数据表的全部信息
func (_self *DBModel) Get(num ...int64) (mysql.DbQueryReturn, error){
	if len(_self.value) == 0 {
		_self.value = "*"
	}
	if len(_self.where) <= 0 {
		_self.where = " "
	}
	_self.sqlStr = "select " + _self.value + " from " + _self.Table +_self.where + " " + _self.orderBy
	if len(num) == 1 {
		_self.sqlStr += " limit 0," + strconv.FormatInt(num[0], 10)
	}
	if len(num) == 2 {
		_self.sqlStr += " limit "+strconv.FormatInt(num[0], 10)+"," + strconv.FormatInt(num[1], 10)
	}

	_self.reset()

	return _self.origin.Query(_self.sqlStr)

}

// 无返回的执行
func (_self *DBModel) Save()(sql.Result, error){

	switch _self.execHeader {

	case "replace into":
		_self.sqlStr = _self.execHeader + " " + _self.Table + _self.insert +  _self.where
		break

	case "insert into":
		_self.sqlStr = _self.execHeader + " " + _self.Table +  _self.insert +  _self.where
		break

	case "update":
		_self.sqlStr = _self.execHeader + " " + _self.Table + " set " + _self.update + _self.where + _self.orderBy
		break

	}

	_self.reset()

	return _self.origin.Exec(_self.sqlStr)
	//return true
}

// 删除
func (_self *DBModel) Del()( sql.Result, error ) {
	_self.sqlStr = "delete from " + _self.Table + _self.where
	_self.reset()
	return _self.origin.Exec(_self.sqlStr)
}

// 按需求输出部分结果
func (_self *DBModel) Value (n ...string)*DBModel{
	_self.value = strings.Join(n, ",")
	return _self
}

// 条件
func (_self *DBModel) Where ( s ...string)*DBModel {
	if len(s) <1 {
		_self.where = " "
	}else{
		_self.where = " where " + strings.Join(s, " ") + " "
	}
	return _self
}

// 添加新数据
func (_self *DBModel) New(k2v map[string]interface{}, replace bool)*DBModel{
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
func (_self *DBModel) Update(k2v map[string]interface{})*DBModel{
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
func (_self *DBModel) OrderBy (order ...string) *DBModel {
	_self.orderBy = strings.Join(order, " ")
	return _self
}




