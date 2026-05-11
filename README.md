
## 项目概述

基于 Golang 开发的 web 实验报告系统。支持项目发布、报告上传/预览、打包下载、导入学生选课表等业务。用户界面包含学生端与教师端，使用 JWT 验证登录态和身份信息。  

前后端不明显分离，页面在服务端生成并返回，使用 HTMX 完成 AJAX 请求。 

## 环境变量与配置

提供三种配置方式，按优先级从高到低：环境变量 > 配置文件 > 默认值。

**环境变量**：
```
IP_ADDR:        服务IP地址
PORT:           服务端口
ENABLE_TLS:     是否开启https

JWT_SECRET:     服务端的JWT密钥

DATABASE_ADDR:  数据库服务的IP地址 
DATABASE_PORT:  数据库服务的端口
DATABASE_USER:  数据库用户
DATABASE_PASSWD:数据库用户密码
```

**config.json**

与环境变量选项相比，JWT密钥与数据库密码设置改为填写对应的文件路径：`jwt_secret_file`、`database_passwd_file`。

```
{
  "ip_addr": "0.0.0.0",
  "port": "8080",
  "enable_tls": false,
  "jwt_secret_file": "./jwt.key",

  "database_addr": "127.0.0.1",
  "database_port": "3306",
  "database_user":"root",
  "database_passwd_file":"./db.pass"
}
```

**默认值**

即配置文件写好的值。在构建镜像时该文件会被写入镜像文件系统中默认使用。

## 构建与部署

### Docker 容器部署

构建实验报告系统镜像，在项目根目录：
```
docker build . -t labsys:v0.1
```

启动实验报告系统的容器（示例）：
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
  labsys:v0.1
```

或使用 docker compose：  
参考 [compose.yaml](./compose.yaml)

### 单文件部署

在项目根目录运行编译：
```
go build -o <文件名>
```

## 基准测试

准备测试中

## 设计与实现

参考[设计文档](docs/design.md)

参考[开发日志](https://study.fifseason.top/2026/04/27/LabSystem-log/)

## 注意事项

1. 启动前需确保数据库服务正常运行。
