// ===== model/statistics.go =====
package model

import (
	"one-api/common"
	"strconv"
	"time"
)

// ChannelStatistics 按渠道统计的数据结构
type ChannelStatistics struct {
	Time        string `json:"time"`
	ChannelId   int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Quota       int    `json:"quota"`
	Count       int    `json:"count"`
}

// TokenStatistics 按令牌统计的数据结构
type TokenStatistics struct {
	Time      string `json:"time"`
	TokenName string `json:"token_name"`
	Quota     int    `json:"quota"`
	Count     int    `json:"count"`
}

// UserStatistics 按用户统计的数据结构
type UserStatistics struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	Quota    int    `json:"quota"`
	Count    int    `json:"count"`
}

// GetChannelStatistics 获取按渠道统计的数据
func GetChannelStatistics(startTimestamp, endTimestamp int, username, tokenName, modelName string, channel int, group, defaultTime string) ([]ChannelStatistics, error) {
	var statistics []ChannelStatistics

	// 构建时间格式化表达式
	var timeField string
	switch defaultTime {
	case "hour":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d %H:00:00') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD HH24:00:00') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d %H:00:00', datetime(created_at, 'unixepoch')) as time_group"
		}
	case "day":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as time_group"
		}
	case "week":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%U') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-IW') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%W', datetime(created_at, 'unixepoch')) as time_group"
		}
	default:
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as time_group"
		}
	}

	// 构建查询条件
	tx := LOG_DB.Table("logs").Select(
		timeField+", channel_id, COUNT(*) as count, SUM(quota) as quota, MIN(created_at) as min_created_at",
	).Where("type = ?", 2) // LogTypeConsume = 2

	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if modelName != "" {
		tx = tx.Where("model_name LIKE ?", "%"+modelName+"%")
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where("`group` = ?", group)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}

	// 按渠道和时间分组
	tx = tx.Group("time_group, channel_id").Order("min_created_at ASC")

	// 执行查询
	var results []struct {
		TimeGroup    string `json:"time_group"`
		ChannelId    int    `json:"channel_id"`
		Count        int    `json:"count"`
		Quota        int    `json:"quota"`
		MinCreatedAt int64  `json:"min_created_at"`
	}

	err := tx.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 获取渠道名称映射
	channelIds := make([]int, 0)
	channelMap := make(map[int]string)
	for _, result := range results {
		if result.ChannelId != 0 {
			channelIds = append(channelIds, result.ChannelId)
		}
	}

	if len(channelIds) > 0 {
		var channels []struct {
			Id   int    `gorm:"column:id"`
			Name string `gorm:"column:name"`
		}
		if err = DB.Table("channels").Select("id, name").Where("id IN ?", channelIds).Find(&channels).Error; err != nil {
			return nil, err
		}
		for _, channel := range channels {
			channelMap[channel.Id] = channel.Name
		}
	}

	// 构建返回结果
	for _, result := range results {
		statistics = append(statistics, ChannelStatistics{
			Time:        result.TimeGroup,
			ChannelId:   result.ChannelId,
			ChannelName: channelMap[result.ChannelId],
			Quota:       result.Quota,
			Count:       result.Count,
		})
	}

	return statistics, nil
}

// GetTokenStatistics 获取按令牌统计的数据
func GetTokenStatistics(startTimestamp, endTimestamp int, username, tokenName, modelName string, channel int, group, defaultTime string) ([]TokenStatistics, error) {
	var statistics []TokenStatistics

	// 构建时间格式化表达式
	var timeField string
	switch defaultTime {
	case "hour":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d %H:00:00') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD HH24:00:00') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d %H:00:00', datetime(created_at, 'unixepoch')) as time_group"
		}
	case "day":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as time_group"
		}
	case "week":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%U') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-IW') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%W', datetime(created_at, 'unixepoch')) as time_group"
		}
	default:
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as time_group"
		}
	}

	// 构建查询条件
	tx := LOG_DB.Table("logs").Select(
		timeField+", token_name, COUNT(*) as count, SUM(quota) as quota, MIN(created_at) as min_created_at",
	).Where("type = ?", 2) // LogTypeConsume = 2

	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if modelName != "" {
		tx = tx.Where("model_name LIKE ?", "%"+modelName+"%")
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where("`group` = ?", group)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}

	// 按令牌和时间分组
	tx = tx.Group("time_group, token_name").Order("min_created_at ASC")

	// 执行查询
	var results []struct {
		TimeGroup    string `json:"time_group"`
		TokenName    string `json:"token_name"`
		Count        int    `json:"count"`
		Quota        int    `json:"quota"`
		MinCreatedAt int64  `json:"min_created_at"`
	}

	err := tx.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 构建返回结果
	for _, result := range results {
		statistics = append(statistics, TokenStatistics{
			Time:      result.TimeGroup,
			TokenName: result.TokenName,
			Quota:     result.Quota,
			Count:     result.Count,
		})
	}

	return statistics, nil
}

// GetUserStatistics 获取按用户统计的数据
func GetUserStatistics(startTimestamp, endTimestamp int, username, tokenName, modelName string, channel int, group, defaultTime string) ([]UserStatistics, error) {
	var statistics []UserStatistics

	// 构建时间格式化表达式
	var timeField string
	switch defaultTime {
	case "hour":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d %H:00:00') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD HH24:00:00') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d %H:00:00', datetime(created_at, 'unixepoch')) as time_group"
		}
	case "day":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as time_group"
		}
	case "week":
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%U') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-IW') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%W', datetime(created_at, 'unixepoch')) as time_group"
		}
	default:
		if common.UsingMySQL {
			timeField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as time_group"
		} else if common.UsingPostgreSQL {
			timeField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD') as time_group"
		} else {
			// SQLite
			timeField = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as time_group"
		}
	}

	// 构建查询条件
	tx := LOG_DB.Table("logs").Select(
		timeField+", username, COUNT(*) as count, SUM(quota) as quota, MIN(created_at) as min_created_at",
	).Where("type = ?", 2) // LogTypeConsume = 2

	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if modelName != "" {
		tx = tx.Where("model_name LIKE ?", "%"+modelName+"%")
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where("`group` = ?", group)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}

	// 按用户和时间分组
	tx = tx.Group("time_group, username").Order("min_created_at ASC")

	// 执行查询
	var results []struct {
		TimeGroup    string `json:"time_group"`
		Username     string `json:"username"`
		Count        int    `json:"count"`
		Quota        int    `json:"quota"`
		MinCreatedAt int64  `json:"min_created_at"`
	}

	err := tx.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 构建返回结果
	for _, result := range results {
		statistics = append(statistics, UserStatistics{
			Time:     result.TimeGroup,
			Username: result.Username,
			Quota:    result.Quota,
			Count:    result.Count,
		})
	}

	return statistics, nil
}

// formatTime 根据时间粒度格式化时间
func formatTime(timestamp int64, defaultTime string) string {
	t := time.Unix(timestamp, 0)
	switch defaultTime {
	case "hour":
		return t.Format("2006-01-02 15:00")
	case "day":
		return t.Format("2006-01-02")
	case "week":
		_, week := t.ISOWeek()
		return t.Format("2006-01") + "-W" + strconv.Itoa(week)
	default:
		return t.Format("2006-01-02")
	}
}

// ===== 如果你还需要日志汇总功能,添加以下代码 =====

// LogSummary 日志汇总结构
type LogSummary struct {
	Date             string  `json:"date"`
	TokenName        string  `json:"token_name"`
	ModelName        string  `json:"model_name"`
	TotalRequests    int     `json:"total_requests"`
	SuccessRequests  int     `json:"success_requests"`
	SuccessRate      float64 `json:"success_rate"`
	TotalTokens      int64   `json:"total_tokens"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalQuota       int     `json:"total_quota"`
}

// buildSuccessRateSQL 构建成功率计算SQL,兼容不同数据库
func buildSuccessRateSQL() string {
	if common.UsingMySQL {
		// MySQL: 直接使用数值除法,不使用CAST AS FLOAT
		return "ROUND(SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2)"
	} else if common.UsingPostgreSQL {
		// PostgreSQL: 使用CAST AS FLOAT
		return "ROUND(CAST(SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) AS FLOAT) * 100.0 / COUNT(*), 2)"
	} else {
		// SQLite: 使用CAST AS REAL
		return "ROUND(CAST(SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) AS REAL) * 100.0 / COUNT(*), 2)"
	}
}

// GetMonthlySummary 获取月度汇总数据
func GetMonthlySummary(startTimestamp, endTimestamp int64, page, pageSize int) ([]LogSummary, int64, error) {
	var summaries []LogSummary
	var total int64

	// 构建时间格式化字段
	var dateField, groupBy string
	if common.UsingMySQL {
		dateField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m') as date"
		groupBy = "date, token_name, model_name"
	} else if common.UsingPostgreSQL {
		dateField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM') as date"
		groupBy = "date, token_name, model_name"
	} else {
		dateField = "strftime('%Y-%m', datetime(created_at, 'unixepoch')) as date"
		groupBy = "date, token_name, model_name"
	}

	// 构建成功率计算SQL
	successRateSQL := buildSuccessRateSQL()

	// 完整的SELECT语句
	selectSQL := dateField + `,
		token_name,
		model_name,
		COUNT(*) as total_requests,
		SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) as success_requests,
		` + successRateSQL + ` as success_rate,
		SUM(prompt_tokens + completion_tokens) as total_tokens,
		SUM(prompt_tokens) as prompt_tokens,
		SUM(completion_tokens) as completion_tokens,
		SUM(quota) as total_quota`

	// 构建子查询
	subQuery := LOG_DB.Table("logs").
		Select(selectSQL).
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Group(groupBy)

	// 获取总数
	if err := LOG_DB.Table("(?) as summary", subQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := LOG_DB.Table("(?) as summary", subQuery).
		Order("date DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&summaries).Error; err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}
