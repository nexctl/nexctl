package task

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// 五段式：分 时 日 月 周（UTC），与 migration 注释一致。
var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// parseCronNext 校验并解析 CRON，返回不早于 from 的 UTC 下一次触发时间。
func parseCronNext(expr string, from time.Time) (time.Time, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return time.Time{}, fmt.Errorf("cron expression is empty")
	}
	s, err := cronParser.Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	next := s.Next(from.UTC())
	return next, nil
}
