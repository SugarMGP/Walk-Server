# Walk-server
本项目为精弘毅行后端代码，包括报名系统(微信网页端)、扫码打卡系统和后台管理系统（微信小程序）。

> 本文档中的所有路径都用 / 开头，这个根目录指代项目所在的目录

### 功能说明
当前版本为 V1.0.2

主要新增功能
1. 学生教师注册时新增统一验证功能
2. 管理端查询各线路已报名的队伍情况
3. 管理端查询各线路人员情况和excel文件导出
4. 生成测试队伍进行以上功能的测试
5. 重构部分管理端扫码打卡功能和若干管理功能

上一版本：V1.0.1

已完成事项
需要做的事情有
1. 随机队伍
2. 调整部分路由的关系，上中间件能更加方便

### 数据返回说明
一定要使用 /utility/response.go 下的函数来返回数据

### 项目文件说明
```text
./
├── LICENSE
├── README.md
├── config     (配置文件目录)
│        ├── config.example.yaml
│        └── config.yaml
├── controller (控制器 -> 每个路由对应的回调函数)
│        ├── basic.go
│        ├── ······
│        ├── team.go
│        └── user.go
├── go.mod (go 项目文件)
├── go.sum (go 依赖版本控制文件)
├── main.go
├── middleware (中间件 -> 在请求传入到控制函数前对请求数据做一些处理)
│        ├── auth.go
│        └── validity.go
├── model      (数据库模型 -> 用来描述数据库表的结构体)
│        ├── person.go
│        ├── team.go
│        └── team_count.go
├── utility    (工具函数 -> 一些常用的工具函数 比如说获取当前是毅行报名第几天的函数)
│        ├── crypto.go
│        ├── date.go
│        ├── initial
│        │       ├── init_config.go
│        │       ├── init_db.go
│        │       └── init_router.go
│        ├── jwt.go
│        ├── response.go
│        ├── serve.go
│        └── wechat.go
└── walk-server
```

### 如何启动
#### 开启 go module 和换源（如果拉取依赖较慢才考虑这个）
[https://goproxy.cn](https://goproxy.cn) 

请按照这个网站的说明开启 go module 特性, 并切换 go proxy


#### 配置配置文件
配置文件样例为 /config/config.example.yaml 文件
配置文件默认在 /config 目录下的 config.yaml（日后会添加上动态生成配置文件，和读取不同位置的配置文件的功能）
```
cp /config/config.example.yaml /config/config.yaml
```
再将自己的mysql和redis的配置写入 config.yaml 文件中,(微信相关配置在单后端测试中可不填写，如需要请找上一届负责人获取配置文件)
```
vim /config/config.yaml
```
注：配置文件 /config/config.yaml 不可以上传到 Github 上，否则重要开发信息泄漏，后果自负

#### 测试运行
##### 微信网页端
先通过apifox的Basic的测试登录拿到token(参数openid可以随便输)
拿到token后，在apifox的Register的学生/教师注册接口，新增学生或教师用户，前者输入的open_id即为用户主键
接着其它就和正常没啥差别了
##### 微信小程序
在数据库新增数据，wechat_open_id可以不填，然后通过apifox的Admin中的测试登录登录获取token即可。
管理端密钥在配置文件里。
接着其它就和正常没啥差别了

#### 调整 mysql 并调大 Linux 内核支持的最大文件句柄数（**服务上线时**，对 Linux 的调整，测试时不用管)

首先调整 mysql 的线程数和最大连接数

根据 **2021** 年毅行报名的经验:

> 最大连接数可以调整为 900
> 
> 线程池数量可以调整为 64

然后请根据这个网站调整 Linux 服务器内核支持打开的最大文件数

[http://woshub.com/too-many-open-files-error-linux/](http://woshub.com/too-many-open-files-error-linux/)

#### 编译项目
```bash
make build-linux
```

#### 后台运行
```bash
nohup ./程序名 &
```

> Go 编译会自动安装依赖