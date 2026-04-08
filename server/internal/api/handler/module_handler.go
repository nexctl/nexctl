package handler

import (
	"net/http"

	"github.com/nexctl/nexctl/server/internal/alert"
	"github.com/nexctl/nexctl/server/internal/audit"
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

// ListTasks handles task list.
func (h *ModuleHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	resp, err := h.tasks.List(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, errcode.Internal, err.Error())
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
