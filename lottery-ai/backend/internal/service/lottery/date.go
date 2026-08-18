package lottery

import (
	"strings"
	"time"
)

func shanghai() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

// ParseCivilDate 把开奖日期解析成日历日，避免 CST 零点写入 UTC DATE 时少一天。
func ParseCivilDate(raw string) time.Time {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "("); i > 0 {
		s = strings.TrimSpace(s[:i])
	}
	if j := strings.IndexAny(s, " T"); j > 0 {
		s = s[:j]
	}
	for _, layout := range []string{"2006-01-02", "2006-1-2"} {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t
		}
	}
	return time.Time{}
}

// DateParam 写入 PostgreSQL DATE 时用日历字符串，不依赖数据库时区。
func DateParam(t time.Time) string {
	if t.IsZero() {
		return time.Now().In(shanghai()).Format("2006-01-02")
	}
	if t.Location() == time.UTC && t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format("2006-01-02")
	}
	return t.In(shanghai()).Format("2006-01-02")
}

// CivilDateString 给前端返回 YYYY-MM-DD，避免 RFC3339 被切成前一天。
func CivilDateString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return DateParam(t)
}
