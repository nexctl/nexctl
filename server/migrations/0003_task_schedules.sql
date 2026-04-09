-- CRON 计划任务（UTC、五段式 cron：分 时 日 月 周）。
-- 依赖 control_tasks 已存在（见 0002_control_tasks.sql）。

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
