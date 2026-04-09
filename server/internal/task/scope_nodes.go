package task

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// parseScopeNodeIDs 解析 scope_value：单个 id 或逗号分隔的 id 列表，去重、升序。
func parseScopeNodeIDs(scopeValue string) ([]int64, error) {
	s := strings.TrimSpace(scopeValue)
	if s == "" {
		return nil, fmt.Errorf("empty scope_value")
	}
	parts := strings.Split(s, ",")
	seen := make(map[int64]struct{})
	var ids []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid node id %q", p)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	if len(ids) == 0 {
		return nil, fmt.Errorf("no valid node ids")
	}
	return ids, nil
}

func joinScopeNodeIDs(ids []int64) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}
