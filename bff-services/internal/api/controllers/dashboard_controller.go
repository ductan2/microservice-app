package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/types"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	userService   services.UserService
	lessonService services.LessonService
}

func NewDashboardController(userService services.UserService, lessonService services.LessonService) *DashboardController {
	return &DashboardController{
		userService:   userService,
		lessonService: lessonService,
	}
}

func (d *DashboardController) GetSummary(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	var (
		sessionStats dto.SessionStatsResponse
		lessonStats  dto.UserLessonStatsResponse
		points       dto.UserPointsResponse
		streak       dto.UserStreakResponse
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		resp, err := d.userService.GetSessionStats(ctx, userID, email, sessionID)
		if err != nil {
			return fmt.Errorf("session stats request: %w", err)
		}
		data, err := decodeServiceResponse[dto.SessionStatsResponse](resp)
		if err != nil {
			return fmt.Errorf("session stats decode: %w", err)
		}
		sessionStats = *data
		return nil
	})

	g.Go(func() error {
		resp, err := d.lessonService.GetUserLessonStats(ctx, userID, email, sessionID)
		if err != nil {
			return fmt.Errorf("lesson stats request: %w", err)
		}
		data, err := decodeServiceResponse[dto.UserLessonStatsResponse](resp)
		if err != nil {
			return fmt.Errorf("lesson stats decode: %w", err)
		}
		lessonStats = *data
		return nil
	})

	g.Go(func() error {
		resp, err := d.lessonService.GetUserPoints(ctx, userID)
		if err != nil {
			return fmt.Errorf("user points request: %w", err)
		}
		data, err := decodeServiceResponse[dto.UserPointsResponse](resp)
		if err != nil {
			return fmt.Errorf("user points decode: %w", err)
		}
		points = *data
		return nil
	})

	g.Go(func() error {
		resp, err := d.lessonService.GetMyStreak(ctx, userID, email, sessionID)
		if err != nil {
			return fmt.Errorf("streak request: %w", err)
		}
		data, err := decodeServiceResponse[dto.UserStreakResponse](resp)
		if err != nil {
			return fmt.Errorf("streak decode: %w", err)
		}
		streak = *data
		return nil
	})

	if err := g.Wait(); err != nil {
		utils.Fail(c, "Unable to build dashboard summary", http.StatusBadGateway, err.Error())
		return
	}

	studyMinutes := minutesFromMillis(sessionStats.TotalDurationMs)
	summary := dto.DashboardSummary{
		LessonsCompleted:   lessonStats.Completed,
		LessonsInProgress:  lessonStats.InProgress,
		StudySessions:      sessionStats.TotalSessions,
		StudyTimeMinutes:   studyMinutes,
		StudyTimeFormatted: formatStudyDuration(sessionStats.TotalDurationMs),
		TotalPoints: dto.PointsBreakdown{
			Lifetime: points.Lifetime,
			Weekly:   points.Weekly,
			Monthly:  points.Monthly,
		},
		CurrentStreakDays:  streak.CurrentLen,
		LongestStreakDays:  streak.LongestLen,
		LastStreakActivity: streak.LastDay,
	}

	utils.Success(c, summary)
}

type serviceResponse[T any] struct {
	Status  string `json:"status"`
	Data    T      `json:"data"`
	Message string `json:"message"`
	Error   any    `json:"error"`
}

func decodeServiceResponse[T any](resp *types.HTTPResponse) (*T, error) {
	if resp == nil {
		return nil, fmt.Errorf("empty response")
	}

	if resp.StatusCode >= http.StatusBadRequest {
		body := strings.TrimSpace(string(resp.Body))
		if len(body) > 256 {
			body = body[:256] + "..."
		}
		return nil, fmt.Errorf("remote status %d: %s", resp.StatusCode, body)
	}

	if len(resp.Body) == 0 {
		return nil, fmt.Errorf("empty response body")
	}

	var envelope serviceResponse[T]
	if err := json.Unmarshal(resp.Body, &envelope); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if strings.ToLower(envelope.Status) != "success" {
		return nil, fmt.Errorf("remote status: %s", envelope.Status)
	}

	return &envelope.Data, nil
}

func minutesFromMillis(ms int64) int64 {
	if ms <= 0 {
		return 0
	}
	duration := time.Duration(ms) * time.Millisecond
	return int64(duration / time.Minute)
}

func formatStudyDuration(ms int64) string {
	if ms <= 0 {
		return "0m"
	}
	duration := time.Duration(ms) * time.Millisecond
	hours := duration / time.Hour
	minutes := (duration % time.Hour) / time.Minute

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}
