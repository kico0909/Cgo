package log

import (
	olog "log"
	slog "log"
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

type Logger struct {
	mode 				string				// 输出模式
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

	f, err := os.OpenFile(
						path + prefix + "_cgo_log",
						os.O_CREATE|os.O_RDWR|os.O_APPEND,
						0777)

	if err != nil {
		var logger = olog.New(os.Stderr, "", 3)

		resLogger = &Logger{
			file: f,
			autoCutOff: autoCut,
			logger: logger,
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

	slog.Println("功能初始化: Cgo日志系统 --- [ ok ]")



	if autoCut {
		go resLogger.autoCutOffLogFileHandler()
	}

	return resLogger
}

// 设置日志的debug模式, 默认关闭
func (this *Logger) SetDebugMode (key bool){
	this.debugMode = key
}

func (this *Logger) Println (v ...interface{}){
	tmp := make([]interface{}, len(v),len(v))
	tmp[0] = "[Normal]"
	tmp = append(tmp,v...)
	this.logger.Println(tmp...)
}

func (this *Logger) Info (v ...interface{}){
	tmp := make([]interface{}, len(v),len(v))
	tmp[0] = "[INFO]"
	tmp = append(tmp,v...)
	this.logger.Println(tmp...)
}

func (this *Logger) Warn (v ...interface{}){
	tmp := make([]interface{}, len(v),len(v))
	tmp[0] = "[WARN]"
	tmp = append(tmp,v...)
	this.logger.Println(tmp...)
}

func (this *Logger) Error (v ...interface{}){
	tmp := make([]interface{}, len(v),len(v))
	tmp[0] = "[ERROR]"
	tmp = append(tmp,v...)
	this.logger.Println(tmp...)
}

// debug模式下可以使用,设置为非debug 模式则不
func (this *Logger) Debug (v ...interface{}){
	if !this.debugMode {
		return
	}
	tmp := make([]interface{}, len(v),len(v))
	tmp[0] = "[DEBUG]"
	tmp = append(tmp,v...)
	this.logger.Println(tmp...)
}

// 开一个定时线程执行文件分割, 按天执行
func (this *Logger) autoCutOffLogFileHandler(){

	if this.mode == "Terminal" {
		return
	}

	// 启动时先进行一次分割
	this.cunFile()
	time.Sleep( time.Duration(getSurplusSecond()+10) * time.Second )

	go this.autoCutOffLogFileHandler()

}

// 检测并切割文件
func (this *Logger) cunFile(){

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

func getSurplusSecond()int64{
	now := time.Now()
	return int64(OneDay - now.Hour() * Hour - now.Minute() * Minute - now.Second() * Second)
}

