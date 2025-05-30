-- 创建数据库
CREATE DATABASE IF NOT EXISTS distributed_job DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE distributed_job;

-- 创建部门表
CREATE TABLE IF NOT EXISTS department (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(500),
    parent_id BIGINT,
    status TINYINT NOT NULL DEFAULT 1,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建权限表
CREATE TABLE IF NOT EXISTS permission (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    status TINYINT NOT NULL DEFAULT 1,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建角色表
CREATE TABLE IF NOT EXISTS role (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    status TINYINT NOT NULL DEFAULT 1,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建角色权限关联表
CREATE TABLE IF NOT EXISTS role_permission (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_role_perm (role_id, permission_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建用户表
CREATE TABLE IF NOT EXISTS user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(100) NOT NULL,
    real_name VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    phone VARCHAR(20),
    department_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    status TINYINT NOT NULL DEFAULT 1,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_username (username),
    INDEX idx_department (department_id),
    INDEX idx_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建任务表
CREATE TABLE IF NOT EXISTS task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    cron_expression VARCHAR(50),
    handler VARCHAR(255) NOT NULL,
    params TEXT,
    status TINYINT NOT NULL DEFAULT 0,
    max_retry INT NOT NULL DEFAULT 0,
    retry_count INT NOT NULL DEFAULT 0,
    last_execute_time DATETIME,
    next_execute_time DATETIME,
    creator_id BIGINT NOT NULL,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_creator (creator_id),
    INDEX idx_status (status),
    INDEX idx_next_exec (next_execute_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建执行记录表
CREATE TABLE IF NOT EXISTS record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id BIGINT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    status TINYINT NOT NULL DEFAULT 0,
    result TEXT,
    error TEXT,
    executor VARCHAR(100),
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_task_id (task_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入初始数据
-- 部门数据
INSERT INTO department (id, name, description, parent_id, status) 
VALUES (1, '系统管理部', '负责系统管理的部门', NULL, 1);

INSERT INTO department (id, name, description, parent_id, status) 
VALUES 
(2, '研发部', '负责产品研发的部门', NULL, 1),
(3, '运维部', '负责系统运维的部门', NULL, 1),
(4, '测试部', '负责产品测试的部门', NULL, 1),
(5, '产品部', '负责产品规划的部门', NULL, 1),
(6, '前端组', '负责前端开发', 2, 1),
(7, '后端组', '负责后端开发', 2, 1),
(8, '移动组', '负责移动端开发', 2, 1),
(9, '数据库组', '负责数据库维护', 3, 1),
(10, '网络组', '负责网络维护', 3, 1);

-- 权限数据
INSERT INTO permission (id, name, code, description, status) 
VALUES 
(1, '系统管理', 'SYSTEM_ADMIN', '系统管理权限', 1),
(2, '任务创建', 'TASK_CREATE', '创建任务权限', 1),
(3, '任务修改', 'TASK_UPDATE', '修改任务权限', 1),
(4, '任务删除', 'TASK_DELETE', '删除任务权限', 1),
(5, '任务执行', 'TASK_EXECUTE', '执行任务权限', 1),
(6, '任务查看', 'TASK_VIEW', '查看任务权限', 1),
(7, '用户管理', 'USER_MANAGE', '管理用户权限', 1),
(8, '角色管理', 'ROLE_MANAGE', '管理角色权限', 1),
(9, '部门管理', 'DEPT_MANAGE', '管理部门权限', 1),
(10, '日志查看', 'LOG_VIEW', '查看日志权限', 1);

-- 角色数据
INSERT INTO role (id, name, description, status) 
VALUES 
(1, '系统管理员', '拥有所有权限的角色', 1),
(2, '开发人员', '拥有任务创建、修改、执行权限', 1),
(3, '运维人员', '拥有任务执行、查看权限', 1),
(4, '测试人员', '拥有任务查看权限', 1),
(5, '部门主管', '拥有部门内所有任务管理权限', 1);

-- 角色权限关联数据
INSERT INTO role_permission (role_id, permission_id) 
VALUES 
-- 系统管理员拥有所有权限
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8), (1, 9), (1, 10),
-- 开发人员权限
(2, 2), (2, 3), (2, 5), (2, 6),
-- 运维人员权限
(3, 5), (3, 6), (3, 10),
-- 测试人员权限
(4, 6),
-- 部门主管权限
(5, 2), (5, 3), (5, 4), (5, 5), (5, 6);

-- -- 用户数据 (admin123)
-- INSERT INTO user (id, username, password, real_name, email, phone, department_id, role_id, status) 
-- VALUES 
-- (1, 'admin', '$2a$10$Pe.qJQONn5mJqGa5j3B1tub/IkREfoxDS0A85wcnY8kTWw6PIg7iO', '系统管理员', 'admin@example.com', '13800000000', 1, 1, 1),
-- (2, 'dev1', '$2a$10$Pe.qJQONn5mJqGa5j3B1tub/IkREfoxDS0A85wcnY8kTWw6PIg7iO', '开发者1', 'dev1@example.com', '13800000001', 7, 2, 1),
-- (3, 'dev2', '$2a$10$Pe.qJQONn5mJqGa5j3B1tub/IkREfoxDS0A85wcnY8kTWw6PIg7iO', '开发者2', 'dev2@example.com', '13800000002', 6, 2, 1),
-- (4, 'ops1', '$2a$10$Pe.qJQONn5mJqGa5j3B1tub/IkREfoxDS0A85wcnY8kTWw6PIg7iO', '运维1', 'ops1@example.com', '13800000003', 3, 3, 1),
-- (5, 'ops2', '$2a$10$Pe.qJQONn5mJqGa5j3B1tub/IkREfoxDS0A85wcnY8kTWw6PIg7iO', '运维2', 'ops2@example.com', '13800000004', 9, 3, 1),
-- (6, 'test1', '$2a$10$Pe.qJQONn5mJqGa5j3B1tub/IkREfoxDS0A85wcnY8kTWw6PIg7iO', '测试1', 'test1@example.com', '13800000005', 4, 4, 1),
-- (7, 'manager1', '$2a$10$Pe.qJQONn5mJqGa5j3B1tub/IkREfoxDS0A85wcnY8kTWw6PIg7iO', '研发主管', 'manager1@example.com', '13800000006', 2, 5, 1),
-- (8, 'manager2', '$2a$10$Pe.qJQONn5mJqGa5j3B1tub/IkREfoxDS0A85wcnY8kTWw6PIg7iO', '运维主管', 'manager2@example.com', '13800000007', 3, 5, 1);

-- -- 任务数据
-- INSERT INTO task (id, name, description, cron_expression, handler, params, status, max_retry, creator_id, next_execute_time) 
-- VALUES 
-- (1, '数据库备份', '每天凌晨自动备份数据库', '0 0 2 * * ?', 'database.backup', '{"dbName": "distributed_job", "backupPath": "/backup"}', 1, 3, 1, '2025-04-28 02:00:00'),
-- (2, '日志清理', '每周清理一次过期日志', '0 0 3 ? * MON', 'log.cleanup', '{"days": 30, "logPath": "/logs"}', 1, 2, 1, '2025-05-05 03:00:00'),
-- (3, '数据统计', '每天统计系统使用数据', '0 30 23 * * ?', 'stats.daily', '{"modules": ["user", "task", "record"]}', 1, 2, 2, '2025-04-27 23:30:00'),
-- (4, '健康检查', '每小时检查系统健康状态', '0 0 * * * ?', 'system.healthcheck', '{"endpoints": ["api", "db", "cache"]}', 1, 5, 4, '2025-04-27 23:00:00'),
-- (5, '缓存刷新', '每6小时刷新一次系统缓存', '0 0 */6 * * ?', 'cache.refresh', '{"cacheNames": ["user", "permission"]}', 1, 3, 5, '2025-04-28 00:00:00'),
-- (6, '报表生成', '每月1日生成上月报表', '0 0 1 1 * ?', 'report.monthly', '{"month": "previous", "format": "pdf"}', 1, 3, 7, '2025-05-01 01:00:00'),
-- (7, '数据同步', '与第三方系统数据同步', '0 0 4 * * ?', 'data.sync', '{"target": "erp", "tables": ["product", "order"]}', 0, 5, 2, null),
-- (8, '临时测试任务', '用于测试系统的临时任务', null, 'test.temp', '{"mode": "test"}', 2, 0, 6, null);

-- -- 执行记录数据
-- INSERT INTO record (id, task_id, start_time, end_time, status, result, error, executor) 
-- VALUES 
-- (1, 1, '2025-04-26 02:00:00', '2025-04-26 02:03:25', 1, '{"backupSize": "156MB", "files": 1}', null, 'worker-01'),
-- (2, 2, '2025-04-20 03:00:00', '2025-04-20 03:01:12', 1, '{"deletedFiles": 128, "freedSpace": "34MB"}', null, 'worker-02'),
-- (3, 3, '2025-04-26 23:30:00', '2025-04-26 23:31:45', 1, '{"processedRecords": 5830}', null, 'worker-01'),
-- (4, 4, '2025-04-26 22:00:00', '2025-04-26 22:00:08', 1, '{"status": "healthy", "services": 3}', null, 'worker-03'),
-- (5, 5, '2025-04-26 18:00:00', '2025-04-26 18:00:32', 1, '{"clearedEntries": 256}', null, 'worker-02'),
-- (6, 7, '2025-04-26 04:00:00', '2025-04-26 04:02:18', 0, null, '{"code": "CONN_ERROR", "message": "Failed to connect to ERP API"}', 'worker-01'),
-- (7, 7, '2025-04-26 04:10:00', '2025-04-26 04:12:30', 1, '{"syncedRecords": 1240}', null, 'worker-01'),
-- (8, 8, '2025-04-26 15:00:00', '2025-04-26 15:00:05', 1, '{"result": "success"}', null, 'worker-03'),
-- (9, 4, '2025-04-26 23:00:00', '2025-04-26 23:00:09', 1, '{"status": "healthy", "services": 3}', null, 'worker-02'),
-- (10, 1, '2025-04-27 02:00:00', '2025-04-27 02:03:48', 1, '{"backupSize": "158MB", "files": 1}', null, 'worker-01');