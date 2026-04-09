package migrations

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInitSQLDefinesExpectedTables(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	raw, err := os.ReadFile(filepath.Join(dir, "0001_init.sql"))
	if err != nil {
		t.Fatalf("read 0001_init.sql: %v", err)
	}
	s := strings.ToLower(string(raw))
	for _, name := range []string{
		"create table",
		"users",
		"roles",
		"user_roles",
		"nodes",
		"install_tokens",
		"audit_logs",
		"node_runtime_states",
		"agent_id",
		"max_uses",
		"actor_name",
		"enrollment_token_hash",
		"enrollment_expires_at",
	} {
		if !strings.Contains(s, name) {
			t.Errorf("0001_init.sql should mention %q", name)
		}
	}
}

func TestTaskSchedulesMigrationExists(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	raw, err := os.ReadFile(filepath.Join(dir, "0003_task_schedules.sql"))
	if err != nil {
		t.Fatalf("read 0003_task_schedules.sql: %v", err)
	}
	s := strings.ToLower(string(raw))
	for _, name := range []string{"create table", "task_schedules", "cron_expr", "next_run_at"} {
		if !strings.Contains(s, name) {
			t.Errorf("0003_task_schedules.sql should mention %q", name)
		}
	}
}
