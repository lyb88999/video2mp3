-- 创建数据库
CREATE DATABASE IF NOT EXISTS video_converter CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE video_converter;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    task_count INT DEFAULT 0,
    completed_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_users_username (username),
    INDEX idx_users_email (email),
    INDEX idx_users_deleted_at (deleted_at)
);

-- 创建转换任务表
CREATE TABLE IF NOT EXISTS conversion_tasks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'queued',
    title VARCHAR(255),
    description TEXT,
    input_path VARCHAR(500),
    output_path VARCHAR(500),
    thumbnail_url VARCHAR(500),
    original_url VARCHAR(1000),
    original_name VARCHAR(255),
    file_size BIGINT,
    output_size BIGINT,
    duration DOUBLE,
    progress DOUBLE DEFAULT 0,
    audio_codec VARCHAR(50) DEFAULT 'libmp3lame',
    audio_bitrate VARCHAR(20) DEFAULT '192k',
    sample_rate VARCHAR(20) DEFAULT '44100',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_conversion_tasks_user_id (user_id),
    INDEX idx_conversion_tasks_status (status),
    INDEX idx_conversion_tasks_deleted_at (deleted_at),
    INDEX idx_conversion_tasks_created_at (created_at),
    INDEX idx_conversion_tasks_type_status (type, status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 插入默认用户
INSERT IGNORE INTO users (id, username, email, password) VALUES 
('5ba6d3f8-4478-11f0-8574-74df036e1f50', 'default', 'default@example.com', 'password');

-- 设置字符集
ALTER DATABASE video_converter CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci; 
