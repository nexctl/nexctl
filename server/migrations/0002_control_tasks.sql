-- 任务执行实例（手动或计划触发）。计划任务表见 0003_task_schedules.sql。

CREATE TABLE IF NOT EXISTS control_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    schedule_id BIGINT NULL DEFAULT NULL COMMENT '非空表示由计划任务触发',
    task_type VARCHAR(64) NOT NULL,
    scope_type VARCHAR(32) NOT NULL DEFAULT 'node',
    scope_value VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    progress INT NOT NULL DEFAULT 0,
    operator_id BIGINT NOT NULL DEFAULT 0,
    operator_name VARCHAR(128) NOT NULL DEFAULT '',
    payload JSON NULL,
    detail TEXT NOT NULL DEFAULT '',
    output TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    finished_at DATETIME NULL,
    INDEX idx_control_tasks_status (status),
    INDEX idx_control_tasks_created (created_at),
    INDEX idx_control_tasks_schedule (schedule_id)
);
