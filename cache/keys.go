package cache

import (
	"fmt"
	"strconv"
)

const (
	// DailyRankKey 每日排行
	DailyRankKey = "rank:daily"
)

// VideoViewKey 视频点击数的Key
func VideoViewKey(id uint) string {
	return fmt.Sprintf("view:view:%s", strconv.Itoa(int(id)))
}
