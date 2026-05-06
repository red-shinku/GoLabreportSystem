# 后端接口

## 登录

* POST /api/v1/login

登录数据在表单中:
```
usernumber：学号/工号
passwd： 	密码
```


## 项目操作

* /api/v1/project

### 获取项目列表视图

GET /api/v1/project/list?role={}

student or teacher

### 下载/预览项目要求文件

GET /api/v1/project/file?projectID={}

### 查看学生完成情况（教师）

GET /api/v1/project/submissions?projectID={}

### 开启/关闭项目（教师）

PATCH /api/v1/project/{projectID}

后端将反转项目的状态

### 新建项目（教师）

POST /api/v1/project/new/{OfferingID}

新项目的数据在表单中

```
	projectname ：项目名
	filename    ：文件名
	closetime   ：截止时间
```

### 重新上传项目要求文件（教师）

POST /api/vi/project/file/{projectID}

表单：

```
	filename ：文件名
```

### 删除项目（教师）

DELETE /api/v1/project/{projectID}


## 实验报告操作

### 提交实验报告（学生）

POST /api/v1/report/submit/{projectID}

学生ID将从token获取

### 下载/预览实验报告

GET /api/v1/report/file?reportID={uint}&preview={bool}

### 打包下载实验报告（教师）

GET /api/v1/report/files?projectID={uint}


## 用户数据操作




