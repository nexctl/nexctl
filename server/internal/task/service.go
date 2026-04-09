package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/nexctl/nexctl/server/internal/audit"
	"github.com/nexctl/nexctl/server/internal/model"
	"github.com/nexctl/nexctl/server/internal/repository"
	"github.com/nexctl/nexctl/server/internal/ws"
	"github.com/nexctl/nexctl/server/pkg/errcode"
	"go.uber.org/zap"
)

// Service 任务编排与下发。
type Service struct {
	repo      repository.TaskRepository
	schedules repository.ScheduleRepository
	nodes     repository.NodeRepository
	wsSvc     *ws.Service
	audit     *audit.Service
	logger    *zap.Logger
}

// NewService creates a task service.
func NewService(repo repository.TaskRepository, schedules repository.ScheduleRepository, nodes repository.NodeRepository, wsSvc *ws.Service, audit *audit.Service, logger *zap.Logger) *Service {
	return &Service{repo: repo, schedules: schedules, nodes: nodes, wsSvc: wsSvc, audit: audit, logger: logger}
}

// Ping 占位。
func (s *Service) Ping(context.Context) error {
	return nil
}

// List 支持 query：status、keyword。
func (s *Service) List(ctx context.Context, status, keyword string) (*ListResponse, error) {
	rows, err := s.repo.List(ctx, status, keyword, 500)
	if err != nil {
		return nil, err
	}
	items := make([]ListItem, 0, len(rows))
	for _, t := range rows {
		items = append(items, s.toListItem(t))
	}
	return &ListResponse{Items: items}, nil
}

// ListSchedules 返回计划任务列表（新建任务下拉等）。
func (s *Service) ListSchedules(ctx context.Context) (*ScheduleListResponse, error) {
	rows, err := s.schedules.List(ctx, 500)
	if err != nil {
		return nil, err
	}
	items := make([]ScheduleListItem, 0, len(rows))
	for _, sch := range rows {
		items = append(items, ScheduleListItem{
			ID:        sch.ID,
			Name:      sch.Name,
			CronExpr:  sch.CronExpr,
			TaskType:  sch.TaskType,
			Scope:     formatScopeParts(sch.ScopeType, sch.ScopeValue),
			Detail:    sch.Detail,
			Enabled:   sch.Enabled,
			NextRunAt: sch.NextRunAt.UTC().Format(time.RFC3339),
		})
	}
	return &ScheduleListResponse{Items: items}, nil
}

// CreateSchedule 新建计划任务（写入 task_schedules，并计算 next_run_at）。
func (s *Service) CreateSchedule(ctx context.Context, req CreateScheduleRequest, operatorID int64, operatorName string) (*ScheduleListItem, *errcode.AppError) {
	req.TaskType = strings.TrimSpace(strings.ToLower(req.TaskType))
	req.ScopeType = strings.TrimSpace(strings.ToLower(req.ScopeType))
	req.ScopeValue = strings.TrimSpace(req.ScopeValue)
	req.Name = strings.TrimSpace(req.Name)
	req.CronExpr = strings.TrimSpace(req.CronExpr)
	req.Detail = strings.TrimSpace(req.Detail)

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if req.TaskType == "" || req.ScopeType == "" {
		return nil, errcode.New(errcode.InvalidArgument, "task_type and scope_type are required")
	}
	if req.TaskType != "echo" && req.TaskType != "shell_command" {
		return nil, errcode.New(errcode.InvalidArgument, "task_type must be echo or shell_command")
	}
	if req.ScopeType != "node" {
		return nil, errcode.New(errcode.InvalidArgument, "scope_type must be node")
	}
	if req.TaskType == "shell_command" && req.Detail == "" {
		return nil, errcode.New(errcode.InvalidArgument, "detail is required for shell_command")
	}
	nodeID, err := strconv.ParseInt(req.ScopeValue, 10, 64)
	if err != nil || nodeID <= 0 {
		return nil, errcode.New(errcode.InvalidArgument, "scope_value must be a positive node id")
	}
	nodeRow, err := s.nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "load node failed", err)
	}
	if nodeRow == nil {
		return nil, errcode.New(errcode.NotFound, "node not found")
	}
	_ = nodeRow

	nextAt, err := parseCronNext(req.CronExpr, time.Now().UTC())
	if err != nil {
		return nil, errcode.New(errcode.InvalidArgument, "invalid cron_expr: "+err.Error())
	}

	rec := &model.TaskSchedule{
		Name:         req.Name,
		CronExpr:     req.CronExpr,
		TaskType:     req.TaskType,
		ScopeType:    req.ScopeType,
		ScopeValue:   req.ScopeValue,
		Detail:       req.Detail,
		Enabled:      enabled,
		OperatorID:   operatorID,
		OperatorName: strings.TrimSpace(operatorName),
		NextRunAt:    nextAt,
	}
	if err := s.schedules.Create(ctx, rec); err != nil {
		s.logger.Warn("create task_schedules row failed", zap.Error(err))
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1146 {
			return nil, errcode.New(errcode.InvalidArgument, "数据库中不存在 task_schedules 表。请在 MySQL 中执行 server/migrations/0003_task_schedules.sql；若 control_tasks 也不存在，请先执行 0002_control_tasks.sql（Docker 新库需挂载全部迁移或手动导入）")
		}
		return nil, errcode.Wrap(errcode.Internal, fmt.Sprintf("create schedule failed: %v", err), err)
	}

	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "user",
		ActorID:      strconv.FormatInt(operatorID, 10),
		ActorName:    operatorName,
		Action:       "task_schedule.create",
		ResourceType: "task_schedule",
		ResourceID:   strconv.FormatInt(rec.ID, 10),
		Detail:       req.CronExpr + " node=" + req.ScopeValue,
	})

	sch2, err := s.schedules.GetByID(ctx, rec.ID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "reload schedule failed", err)
	}
	if sch2 == nil {
		return nil, errcode.New(errcode.Internal, "schedule missing after create")
	}
	item := ScheduleListItem{
		ID:        sch2.ID,
		Name:      sch2.Name,
		CronExpr:  sch2.CronExpr,
		TaskType:  sch2.TaskType,
		Scope:     formatScopeParts(sch2.ScopeType, sch2.ScopeValue),
		Detail:    sch2.Detail,
		Enabled:   sch2.Enabled,
		NextRunAt: sch2.NextRunAt.UTC().Format(time.RFC3339),
	}
	return &item, nil
}

func (s *Service) toListItem(t *model.ControlTask) ListItem {
	return ListItem{
		ID:         t.ID,
		ScheduleID: scheduleIDPtr(t.ScheduleID),
		Type:       t.TaskType,
		Scope:      formatScope(t),
		Status:     t.Status,
		Progress:   t.Progress,
		Operator:   t.OperatorName,
		CreatedAt:  t.CreatedAt.UTC().Format(time.RFC3339),
		FinishedAt: finishedAtStr(t.FinishedAt),
		Detail:     t.Detail,
		Output:     t.Output,
	}
}

func scheduleIDPtr(nt sql.NullInt64) *int64 {
	if !nt.Valid {
		return nil
	}
	v := nt.Int64
	return &v
}

func finishedAtStr(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.UTC().Format(time.RFC3339)
}

func formatScope(t *model.ControlTask) string {
	return formatScopeParts(t.ScopeType, t.ScopeValue)
}

func formatScopeParts(scopeType, scopeValue string) string {
	st := strings.TrimSpace(scopeType)
	sv := strings.TrimSpace(scopeValue)
	if st == "node" && sv != "" {
		return "node:" + sv
	}
	if sv == "" {
		return st
	}
	return st + ":" + sv
}

// Create 创建任务并尝试向 Agent 下发 task_dispatch（节点类任务）。
func (s *Service) Create(ctx context.Context, req CreateRequest, operatorID int64, operatorName string) (*DetailResponse, *errcode.AppError) {
	var taskType, scopeType, scopeValue, detail string
	var schedID sql.NullInt64

	if req.ScheduleID != nil && *req.ScheduleID > 0 {
		sch, err := s.schedules.GetByID(ctx, *req.ScheduleID)
		if err != nil {
			return nil, errcode.Wrap(errcode.Internal, "load schedule failed", err)
		}
		if sch == nil {
			return nil, errcode.New(errcode.NotFound, "schedule not found")
		}
		if !sch.Enabled {
			return nil, errcode.New(errcode.InvalidArgument, "schedule is disabled")
		}
		taskType = strings.TrimSpace(strings.ToLower(sch.TaskType))
		scopeType = strings.TrimSpace(strings.ToLower(sch.ScopeType))
		scopeValue = strings.TrimSpace(sch.ScopeValue)
		detail = sch.Detail
		schedID = sql.NullInt64{Valid: true, Int64: sch.ID}
	} else {
		taskType = strings.TrimSpace(strings.ToLower(req.TaskType))
		scopeType = strings.TrimSpace(strings.ToLower(req.ScopeType))
		scopeValue = strings.TrimSpace(req.ScopeValue)
		detail = req.Detail
	}

	if taskType == "" || scopeType == "" {
		return nil, errcode.New(errcode.InvalidArgument, "task_type and scope_type are required")
	}
	if taskType != "echo" && taskType != "shell_command" {
		return nil, errcode.New(errcode.InvalidArgument, "task_type must be echo or shell_command")
	}
	if scopeType != "node" {
		return nil, errcode.New(errcode.InvalidArgument, "scope_type must be node")
	}
	nodeID, err := strconv.ParseInt(scopeValue, 10, 64)
	if err != nil || nodeID <= 0 {
		return nil, errcode.New(errcode.InvalidArgument, "scope_value must be a positive node id")
	}
	nodeRow, err := s.nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "load node failed", err)
	}
	if nodeRow == nil {
		return nil, errcode.New(errcode.NotFound, "node not found")
	}
	_ = nodeRow

	rec := &model.ControlTask{
		ScheduleID:   schedID,
		TaskType:     taskType,
		ScopeType:    scopeType,
		ScopeValue:   scopeValue,
		Status:       "pending",
		Progress:     0,
		OperatorID:   operatorID,
		OperatorName: strings.TrimSpace(operatorName),
		Detail:       detail,
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, errcode.Wrap(errcode.Internal, "create task failed", err)
	}

	s.tryDispatch(ctx, rec, nodeID)

	auditDetail := taskType + " scope=node:" + scopeValue
	if schedID.Valid {
		auditDetail += " schedule_id=" + strconv.FormatInt(schedID.Int64, 10)
	}
	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "user",
		ActorID:      strconv.FormatInt(operatorID, 10),
		ActorName:    operatorName,
		Action:       "task.create",
		ResourceType: "task",
		ResourceID:   strconv.FormatInt(rec.ID, 10),
		Detail:       auditDetail,
	})

	t2, err := s.repo.GetByID(ctx, rec.ID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "reload task failed", err)
	}
	if t2 == nil {
		return nil, errcode.New(errcode.Internal, "task missing after create")
	}
	d := s.toDetail(t2)
	return &d, nil
}

func (s *Service) tryDispatch(ctx context.Context, rec *model.ControlTask, nodeID int64) {
	if !s.wsSvc.AgentHub.Online(nodeID) {
		_ = s.repo.UpdateResult(ctx, rec.ID, "failed", 0, "agent offline")
		return
	}
	payload := ws.TaskDispatchPayload{
		TaskID:   rec.ID,
		TaskType: rec.TaskType,
		Detail:   rec.Detail,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		_ = s.repo.UpdateResult(ctx, rec.ID, "failed", 0, "marshal dispatch payload failed")
		return
	}
	msg := ws.Message{
		Type:      ws.MessageTypeTaskDispatch,
		RequestID: strconv.FormatInt(rec.ID, 10),
		Timestamp: time.Now().UTC(),
		Payload:   raw,
	}
	if err := s.wsSvc.AgentHub.Send(nodeID, msg); err != nil {
		s.logger.Warn("task dispatch send failed", zap.Int64("task_id", rec.ID), zap.Error(err))
		_ = s.repo.UpdateResult(ctx, rec.ID, "failed", 0, err.Error())
		return
	}
	_ = s.repo.UpdateDispatch(ctx, rec.ID, "running", 5)
}

// Get 单条详情。
func (s *Service) Get(ctx context.Context, id int64) (*DetailResponse, *errcode.AppError) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "get task failed", err)
	}
	if t == nil {
		return nil, errcode.New(errcode.NotFound, "task not found")
	}
	d := s.toDetail(t)
	return &d, nil
}

func (s *Service) toDetail(t *model.ControlTask) DetailResponse {
	return DetailResponse{
		ID:         t.ID,
		ScheduleID: scheduleIDPtr(t.ScheduleID),
		Type:       t.TaskType,
		Scope:      formatScope(t),
		Status:     t.Status,
		Progress:   t.Progress,
		Operator:   t.OperatorName,
		CreatedAt:  t.CreatedAt.UTC().Format(time.RFC3339),
		FinishedAt: finishedAtStr(t.FinishedAt),
		Detail:     t.Detail,
		Output:     t.Output,
		ScopeType:  t.ScopeType,
		ScopeValue: t.ScopeValue,
	}
}

// ApplyAgentReport Agent 上报任务执行结果。
func (s *Service) ApplyAgentReport(ctx context.Context, nodeID int64, p ws.TaskReportPayload) *errcode.AppError {
	if p.TaskID <= 0 {
		return errcode.New(errcode.InvalidArgument, "invalid task_id")
	}
	t, err := s.repo.GetByID(ctx, p.TaskID)
	if err != nil {
		return errcode.Wrap(errcode.Internal, "get task failed", err)
	}
	if t == nil {
		return errcode.New(errcode.NotFound, "task not found")
	}
	if strings.TrimSpace(t.ScopeType) != "node" || strings.TrimSpace(t.ScopeValue) != strconv.FormatInt(nodeID, 10) {
		return errcode.New(errcode.Forbidden, "task scope does not match agent node")
	}
	st := strings.TrimSpace(strings.ToLower(p.Status))
	if st != "success" && st != "failed" && st != "running" {
		st = "failed"
	}
	prog := p.Progress
	if prog < 0 {
		prog = 0
	}
	if prog > 100 {
		prog = 100
	}
	out := p.Output
	if st == "running" {
		_ = s.repo.UpdateDispatch(ctx, p.TaskID, st, prog)
		return nil
	}
	if err := s.repo.UpdateResult(ctx, p.TaskID, st, prog, out); err != nil {
		return errcode.Wrap(errcode.Internal, "update task result failed", err)
	}
	return nil
}
