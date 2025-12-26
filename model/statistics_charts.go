package model

import (
	"one-api/common"
)

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

// GetDailySummary 获取日度汇总数据
func GetDailySummary(startTimestamp, endTimestamp int64, page, pageSize int) ([]LogSummary, int64, error) {
	var summaries []LogSummary
	var total int64

	// 构建时间格式化字段
	var dateField, groupBy string
	if common.UsingMySQL {
		dateField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as date"
		groupBy = "date, token_name, model_name"
	} else if common.UsingPostgreSQL {
		dateField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD') as date"
		groupBy = "date, token_name, model_name"
	} else {
		dateField = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as date"
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

// GetHourlySummary 获取小时汇总数据
func GetHourlySummary(startTimestamp, endTimestamp int64, page, pageSize int) ([]LogSummary, int64, error) {
	var summaries []LogSummary
	var total int64

	// 构建时间格式化字段
	var dateField, groupBy string
	if common.UsingMySQL {
		dateField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d %H:00:00') as date"
		groupBy = "date, token_name, model_name"
	} else if common.UsingPostgreSQL {
		dateField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-MM-DD HH24:00:00') as date"
		groupBy = "date, token_name, model_name"
	} else {
		dateField = "strftime('%Y-%m-%d %H:00:00', datetime(created_at, 'unixepoch')) as date"
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

// GetWeeklySummary 获取周度汇总数据
func GetWeeklySummary(startTimestamp, endTimestamp int64, page, pageSize int) ([]LogSummary, int64, error) {
	var summaries []LogSummary
	var total int64

	// 构建时间格式化字段
	var dateField, groupBy string
	if common.UsingMySQL {
		dateField = "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%U') as date"
		groupBy = "date, token_name, model_name"
	} else if common.UsingPostgreSQL {
		dateField = "TO_CHAR(TO_TIMESTAMP(created_at), 'YYYY-IW') as date"
		groupBy = "date, token_name, model_name"
	} else {
		dateField = "strftime('%Y-%W', datetime(created_at, 'unixepoch')) as date"
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
