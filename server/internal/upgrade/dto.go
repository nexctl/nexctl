package upgrade

// ReleaseItem is the release list response item.
type ReleaseItem struct {
	ID            int64  `json:"id"`
	Version       string `json:"version"`
	Channel       string `json:"channel"`
	Notes         string `json:"notes"`
	CreatedAt     string `json:"created_at"`
	RolloutStatus string `json:"rollout_status"`
}

// ReleaseListResponse is the release list response body.
type ReleaseListResponse struct {
	Items []ReleaseItem `json:"items"`
}
