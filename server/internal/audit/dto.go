package audit

// LogItem is the audit log list response item.
type LogItem struct {
	ID        int64  `json:"id"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Detail    string `json:"detail"`
	CreatedAt string `json:"created_at"`
}

// ListResponse is the audit log list response body.
type ListResponse struct {
	Items []LogItem `json:"items"`
}
