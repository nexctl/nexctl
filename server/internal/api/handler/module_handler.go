package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nexctl/nexctl/server/internal/alert"
	"github.com/nexctl/nexctl/server/internal/audit"
	"github.com/nexctl/nexctl/server/internal/api/middleware"
	"github.com/nexctl/nexctl/server/internal/filemgr"
	"github.com/nexctl/nexctl/server/internal/task"
	"github.com/nexctl/nexctl/server/internal/upgrade"
	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/response"
)

// ModuleHandler serves placeholder list endpoints for reserved modules.
type ModuleHandler struct {
	tasks    *task.Service
	files    *filemgr.Service
	upgrades *upgrade.Service
	alerts   *alert.Service
	audits   *audit.Service
}

// NewModuleHandler creates a module handler.
func NewModuleHandler(tasks *task.Service, files *filemgr.Service, upgrades *upgrade.Service, alerts *alert.Service, audits *audit.Service) *ModuleHandler {
	return &ModuleHandler{
		tasks:    tasks,
		files:    files,
		upgrades: upgrades,
		alerts:   alerts,
		audits:   audits,
	}
}

// ListTaskSchedules 计划任务列表（新建任务下拉）。
func (h *ModuleHandler) ListTaskSchedules(w http.ResponseWriter, r *http.Request) {
	resp, err := h.tasks.ListSchedules(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}
	response.WriteOK(w, resp)
}

// CreateTaskSchedule 创建计划任务。
func (h *ModuleHandler) CreateTaskSchedule(w http.ResponseWriter, r *http.Request) {
	var req task.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid request body")
		return
	}
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, errcode.Unauthorized, "unauthorized")
		return
	}
	resp, appErr := h.tasks.CreateSchedule(r.Context(), req, claims.UserID, claims.Username)
	if appErr != nil {
		status := http.StatusBadRequest
		switch appErr.Code {
		case errcode.NotFound:
			status = http.StatusNotFound
		case errcode.Internal:
			status = http.StatusInternalServerError
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteCreated(w, resp)
}

// ListTasks handles task list（支持 query: status、keyword）。
func (h *ModuleHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("keyword")
	resp, err := h.tasks.List(r.Context(), status, keyword)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}
	response.WriteOK(w, resp)
}

// CreateTask 创建任务并下发至 Agent（若在线）。
func (h *ModuleHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req task.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid request body")
		return
	}
	claims, ok := middleware.UserClaimsFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, errcode.Unauthorized, "unauthorized")
		return
	}
	resp, appErr := h.tasks.Create(r.Context(), req, claims.UserID, claims.Username)
	if appErr != nil {
		status := http.StatusBadRequest
		switch appErr.Code {
		case errcode.NotFound:
			status = http.StatusNotFound
		case errcode.Forbidden:
			status = http.StatusForbidden
		case errcode.Internal:
			status = http.StatusInternalServerError
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteCreated(w, resp)
}

// GetTask 单条任务详情。
func (h *ModuleHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(w, http.StatusBadRequest, errcode.InvalidArgument, "invalid task id")
		return
	}
	resp, appErr := h.tasks.Get(r.Context(), id)
	if appErr != nil {
		status := http.StatusInternalServerError
		if appErr.Code == errcode.NotFound {
			status = http.StatusNotFound
		}
		response.WriteError(w, status, appErr.Code, appErr.Message)
		return
	}
	response.WriteOK(w, resp)
}

// ListFiles handles file list.
func (h *ModuleHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	resp, err := h.files.List(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}
	response.WriteOK(w, resp)
}

// ListReleases handles release list.
func (h *ModuleHandler) ListReleases(w http.ResponseWriter, r *http.Request) {
	resp, err := h.upgrades.ListReleases(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}
	response.WriteOK(w, resp)
}

// ListAlertRules handles alert rule list.
func (h *ModuleHandler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
	resp, err := h.alerts.ListRules(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}
	response.WriteOK(w, resp)
}

// ListAlertEvents handles alert event list.
func (h *ModuleHandler) ListAlertEvents(w http.ResponseWriter, r *http.Request) {
	resp, err := h.alerts.ListEvents(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}
	response.WriteOK(w, resp)
}

// ListAuditLogs handles audit log list.
func (h *ModuleHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	resp, err := h.audits.List(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, errcode.Internal, err.Error())
		return
	}
	response.WriteOK(w, resp)
}
