package alert

// RuleItem is the alert rule list response item.
type RuleItem struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
}

// EventItem is the alert event list response item.
type EventItem struct {
	ID        int64  `json:"id"`
	Severity  string `json:"severity"`
	NodeName  string `json:"node_name"`
	Summary   string `json:"summary"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// RuleListResponse is the alert rule list response body.
type RuleListResponse struct {
	Items []RuleItem `json:"items"`
}

// EventListResponse is the alert event list response body.
type EventListResponse struct {
	Items []EventItem `json:"items"`
}
