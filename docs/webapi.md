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
set-cookie: jwt{user_id, identity}; path=/; httponly; samesite=lax;
```

## 页面

* / 用户主界面

* /sessions 登录界面 

## 课程资源

* /api/v1/courses

### 导入实验课信息表(导入excel表)

POST /api/v1/courses

身份：教师（JWT 提供 `teacherID`）。  
请求体：`multipart/form-data`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `filename` | file | 课程信息表（xlsx）。首行表头固定为 `ID \| SNO \| Sname \| <项目名 1> \| <项目名 2> \| ...` |
| `courseName` | string | 课程名（必填） |
| `className` | string | 班级名；空字符串时服务端兜底为 `'-'` |
| `term` | string | 学期标识（必填，如 `2025-2026-1`） |
| `closeTime` | string | 项目默认截止时间（RFC3339） |

返回：
- 201：导入完成。已存在的学生 / 课程 / 开课 / 选课 / 项目复用，不报错。
- 400：表单字段缺失、closeTime 格式非法、Excel 表头/项目列异常（`ErrSheetFormat`）。
- 401：未登录。
- 500：数据库查询或写入失败。

## 班级资源

* /api/v1/offeringclass

### 新建项目（教师）

POST /api/v1/offeringclass/{offeringId}

**Body**

```
{
  "projectname": "...",
  "closeTime": "..."
}
```

## 项目资源

* /api/v1/projects

### 获取项目列表视图

GET /api/v1/projects

身份在cookie中获取。  
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

### 上传/重传项目要求文件（教师）

PUT /api/v1/projects/{projectId}/requirement

Form-data：

```
filename: xxx
```

### 删除项目（教师）

DELETE /api/v1/projects/{projectID}


### 提交实验报告（学生）

POST /api/v1/projects/{projectId}/submissions

From-data:
```
filename: xxx
```

学生ID将从cookie获取

## 实验报告操作


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
