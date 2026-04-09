package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/nexctl/nexctl/server/internal/model"
)

// TaskRepository 控制面任务持久化。
type TaskRepository interface {
	Create(ctx context.Context, t *model.ControlTask) error
	GetByID(ctx context.Context, id int64) (*model.ControlTask, error)
	List(ctx context.Context, status, keyword string, limit int) ([]*model.ControlTask, error)
	UpdateDispatch(ctx context.Context, id int64, status string, progress int) error
	UpdateResult(ctx context.Context, id int64, status string, progress int, output string) error
}

// MySQLTaskRepository implements TaskRepository.
type MySQLTaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a task repository.
func NewTaskRepository(db *sql.DB) *MySQLTaskRepository {
	return &MySQLTaskRepository{db: db}
}

const taskCols = `id, schedule_id, task_type, scope_type, scope_value, status, progress, operator_id, operator_name, payload, detail, output, created_at, updated_at, finished_at`

func (r *MySQLTaskRepository) Create(ctx context.Context, t *model.ControlTask) error {
	const q = `INSERT INTO control_tasks (schedule_id, task_type, scope_type, scope_value, status, progress, operator_id, operator_name, payload, detail, output, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	var payload any
	if t.Payload.Valid {
		payload = t.Payload.String
	}
	var sid any
	if t.ScheduleID.Valid {
		sid = t.ScheduleID.Int64
	}
	res, err := r.db.ExecContext(ctx, q, sid, t.TaskType, t.ScopeType, t.ScopeValue, t.Status, t.Progress, t.OperatorID, t.OperatorName, payload, t.Detail, t.Output)
	if err != nil {
		return fmt.Errorf("insert control_tasks: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	t.ID = id
	return nil
}

func (r *MySQLTaskRepository) GetByID(ctx context.Context, id int64) (*model.ControlTask, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+taskCols+` FROM control_tasks WHERE id = ?`, id)
	t, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

func scanTask(row *sql.Row) (*model.ControlTask, error) {
	var t model.ControlTask
	err := row.Scan(
		&t.ID, &t.ScheduleID, &t.TaskType, &t.ScopeType, &t.ScopeValue, &t.Status, &t.Progress,
		&t.OperatorID, &t.OperatorName, &t.Payload, &t.Detail, &t.Output,
		&t.CreatedAt, &t.UpdatedAt, &t.FinishedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// List status 为空或 all 时不过滤状态；keyword 在 task_type、scope_value、detail 中模糊匹配。
func (r *MySQLTaskRepository) List(ctx context.Context, status, keyword string, limit int) ([]*model.ControlTask, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	var args []any
	qs := `SELECT ` + taskCols + ` FROM control_tasks WHERE 1=1`
	if strings.TrimSpace(status) != "" && strings.ToLower(strings.TrimSpace(status)) != "all" {
		qs += ` AND status = ?`
		args = append(args, strings.TrimSpace(status))
	}
	kw := strings.TrimSpace(keyword)
	if kw != "" {
		pat := "%" + kw + "%"
		qs += ` AND (task_type LIKE ? OR scope_value LIKE ? OR detail LIKE ? OR CAST(id AS CHAR) LIKE ?)`
		args = append(args, pat, pat, pat, pat)
	}
	qs += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, qs, args...)
	if err != nil {
		return nil, fmt.Errorf("list control_tasks: %w", err)
	}
	defer rows.Close()

	var out []*model.ControlTask
	for rows.Next() {
		var t model.ControlTask
		if err := rows.Scan(
			&t.ID, &t.ScheduleID, &t.TaskType, &t.ScopeType, &t.ScopeValue, &t.Status, &t.Progress,
			&t.OperatorID, &t.OperatorName, &t.Payload, &t.Detail, &t.Output,
			&t.CreatedAt, &t.UpdatedAt, &t.FinishedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

func (r *MySQLTaskRepository) UpdateDispatch(ctx context.Context, id int64, status string, progress int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE control_tasks SET status = ?, progress = ?, updated_at = NOW() WHERE id = ?`, status, progress, id)
	return err
}

func (r *MySQLTaskRepository) UpdateResult(ctx context.Context, id int64, status string, progress int, output string) error {
	const q = `UPDATE control_tasks SET status = ?, progress = ?, output = ?,
		finished_at = CASE WHEN ? IN ('success','failed','cancelled') THEN NOW() ELSE NULL END,
		updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, status, progress, output, status, id)
	return err
}
