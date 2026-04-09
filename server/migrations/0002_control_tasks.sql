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


-- CRON 计划任务（UTC、五段式 cron：分 时 日 月 周）。依赖 control_tasks 已存在（见 0002）。

CREATE TABLE IF NOT EXISTS task_schedules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL DEFAULT '',
    cron_expr VARCHAR(128) NOT NULL COMMENT '5-field cron: min hour dom mon dow, UTC',
    task_type VARCHAR(64) NOT NULL,
    scope_type VARCHAR(32) NOT NULL DEFAULT 'node',
    scope_value VARCHAR(255) NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT '',
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    operator_id BIGINT NOT NULL DEFAULT 0,
    operator_name VARCHAR(128) NOT NULL DEFAULT '',
    next_run_at DATETIME NOT NULL,
    last_run_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_task_schedules_due (enabled, next_run_at)
);
