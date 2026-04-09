package task

// ListItem 任务列表行。
type ListItem struct {
	ID         int64  `json:"id"`
	ScheduleID *int64 `json:"schedule_id,omitempty"`
	Type       string `json:"type"`
	Scope      string `json:"scope"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	Operator   string `json:"operator"`
	CreatedAt  string `json:"created_at"`
	FinishedAt string `json:"finished_at,omitempty"`
	Detail     string `json:"detail"`
	Output     string `json:"output,omitempty"`
}

// ListResponse 列表接口。
type ListResponse struct {
	Items []ListItem `json:"items"`
}

// CreateRequest 创建任务。
// 若设置 schedule_id（>0），则按该计划任务的类型与范围创建一次执行实例，忽略 task_type / scope_* / detail。
type CreateRequest struct {
	ScheduleID *int64 `json:"schedule_id,omitempty"`
	TaskType   string `json:"task_type"`
	ScopeType  string `json:"scope_type"`
	ScopeValue string `json:"scope_value"`
	Detail     string `json:"detail"`
}

// DetailResponse 单条任务详情。
type DetailResponse struct {
	ID         int64  `json:"id"`
	ScheduleID *int64 `json:"schedule_id,omitempty"`
	Type       string `json:"type"`
	Scope      string `json:"scope"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	Operator   string `json:"operator"`
	CreatedAt  string `json:"created_at"`
	FinishedAt string `json:"finished_at,omitempty"`
	Detail     string `json:"detail"`
	Output     string `json:"output,omitempty"`
	ScopeType  string `json:"scope_type"`
	ScopeValue string `json:"scope_value"`
}

// ScheduleListItem 计划任务列表行（用于下拉与展示）。
type ScheduleListItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CronExpr  string `json:"cron_expr"`
	TaskType  string `json:"task_type"`
	Scope     string `json:"scope"`
	Detail    string `json:"detail"`
	Enabled   bool   `json:"enabled"`
	NextRunAt string `json:"next_run_at"`
}

// CreateScheduleRequest POST /task-schedules。
type CreateScheduleRequest struct {
	Name       string `json:"name"`
	CronExpr   string `json:"cron_expr"`
	TaskType   string `json:"task_type"`
	ScopeType  string `json:"scope_type"`
	ScopeValue string `json:"scope_value"`
	Detail     string `json:"detail"`
	Enabled    *bool  `json:"enabled,omitempty"`
}

// ScheduleListResponse GET /task-schedules。
type ScheduleListResponse struct {
	Items []ScheduleListItem `json:"items"`
}
