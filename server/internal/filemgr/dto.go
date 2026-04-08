package filemgr

// ListItem is the file list response item.
type ListItem struct {
	ID                int64  `json:"id"`
	FileName          string `json:"file_name"`
	SizeText          string `json:"size_text"`
	Checksum          string `json:"checksum"`
	CreatedAt         string `json:"created_at"`
	DistributionCount int    `json:"distribution_count"`
}

// ListResponse is the file list response body.
type ListResponse struct {
	Items []ListItem `json:"items"`
}
