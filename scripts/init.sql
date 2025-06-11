-- 创建数据库和用户
CREATE DATABASE IF NOT EXISTS go_job CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建应用用户（如果不存在）
CREATE USER IF NOT EXISTS 'go_job'@'%' IDENTIFIED BY 'password';

-- 授权
GRANT ALL PRIVILEGES ON go_job.* TO 'go_job'@'%';
FLUSH PRIVILEGES;

-- 使用数据库
USE go_job;

-- 设置时区
SET time_zone = '+08:00';

-- 初始化系统默认数据
-- 等待应用启动后通过GORM自动迁移创建表结构
