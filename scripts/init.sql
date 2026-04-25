create database if not exits LabSystem
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_general_ci;

use LabSystem;

-- User table
create table Users(
--     用户标识
    UUID int unsigned AUTO_INCREMENT PRIMARY KEY,
--     身份
    identity tinyint unsigned,
--     学号/工号
    number varchar(16) unique not null,
--     邮箱
    mail varchar(24) unique not null,
--     密码，加密存储
    passwd varchar(24)
);

-- 实验项目表
create table Project(
--     项目标识
    projectID int unsigned AUTO_INCREMENT PRIMARY KEY,
--     项目名
    projectName varchar(24),
--     所属课程ID
    courseID FOREIGN KEY references Course(courseID),
--     项目文件路径
    projectFilePath varchar(24),
--     开始时间
    startTime time,
--     截止时间
    deadline time,
-- 项目开启状态
    isActive bool
);

-- 学生实验报告表
create table StuReport(
--     实验报告标识
    stuReportID int unsigned AUTO_INCREMENT PRIMARY KEY,
--     课程标识
    courseID FOREIGN KEY references Course(courseID),
--     项目标识
    projectID FOREIGN KEY references Project(projectID),
--     报告文件路径
    reportFilePath varchar(24),
--     提交时间
    submitTime time,
);

-- 课程表
create table Course(
--     课程标识
    courseID int unsigned AUTO_INCREMENT PRIMARY KEY,
--     课程名
    courseName varchar(24),
);

-- 学生信息表
create table Student(
--     学号
    stuID int unsigned,
--     参加的课程
    courseID FOREIGN KEY references Course(courseID)
);

-- 教师信息表
create table Teacher(
--     工号
    stuID int unsigned,
--     参加的课程
    courseID FOREIGN KEY references Course(courseID)
);
