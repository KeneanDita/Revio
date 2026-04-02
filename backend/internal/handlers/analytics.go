package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/revio/backend/internal/middleware"
	"github.com/revio/backend/internal/models"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	db *gorm.DB
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

type AnalyticsResponse struct {
	Summary     SummaryMetrics    `json:"summary"`
	DailyTrend  []DailyDataPoint  `json:"daily_trend"`
	TopReviewers []ReviewerMetric `json:"top_reviewers"`
	TopAuthors  []AuthorMetric    `json:"top_authors"`
	StatusBreakdown StatusBreakdown `json:"status_breakdown"`
}

type SummaryMetrics struct {
	TotalPRs           int64    `json:"total_prs"`
	OpenPRs            int64    `json:"open_prs"`
	MergedPRs          int64    `json:"merged_prs"`
	ClosedPRs          int64    `json:"closed_prs"`
	AvgMergeTimeHours  *float64 `json:"avg_merge_time_hours"`
	AvgReviewTimeHours *float64 `json:"avg_review_time_hours"`
	MergeRate          float64  `json:"merge_rate"`
}

type DailyDataPoint struct {
	Date   string `json:"date"`
	Opened int64  `json:"opened"`
	Merged int64  `json:"merged"`
	Closed int64  `json:"closed"`
}

type ReviewerMetric struct {
	Reviewer      string  `json:"reviewer"`
	ReviewerAvatar string `json:"reviewer_avatar"`
	ReviewCount   int64   `json:"review_count"`
	ApprovalCount int64   `json:"approval_count"`
}

type AuthorMetric struct {
	Author      string `json:"author"`
	AuthorAvatar string `json:"author_avatar"`
	PRCount     int64  `json:"pr_count"`
	MergedCount int64  `json:"merged_count"`
}

type StatusBreakdown struct {
	Open   int64 `json:"open"`
	Merged int64 `json:"merged"`
	Closed int64 `json:"closed"`
}

func (h *AnalyticsHandler) Get(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	from, to := parseDateRange(c)
	repoID := c.Query("repo_id")
	author := c.Query("author")

	baseQuery := h.db.
		Model(&models.PullRequest{}).
		Joins("JOIN repositories ON repositories.id = pull_requests.repo_id").
		Where("repositories.user_id = ?", userID).
		Where("pull_requests.github_created_at BETWEEN ? AND ?", from, to)

	if repoID != "" {
		baseQuery = baseQuery.Where("pull_requests.repo_id = ?", repoID)
	}
	if author != "" {
		baseQuery = baseQuery.Where("pull_requests.author ILIKE ?", "%"+author+"%")
	}

	// Summary counts
	var totalPRs, openPRs, mergedPRs, closedPRs int64
	baseQuery.Count(&totalPRs)
	baseQuery.Where("pull_requests.status = ?", "open").Count(&openPRs)
	baseQuery.Where("pull_requests.status = ?", "merged").Count(&mergedPRs)
	baseQuery.Where("pull_requests.status = ?", "closed").Count(&closedPRs)

	// Average merge time
	var avgMergeResult struct{ Avg *float64 }
	h.db.Raw(`
		SELECT AVG(EXTRACT(EPOCH FROM (pull_requests.merged_at - pull_requests.github_created_at)) / 3600) as avg
		FROM pull_requests
		JOIN repositories ON repositories.id = pull_requests.repo_id
		WHERE repositories.user_id = ?
		AND pull_requests.status = 'merged'
		AND pull_requests.github_created_at BETWEEN ? AND ?
	`, userID, from, to).Scan(&avgMergeResult)

	// Average time to first review
	var avgReviewResult struct{ Avg *float64 }
	h.db.Raw(`
		SELECT AVG(EXTRACT(EPOCH FROM (pull_requests.first_review_at - pull_requests.github_created_at)) / 3600) as avg
		FROM pull_requests
		JOIN repositories ON repositories.id = pull_requests.repo_id
		WHERE repositories.user_id = ?
		AND pull_requests.first_review_at IS NOT NULL
		AND pull_requests.github_created_at BETWEEN ? AND ?
	`, userID, from, to).Scan(&avgReviewResult)

	mergeRate := 0.0
	if totalPRs > 0 {
		mergeRate = float64(mergedPRs) / float64(totalPRs) * 100
	}

	// Daily trend (last N days)
	var dailyTrend []DailyDataPoint
	h.db.Raw(`
		SELECT
			DATE(github_created_at) as date,
			COUNT(*) FILTER (WHERE TRUE) as opened,
			COUNT(*) FILTER (WHERE status = 'merged') as merged,
			COUNT(*) FILTER (WHERE status = 'closed') as closed
		FROM pull_requests
		JOIN repositories ON repositories.id = pull_requests.repo_id
		WHERE repositories.user_id = ?
		AND pull_requests.github_created_at BETWEEN ? AND ?
		GROUP BY DATE(github_created_at)
		ORDER BY DATE(github_created_at) ASC
	`, userID, from, to).Scan(&dailyTrend)

	// Top reviewers
	var topReviewers []ReviewerMetric
	h.db.Raw(`
		SELECT
			r.reviewer,
			r.reviewer_avatar,
			COUNT(*) as review_count,
			COUNT(*) FILTER (WHERE r.state = 'approved') as approval_count
		FROM reviews r
		JOIN pull_requests pr ON pr.id = r.pr_id
		JOIN repositories repo ON repo.id = pr.repo_id
		WHERE repo.user_id = ?
		AND pr.github_created_at BETWEEN ? AND ?
		GROUP BY r.reviewer, r.reviewer_avatar
		ORDER BY review_count DESC
		LIMIT 10
	`, userID, from, to).Scan(&topReviewers)

	// Top authors
	var topAuthors []AuthorMetric
	h.db.Raw(`
		SELECT
			pr.author,
			pr.author_avatar_url as author_avatar,
			COUNT(*) as pr_count,
			COUNT(*) FILTER (WHERE pr.status = 'merged') as merged_count
		FROM pull_requests pr
		JOIN repositories repo ON repo.id = pr.repo_id
		WHERE repo.user_id = ?
		AND pr.github_created_at BETWEEN ? AND ?
		GROUP BY pr.author, pr.author_avatar_url
		ORDER BY pr_count DESC
		LIMIT 10
	`, userID, from, to).Scan(&topAuthors)

	c.JSON(http.StatusOK, AnalyticsResponse{
		Summary: SummaryMetrics{
			TotalPRs:           totalPRs,
			OpenPRs:            openPRs,
			MergedPRs:          mergedPRs,
			ClosedPRs:          closedPRs,
			AvgMergeTimeHours:  avgMergeResult.Avg,
			AvgReviewTimeHours: avgReviewResult.Avg,
			MergeRate:          mergeRate,
		},
		DailyTrend:   dailyTrend,
		TopReviewers: topReviewers,
		TopAuthors:   topAuthors,
		StatusBreakdown: StatusBreakdown{
			Open:   openPRs,
			Merged: mergedPRs,
			Closed: closedPRs,
		},
	})
}

func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now().UTC()
	to := now

	fromStr := c.Query("from")
	toStr := c.Query("to")

	from := now.AddDate(0, 0, -30)

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t.Add(24*time.Hour - time.Second)
		}
	}

	return from, to
}
