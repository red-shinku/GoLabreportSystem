# 后端接口

## 登录

* POST /api/v1/sessions

**Body**
```
{
  "userNumber": "...",
  "password": "..."
}
```
**返回**
```
{
  "token": "...",
  "user": {...}
}
```

## 项目操作

* /api/v1/projects

### 获取项目列表视图

GET /api/v1/projects

身份放在token中。  
不同身份的视图结构不一样。

### 下载/预览项目要求文件

GET /api/v1/projects/{projectId}/requirement

**Query**
```
?preview=true
```

### 查看学生完成情况（教师）

GET /api/v1/projects/{projectId}/submissions

### 开启/关闭项目（教师）

PATCH /api/v1/projects/{projectId}

```
{
  "status": "open" // or "closed"
}
```

### 新建项目（教师）

POST /api/v1/projects

**Body**

```
{
  "offeringId": 1,
  "name": "...",
  "closeTime": "..."
}
```

### 上传/重传项目要求文件（教师）

PUT /api/v1/projects/{projectId}/requirement

Form-data：

```
file: xxx
```

### 删除项目（教师）

DELETE /api/v1/projects/{projectID}


## 实验报告操作

### 提交实验报告（学生）

POST /api/v1/projects/{projectId}/submissions

From-data:
```
file: xxx
```

学生ID将从token获取

### 下载/预览实验报告

GET /api/v1/submissions/{submissionId}/file

**Query**
```
?preview=true
```

### 打包下载实验报告（教师）

GET /api/v1/projects/{projectId}/submissions/archive

## 用户数据操作

### 绑定邮箱

POST /api/v1/users/me/email

### 忘记密码

请求重置：

POST /api/v1/password-resets

执行重置：

PUT /api/v1/password-resets/{token}
