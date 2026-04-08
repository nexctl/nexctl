package task

// ListItem is the task list response item.
type ListItem struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	Scope      string `json:"scope"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	Operator   string `json:"operator"`
	CreatedAt  string `json:"created_at"`
	FinishedAt string `json:"finished_at,omitempty"`
	Detail     string `json:"detail"`
}

// ListResponse is the task list response body.
type ListResponse struct {
	Items []ListItem `json:"items"`
}
