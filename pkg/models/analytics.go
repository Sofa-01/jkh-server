// pkg/models/analytics.go

package models

import "time"

// AnalyticsReportRequest — DTO для запроса генерации PDF-отчёта
type AnalyticsReportRequest struct {
	From        string   `json:"from" binding:"required"`
	To          string   `json:"to" binding:"required"`
	Charts      []string `json:"charts" binding:"omitempty,dive,oneof=status_distribution failure_frequency inspector_performance"`
	JkhUnitIDs  []int    `json:"jkh_unit_ids,omitempty"`
	DistrictIDs []int    `json:"district_ids,omitempty"`
}

// AnalyticsPreviewRequest — параметры для preview (query params)
type AnalyticsPreviewRequest struct {
	Chart     string `json:"chart" binding:"required,oneof=status_distribution failure_frequency inspector_performance"`
	From      string `json:"from" binding:"required"`
	To        string `json:"to" binding:"required"`
	JkhUnitID *int   `json:"jkh_unit_id,omitempty"`
}

// ===== Структуры для статистики по районам =====
type DistrictStat struct {
	DistrictID     int     `json:"district_id"`
	DistrictName   string  `json:"district_name"`
	BuildingsCount int     `json:"buildings_count"`
	TasksTotal     int     `json:"tasks_total"`
	TasksCompleted int     `json:"tasks_completed"`
	CompletionRate float64 `json:"completion_rate"`
}

type DistrictStatsResponse struct {
	Districts []DistrictStat `json:"districts"`
}

// ===== Структуры для статистики по инспекторам =====
type InspectorStat struct {
	InspectorID    int     `json:"inspector_id"`
	InspectorName  string  `json:"inspector_name"`
	TasksAssigned  int     `json:"tasks_assigned"`
	TasksCompleted int     `json:"tasks_completed"`
	TasksApproved  int     `json:"tasks_approved"`
	TasksRejected  int     `json:"tasks_rejected"`
	ApprovalRate   float64 `json:"approval_rate"`
}

type InspectorStatsResponse struct {
	Inspectors []InspectorStat `json:"inspectors"`
}

// ===== Структуры для сводной статистики =====
type TaskStatusStat struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type TaskTimelineStat struct {
	Date           time.Time `json:"date"`
	TasksCreated   int       `json:"tasks_created"`
	TasksCompleted int       `json:"tasks_completed"`
}

type SummaryStats struct {
	TotalTasks      int                `json:"total_tasks"`
	StatusBreakdown []TaskStatusStat   `json:"status_breakdown"`
	Timeline        []TaskTimelineStat `json:"timeline"`
	CompletionRate  float64            `json:"completion_rate"`
}
