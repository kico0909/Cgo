package module

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/Cgo/mysql"
	"log"
	"strconv"
	"strings"
)

type DataModlues struct {
	conn *mysql.DatabaseMysql
}

func New(sqlSource *mysql.DatabaseMysql) *DataModlues {
	return &DataModlues{conn: sqlSource}
}

func (dm *DataModlues) Add(tableName ...string) *TableModule {
	return &TableModule{origin: dm.conn, Table: "`" + strings.Join(tableName, "`,`") + "`"}
}

type TableModule struct {
	origin *mysql.DatabaseMysql
	Table,
	value,
	where,
	orderBy,
	insert,
	update,
	execHeader,
	sqlStr string // Sql语句使用到的字符串
}

// 连表
func (dm *TableModule) Join(tableModel ...interface{}) *TableModule {
	var tablesArr []string
	tablesArr = append(tablesArr, strings.Replace(dm.Table, "`", "", 2))
	for _, v := range tableModel {
		tv := v.(TableModule)
		tablesArr = append(tablesArr, strings.Replace(tv.Table, "`", "", 2))
	}
	tmp := &TableModule{origin: dm.origin, Table: "`" + strings.Join(tablesArr, "`,`") + "`"}
	return tmp
}

func (_self *TableModule) reset() {
	_self.value = ""
	_self.where = ""
	_self.insert = ""
	_self.update = ""
	_self.execHeader = ""
	_self.orderBy = ""
}

// 获得数据表的全部信息
func (_self *TableModule) Get(num ...int64) (mysql.DbQueryReturn, error) {
	if len(_self.value) == 0 {
		_self.value = "*"
	}
	if len(_self.where) <= 0 {
		_self.where = " "
	}
	_self.sqlStr = "select " + _self.value + " from " + _self.Table + _self.where + " " + _self.orderBy
	if len(num) == 1 {
		_self.sqlStr += " limit 0," + strconv.FormatInt(num[0], 10)
	}
	if len(num) == 2 {
		_self.sqlStr += " limit " + strconv.FormatInt(num[0], 10) + "," + strconv.FormatInt(num[1], 10)
	}

	_self.reset()

	return _self.Query(_self.sqlStr)

}

// 无返回的执行
func (_self *TableModule) Save() (sql.Result, error) {

	switch _self.execHeader {

	case "replace into":
		_self.sqlStr = _self.execHeader + " " + _self.Table + _self.insert + _self.where
		break

	case "insert into":
		_self.sqlStr = _self.execHeader + " " + _self.Table + _self.insert + _self.where
		break

	case "update":
		_self.sqlStr = _self.execHeader + " " + _self.Table + " set " + _self.update + _self.where + _self.orderBy
		break

	}
	_self.reset()
	return _self.Exec(_self.sqlStr)
}

// 删除
func (_self *TableModule) Del() (sql.Result, error) {
	_self.sqlStr = "delete from " + _self.Table + _self.where
	_self.reset()
	return _self.Exec(_self.sqlStr)
}

// 按需求输出部分结果
func (_self *TableModule) Value(n ...string) *TableModule {
	_self.value = strings.Join(n, ",")
	return _self
}

// 条件
func (_self *TableModule) Where(s ...string) *TableModule {
	if len(s) < 1 {
		_self.where = " "
	} else {
		_self.where = " where " + strings.Join(s, " ") + " "
	}
	return _self
}
func (_self *TableModule) NewForStruct(v interface{}, replace ...bool) *TableModule {

	var tmp interface{}
	tmpMap := make(map[string]interface{})
	b, _ := json.Marshal(v)
	jsonDecoder := json.NewDecoder(bytes.NewBuffer(b))
	jsonDecoder.UseNumber()
	jsonDecoder.Decode(&tmp)
	tmpMap = tmp.(map[string]interface{})

	return _self.New(tmpMap, replace...)
}

// 添加新数据
func (_self *TableModule) New(k2v map[string]interface{}, replace ...bool) *TableModule {
	var replaceKey bool
	if len(replace) > 0 {
		replaceKey = replace[0]
	}

	var keys []string
	var values []string
	if replaceKey {
		_self.execHeader = "replace into"
	} else {
		_self.execHeader = "insert into"
	}

	for k, v := range k2v {

		keys = append(keys, k)

		switch v.(type) {
		case int:
			values = append(values, strconv.FormatInt(int64(v.(int)), 10))
			break
		case int32:
			values = append(values, strconv.FormatInt(int64(v.(int32)), 10))
			break
		case int64:
			values = append(values, strconv.FormatInt(int64(v.(int64)), 10))
			break
		case json.Number:
			values = append(values, string(v.(json.Number)))
			break
		case string:
			values = append(values, "'"+v.(string)+"'")
			break
		}
	}
	_self.insert = " (" + strings.Join(keys, ",") + ") Values (" + strings.Join(values, ",") + ") "
	return _self
}

// 更新数据
func (_self *TableModule) Update(k2v map[string]interface{}) *TableModule {
	var tmpStr []string
	_self.execHeader = "update"
	for k, v := range k2v {
		tmp := ""
		switch v.(type) {
		case int:
			tmp = strconv.FormatInt(int64(v.(int)), 10)
			break
		case int32:
			tmp = strconv.FormatInt(int64(v.(int32)), 10)
			break
		case int64:
			tmp = strconv.FormatInt(v.(int64), 10)
			break
		case string:
			tmp = "'" + v.(string) + "'"
			break
		}
		tmpStr = append(tmpStr, k+"="+tmp)
	}
	_self.update = strings.Join(tmpStr, ",")
	return _self
}

// 规则
func (_self *TableModule) OrderBy(order ...string) *TableModule {
	_self.orderBy = strings.Join(order, " ")
	return _self
}

// 查询操作
func (_self *TableModule) Query(sqlStr string) (mysql.DbQueryReturn, error) {
	_self.sqlStr = sqlStr + ";"
	log.Println("query ==> ", _self.sqlStr)
	return _self.origin.Query(_self.sqlStr)
}

// 执行操作
func (_self *TableModule) Exec(sqlStr string) (sql.Result, error) {
	_self.sqlStr = sqlStr + ";"
	log.Println("exec ==> ", _self.sqlStr)
	return _self.origin.Exec(_self.sqlStr)
}
