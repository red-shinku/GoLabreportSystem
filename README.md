# 该项目目前在测试中

## 项目概述

基于 Golang 开发的 web 实验报告系统。支持项目发布、报告上传/预览、打包下载、导入学生选课表等业务。用户界面包含学生端与教师端，使用 JWT 验证登录态和身份信息。  

前后端不明显分离，页面在服务端生成并返回，使用 HTMX 完成 AJAX 请求。 

## 环境变量与配置

提供三种配置方式，按优先级从高到低：环境变量 > 配置文件 > 默认值。

**环境变量**：
```
IP_ADDR:                    服务IP地址
PORT:                       服务端口
ENABLE_TLS:                 是否开启https

JWT_SECRET:                 服务端的JWT密钥

DATABASE_ADDR:              数据库服务的IP地址 
DATABASE_PORT:              数据库服务的端口
DATABASE_USER:              数据库用户
DATABASE_PASSWD:            数据库用户密码

//数据库连接池配置
DB_MAX_OPEN_CONNS:          最大连接数
DB_MAX_IDLE_CONNS:          最大空闲连接数
DB_CONN_MAX_LIFETIME_SEC:   一个连接的最长存活时间
DB_CONN_MAX_IDLE_TIME_SEC:  空闲连接最长存活时间

ENABLE_PPROF:               开启pprof性能监控
PPROF_ADDR:                 pprof性能工具运行的地址及端口
```

**config.json**

与环境变量选项相比，JWT密钥与数据库密码设置改为填写对应的文件路径：`jwt_secret_file`、`database_passwd_file`。

```
{    
  "ip_addr": "0.0.0.0",                 //服务IP地址
  "port": "8080",                       //服务端口
  "enable_tls": false,                  //是否开始TLS
  "jwt_secret_file": "./jwt.key",       //服务端JWT密钥文件路径

  "database_addr": "127.0.0.1",         //数据库IP地址
  "database_port": "3306",              //数据库端口
  "database_user":"root",               //数据库用户名
  "database_passwd_file":"./db.pass",   //数据库密码文件路径

                                        //数据库连接池相关
  "db_max_open_conns": 512,             //最大连接数
  "db_max_idle_conns": 256,             //最大空闲连接数
  "db_conn_max_lifetime_sec":1800,      //一个连接的最长存活时间
  "db_conn_max_idle_time_sec": 300,     //空闲连接最长存活时间

  "enable_pprof": false,                //开启pprof性能监控
  "pprof_addr": "127.0.0.1:6060"        //pprof性能工具运行的地址及端口
}
```

请特别注意 pprof 监控选项的配置，仅在需要时开启，且不要将其暴露到公网。  
当作为容器部署时，config.json 内的IP相关配置均为服务在容器内的IP地址。在这时可能需要将其配置为0.0.0.0，才能映射到主机的本地地址或公网IP。

**默认值**

即配置文件写好的值。在构建镜像时该文件会被写入镜像文件系统中默认使用。

## 构建与部署

### Docker 容器部署

构建实验报告系统镜像，在项目根目录：
```
docker build . -t red-shinku/labsys:v0.3
```

启动实验报告系统的容器（powershell示例）：
```powershell
docker run -d `
  --name golabsys `
  -p 127.0.0.1:80:8080 `
  -e ENABLE_TLS=false `
  -e DATABASE_ADDR=host.docker.internal `
  -e DATABASE_PORT=3306 `
  -e DATABASE_USER=root `
  -v "$(pwd)/jwt.key:/app/jwt.key:ro" `
  -v "$(pwd)/db.pass:/app/db.pass:ro" `
  -v "$(pwd)/config.json:/app/config.json" `
  red-shinku/labsys:v0.3
  
```

或使用 docker compose：  
参考 [compose.yaml](./compose.yaml)

### 单文件部署

在项目根目录运行编译：
```
go build -o labsys
```
填写配置后运行：

```
./labsys
```

## 基准测试

准备测试中

## 设计与实现

参考[设计文档](docs/design.md)

参考[开发日志](https://study.fifseason.top/2026/04/27/LabSystem-log/)

## 注意事项

1. 启动前需确保数据库服务正常运行。  
2. 关于学生账户的注册方式：教师在导入课程信息表时，服务端会自动提取其中的学生信息，做幂等注册，密码默认与学号相同；在这之后，才完成课程信息的导入。目前对于以这种方式批量导入的课程下的项目，无法做到设定不同的截止时间。
3. 学生预览自己上传的报告的功能还未实现。（下载报告的操作目前只支持打包批量下载）

## 需要完善的地方

1. 文件系统与数据库系统的一致性保证。
2. 密码的加密存储。
3. 管理员界面的设计。
4. TLS证书管理？