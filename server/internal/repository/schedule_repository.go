package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nexctl/nexctl/server/internal/model"
)

// ScheduleRepository 计划任务（CRON）持久化。
type ScheduleRepository interface {
	List(ctx context.Context, limit int) ([]*model.TaskSchedule, error)
	GetByID(ctx context.Context, id int64) (*model.TaskSchedule, error)
	Create(ctx context.Context, sch *model.TaskSchedule) error
}

const scheduleCols = `id, name, cron_expr, task_type, scope_type, scope_value, detail, enabled, operator_id, operator_name, next_run_at, last_run_at, created_at, updated_at`

// MySQLScheduleRepository implements ScheduleRepository.
type MySQLScheduleRepository struct {
	db *sql.DB
}

// NewScheduleRepository creates a schedule repository.
func NewScheduleRepository(db *sql.DB) *MySQLScheduleRepository {
	return &MySQLScheduleRepository{db: db}
}

// List returns recent schedules (newest first).
func (r *MySQLScheduleRepository) List(ctx context.Context, limit int) ([]*model.TaskSchedule, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+scheduleCols+` FROM task_schedules ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list task_schedules: %w", err)
	}
	defer rows.Close()

	var out []*model.TaskSchedule
	for rows.Next() {
		t, err := scanScheduleRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Create inserts a new schedule row.
func (r *MySQLScheduleRepository) Create(ctx context.Context, sch *model.TaskSchedule) error {
	const q = `INSERT INTO task_schedules (name, cron_expr, task_type, scope_type, scope_value, detail, enabled, operator_id, operator_name, next_run_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	var en int8
	if sch.Enabled {
		en = 1
	}
	// 使用 UTC 字符串写入 DATETIME，避免部分 DSN 未设 parseTime=true 时无法绑定 time.Time。
	nextAt := sch.NextRunAt.UTC().Format("2006-01-02 15:04:05")
	if sch.NextRunAt.IsZero() {
		nextAt = time.Now().UTC().Format("2006-01-02 15:04:05")
	}
	res, err := r.db.ExecContext(ctx, q,
		sch.Name, sch.CronExpr, sch.TaskType, sch.ScopeType, sch.ScopeValue, sch.Detail,
		en, sch.OperatorID, sch.OperatorName, nextAt,
	)
	if err != nil {
		return fmt.Errorf("insert task_schedules: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	sch.ID = id
	return nil
}

// GetByID loads one schedule by primary key.
func (r *MySQLScheduleRepository) GetByID(ctx context.Context, id int64) (*model.TaskSchedule, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+scheduleCols+` FROM task_schedules WHERE id = ?`, id)
	t, err := scanScheduleRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

func scanScheduleRow(row interface{ Scan(dest ...any) error }) (*model.TaskSchedule, error) {
	var t model.TaskSchedule
	var enabled int8
	err := row.Scan(
		&t.ID, &t.Name, &t.CronExpr, &t.TaskType, &t.ScopeType, &t.ScopeValue, &t.Detail,
		&enabled, &t.OperatorID, &t.OperatorName, &t.NextRunAt, &t.LastRunAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	t.Enabled = enabled != 0
	return &t, nil
}
