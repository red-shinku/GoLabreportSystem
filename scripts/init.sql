create database if not exists LabSystem
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_general_ci;

use LabSystem;

-- User table
create table Users(
--     用户标识
    UUID int unsigned AUTO_INCREMENT PRIMARY KEY,
--     身份
    identity tinyint unsigned not null comment '1=学生, 2=教师, 9=管理员',
--     学号/工号
    number varchar(16) unique not null,
--     姓名
    name varchar(16),
--     邮箱
    mail varchar(32) unique,
--     密码，加密存储
    passwd varchar(255) not null
);

-- 课程表
create table Course(
--     课程标识
    courseID int unsigned AUTO_INCREMENT PRIMARY KEY,
--     课程名
    courseName varchar(24)
);

-- 选课表，课程与班级
CREATE TABLE CourseOffering (
    offeringID int unsigned AUTO_INCREMENT PRIMARY KEY,
    courseID int unsigned not null,
--     选课班级
    className varchar(32) not null,
--     学期
    term varchar(32) not null,

    FOREIGN KEY (courseID) REFERENCES Course(courseID) ON DELETE CASCADE
);

-- 学生选课表（所选课程）
create table StudentCourse(
--     学号
    studentID varchar(16),
--     参加的课程
    offeringID int unsigned not null,

    FOREIGN KEY (offeringID) REFERENCES CourseOffering(offeringID) ON DELETE CASCADE,
    FOREIGN KEY (studentID) REFERENCES Users(number) ON DELETE CASCADE
);

-- 教师选课表（所管课程）
create table TeacherCourse(
--     工号
    teacherID varchar(16),
--     管理的课程
    offeringID int unsigned not null,

    FOREIGN KEY (offeringID) REFERENCES CourseOffering(offeringID) ON DELETE CASCADE,
    FOREIGN KEY (teacherID) REFERENCES Users(number) ON DELETE CASCADE
);

-- 实验项目表
create table Project(
--     项目标识
    projectID int unsigned AUTO_INCREMENT PRIMARY KEY,
--     所属课程ID，对应选课表的主键，允许同一个课程的不同班级拥有不同项目
    offeringID int unsigned not null,
--     项目名
    projectName varchar(64),
--     项目文件路径
    projectFilePath varchar(128),
--     开始时间
    startTime datetime,
--     截止时间
    deadline datetime,
--  项目开启状态
    isActive bool not null default true,

    FOREIGN KEY (offeringID) REFERENCES CourseOffering(offeringID) ON DELETE CASCADE
);

-- 学生实验报告表
create table StuReport(
--     实验报告标识
    stuReportID int unsigned AUTO_INCREMENT PRIMARY KEY,
--     学生ID
    studentID varchar(16) not null,
--     项目标识
    projectID int unsigned not null,
--     报告文件路径
    reportFilePath varchar(128),
--     提交时间
    submitTime datetime,
--     FOREIGN KEY (studentID) REFERENCES StudentCourse(studentID) ON DELETE CASCADE,
    FOREIGN KEY (studentID) REFERENCES Users(number) ON DELETE CASCADE,
    FOREIGN KEY (projectID) REFERENCES Project(projectID) ON DELETE CASCADE
);

