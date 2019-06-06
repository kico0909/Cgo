# Chunk - 网站 或 http/https 协议 web 应用服务器
---


###第三方包
- 使用了一些三方现成的包降低了开发难度(仓库里不包含,请自行下载)
    - 路由: gorilla/mux
    - mysql驱动: go-sql-driver
    - redis: garyburd/redigo
    - session: astaxie/beego/session
    - golang.org 的一些库
        - crypto
        - net
        - sys
        - text

### 启动方式    
- 服务器采用命令行形式启动
    - start|stop 
    - -d : 服务器静默模式启动


### 配置文件
- 配置文件使用ini文件格式撰写,请遵循ini文件格式
- 需要将 conf_example.ini(服务器配置文件) 改名为 conf.ini 并完成相应的配置才能启动服务器


### 编译
- 基于main.go文件编译服务器的启动文件


### 一些默认的目录设置
- 目录: ./static/ 默认保存所有 以 http://www.domain.com/static/xxx 访问的静态资源
- 目录: ./template/ 保存所有需要渲染的页面模板(模板以html为文件后缀) 


### https
- https 证书不一定放在服务器的目录内(certs目录), 在配置文件内给出证书绝对地址即可


### 开发说明
- 全局变量 Cgo 开头引用:
    - Cgo.Config : 配置文件
    - Cgo.Router : 路由设置
    - Cgo.Session : Session
    - Cgo.Redis : redis缓存
    - Cgo.Mysql : 数据库
    - log : 日志调用(可将日志输出至文件,可分割日志文件)
- 全局类型 Cgo 提供自定义类型:
    - Cgo.RouterHandler : Cgo的路由类型
    - Cgo.TableModule : Cgo 的数据模型
    - Cgo.CasUserinfoType : Cgo的cas验证器的用户信息类型

### 更细说明
2019-04-20: 增加路由实例上的一些简便方法
2019-06-06: 优化日志性能,mysql兼容性,一些BUG
