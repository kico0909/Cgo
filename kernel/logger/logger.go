package log

import (
	olog "log"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	Hour = 3600
	Minute = 60
	Second = 1
	OneDay = 1 * 24 * 60 * 60
)

var std = olog.New(os.Stderr, "", 3)

type Logger struct {
	mode 				string			// 输出模式
	file 				*os.File		// 输出到文件
	autoCutOff			bool			// 自动截断日志文档
	logger				*olog.Logger	//
	debugMode			bool			// DEBUG状态
	nextCutDate			string			// 上一次切割的日期
	prefix				string			// 日志前缀命名
	path				string			// 日志保存路径

}

func New(path, prefix string, autoCut bool)(*Logger){

	var resLogger *Logger

	if len(path)>1{
		f, err := os.OpenFile(
			path + prefix + "_cgo_log",
			os.O_CREATE|os.O_RDWR|os.O_APPEND,
			0777)

		if err != nil {
			resLogger = &Logger{
				file: f,
				autoCutOff: autoCut,
				logger: std,
				debugMode: false,

				prefix: prefix,
				path: path,
				mode: "Terminal"}
		}else{

			var logger = olog.New(f, "", 3)

			resLogger = &Logger{
				file: f,
				autoCutOff: autoCut,
				logger: logger,
				debugMode: false,

				prefix: prefix,
				path: path,
				mode: "File"}
		}

		if autoCut {
			go resLogger.autoCutOffLogFileHandler()
		}
	}

	Println("功能初始化: Cgo日志系统 --- [ ok ]")

	return resLogger
}

// 设置日志的debug模式, 默认关闭
func (this *Logger) SetDebugMode (key bool){
	this.debugMode = key
}

func (this *Logger) Print (v ...interface{}){
	this.logger.Println(splice(v, 0, false, "[Normal]")...)
}

func (this *Logger) Println (v ...interface{}){
	this.logger.Print(splice(v, 0, false, "[Normal]")...)
}

func (this *Logger) Info (v ...interface{}){
	this.logger.Println(splice(v, 0, false, "[INFO]")...)
}

func (this *Logger) Warn (v ...interface{}){
	this.logger.Println(splice(v, 0, false, "[WARN]")...)
}

func (this *Logger) Error (v ...interface{}){
	this.logger.Println(splice(v, 0, false, "[ERROR]")...)
}

func (this *Logger) Fatal (v ...interface{}){
	this.logger.Fatal(splice(v, 0, false, "[Fatal]")...)
}

func (this *Logger) Fatalln (v ...interface{}){
	this.logger.Fatalln(splice(v, 0, false, "[Fatal]")...)
}

// debug模式下可以使用,设置为非debug 模式则不
func (this *Logger) Debug (v ...interface{}){
	if !this.debugMode {
		return
	}
	this.logger.Println(splice(v, 0, false, "[DEBUG]")...)
}

// 开一个定时线程执行文件分割, 按天执行
func (this *Logger) autoCutOffLogFileHandler(){
	if this.mode == "Terminal" {
		return
	}

	// 启动时先进行一次分割
	this.cunFile()
	this.Info("下次日志切割,将在",getSurplusSecond()+10*time.Second,"秒后")
	time.Sleep( getSurplusSecond()+10*time.Second )
	this.autoCutOffLogFileHandler()
}

// 检测并切割文件
func (this *Logger) cunFile(){

	this.Println("执行一次切割")

	finfo, _ := this.file.Stat()

	if finfo.Size() <= 0 {
		return
	}
	content := make([]byte,finfo.Size())
	_, err := this.file.ReadAt(content,0)
	if err != nil {
		this.Error("Cgo LOG SYSTEM ==> Cgo log auto cut for read file error!")
		return
	}
	logArr := strings.Split(string(content),"\n")

	yesterday := yesterdayDate("/")

	regexpStr,_ := regexp.Compile("^"+yesterday+"[\\S|\\s]*$")

	var newFileByte []string

	// 分离按照日期的日志
	for i:=0; i<len(logArr); i++ {
		if regexpStr.MatchString(logArr[i] ) {
			newFileByte = append( newFileByte, logArr[i] )
			logArr = append(logArr[:i], logArr[i+1:]...)
			i--
		}
	}

	// 写入日志分割
	if len(newFileByte) > 0 {

		if !saveCutoffLogFile( this, newFileByte ){
			return
		}

		if !saveRecreatedLogFile( this, logArr ){
			return
		}
	}
}

func saveCutoffLogFile(this *Logger, newFileByte []string)bool{

	f, err := os.OpenFile(this.path + yesterdayDate("-") + "_" + this.prefix + "_cgo_log", os.O_CREATE | os.O_APPEND | os.O_RDWR, 0664)
	defer f.Close()

	if err != nil  {
		this.Error("Cgo LOG SYSTEM ==> Cgo log auto cut create new log file error!("+err.Error()+")")
		return false
	}

	_, err = f.Write([]byte(strings.Join(newFileByte,"\n")+"\n"))
	if err != nil {
		this.Error("Cgo LOG SYSTEM ==> Cgo log auto cut create new log file error!("+err.Error()+")")
		return false
	}

	return true
}

func saveRecreatedLogFile(this *Logger, logArr []string)bool{
	// 重构老日志
	f, err := os.OpenFile(this.path + this.prefix + "_cgo_log", os.O_CREATE | os.O_RDWR | os.O_TRUNC, 0664)
	defer f.Close()

	if err != nil  {
		this.Error("Cgo LOG SYSTEM ==> Cgo log recreate actived log file error!("+err.Error()+")")
		return false
	}

	_, err = f.Write([]byte(strings.Join(logArr,"\n")+"\n"))
	if err != nil {
		this.Error("Cgo LOG SYSTEM ==> Cgo log recreate actived log file error!("+err.Error()+")")
		return false
	}

	return true
}

func nowDate(tag string) string {
	return  time.Now().Format("2006"+tag+"01"+tag+"02")
}

func yesterdayDate(tag string) string {
	return time.Now().AddDate(0,0, -1).Format("2006"+tag+"01"+tag+"02")
}

func getSurplusSecond()time.Duration{
	now := time.Now()
	//return 10 * time.Second
	return time.Duration(int64(OneDay - now.Hour() * Hour - now.Minute() * Minute - now.Second() * Second)) * time.Second
}

// 数组增删字符串
func splice(arr []interface{}, index int64, replace bool, insertValue interface{})[]interface{}{
	var res []interface{}

	if insertValue != nil {
		res = append(res, insertValue)
	}

	if replace {
		res = append(res,arr[index+1:]...)
	}else{
		res = append(res,arr[index:]...)
	}

	res = append(arr[:index],res...)

	return res

}

//
func Println (v ...interface{}){

	std.Println(splice(v, 0, false, "[normal]")...)
}

func Info (v ...interface{}){
	std.Println(splice(v, 0, false, "[INFO]")...)
}

func Warn (v ...interface{}){
	std.Println(splice(v, 0, false, "[WARN]")...)
}

func Error (v ...interface{}){
	std.Println(splice(v, 0, false, "[ERROR]")...)
}

// debug模式下可以使用,设置为非debug 模式则不
func Debug (v ...interface{}){
	std.Println(splice(v, 0, false, "[DEBUG]")...)
}


// debug模式下可以使用,设置为非debug 模式则不
func Fatalln (v ...interface{}){
	std.Println(splice(v, 0, false, "[Fatal]")...)
}

func Fatal (v ...interface{}){
	std.Println(splice(v, 0, false, "[Fatal]")...)
}

