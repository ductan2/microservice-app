package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bff-services/internal/api/dto"
	"bff-services/internal/types"
)

// LessonService defines the contract for interacting with the lesson service API.
type LessonService interface {
	GetUserPoints(ctx context.Context, userID string) (*types.HTTPResponse, error)
	GetUserStreak(ctx context.Context, userID string) (*types.HTTPResponse, error)
	GetMyStreak(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	CheckMyStreak(ctx context.Context, userID, email, sessionID string, payload *dto.StreakCheckRequest) (*types.HTTPResponse, error)
	GetMyStreakStatus(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetStreakLeaderboard(ctx context.Context, userID, email, sessionID string, limit int) (*types.HTTPResponse, error)
	GetDailyActivityToday(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetDailyActivityByDate(ctx context.Context, userID, email, sessionID, activityDate string) (*types.HTTPResponse, error)
	GetDailyActivityRange(ctx context.Context, userID, email, sessionID, dateFrom, dateTo string) (*types.HTTPResponse, error)
	GetDailyActivityWeek(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetDailyActivityMonth(ctx context.Context, userID, email, sessionID, year, month string) (*types.HTTPResponse, error)
	GetDailyActivitySummary(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	IncrementDailyActivity(ctx context.Context, userID, email, sessionID string, payload dto.DailyActivityIncrementRequest) (*types.HTTPResponse, error)
	GetCurrentWeeklyLeaderboard(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error)
	GetCurrentMonthlyLeaderboard(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error)
	GetWeeklyLeaderboardHistory(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error)
	GetMonthlyLeaderboardHistory(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error)
	GetUserLeaderboardHistory(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetWeekLeaderboard(ctx context.Context, weekKey string, limit *int, offset int) (*types.HTTPResponse, error)
	GetMonthLeaderboard(ctx context.Context, monthKey string, limit *int, offset int) (*types.HTTPResponse, error)

	// Course enrollments
	ListMyEnrollments(ctx context.Context, userID, email, sessionID string, status string, limit, offset int) (*types.HTTPResponse, error)
	EnrollCourse(ctx context.Context, userID, email, sessionID string, payload dto.CourseEnrollmentCreate) (*types.HTTPResponse, error)
	GetEnrollment(ctx context.Context, enrollmentID, userID, email, sessionID string) (*types.HTTPResponse, error)
	UpdateEnrollment(ctx context.Context, enrollmentID, userID, email, sessionID string, payload dto.CourseEnrollmentUpdate) (*types.HTTPResponse, error)
	CancelEnrollment(ctx context.Context, enrollmentID, userID, email, sessionID string) (*types.HTTPResponse, error)

	// Course lessons
	ListCourseLessonsByCourseID(ctx context.Context, courseID string) (*types.HTTPResponse, error)
	CreateCourseLesson(ctx context.Context, payload dto.CourseLessonCreate) (*types.HTTPResponse, error)
	UpdateCourseLesson(ctx context.Context, rowID string, payload dto.CourseLessonUpdate) (*types.HTTPResponse, error)
	DeleteCourseLesson(ctx context.Context, rowID string) (*types.HTTPResponse, error)
}

// LessonServiceClient implements LessonService against a remote HTTP REST endpoint.
type LessonServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLessonServiceClient constructs a new LessonServiceClient.
func NewLessonServiceClient(baseURL string, httpClient *http.Client) *LessonServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &LessonServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *LessonServiceClient) GetUserPoints(ctx context.Context, userID string) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	path := "/api/v1/progress/points/user/" + userID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetUserStreak(ctx context.Context, userID string) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	path := "/api/v1/progress/streaks/user/" + userID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetMyStreak(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/streaks/user/me", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) CheckMyStreak(ctx context.Context, userID, email, sessionID string, payload *dto.StreakCheckRequest) (*types.HTTPResponse, error) {
	var body interface{}
	if payload != nil {
		body = payload
	}
	return c.doRequest(ctx, http.MethodPost, "/api/v1/progress/streaks/user/me/check", body, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetMyStreakStatus(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/streaks/user/me/status", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetStreakLeaderboard(ctx context.Context, userID, email, sessionID string, limit int) (*types.HTTPResponse, error) {
	path := "/api/v1/progress/streaks/leaderboard"
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetDailyActivityToday(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/daily-activity/user/me/today", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetDailyActivityByDate(ctx context.Context, userID, email, sessionID, activityDate string) (*types.HTTPResponse, error) {
	if activityDate == "" {
		return nil, fmt.Errorf("activity date is required")
	}
	path := "/api/v1/progress/daily-activity/user/me/date/" + activityDate
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetDailyActivityRange(ctx context.Context, userID, email, sessionID, dateFrom, dateTo string) (*types.HTTPResponse, error) {
	path := "/api/v1/progress/daily-activity/user/me/range"
	query := url.Values{}
	if dateFrom != "" {
		query.Add("date_from", dateFrom)
	}
	if dateTo != "" {
		query.Add("date_to", dateTo)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetDailyActivityWeek(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/daily-activity/user/me/week", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetDailyActivityMonth(ctx context.Context, userID, email, sessionID, year, month string) (*types.HTTPResponse, error) {
	path := "/api/v1/progress/daily-activity/user/me/month"
	query := url.Values{}
	if year != "" {
		query.Add("year", year)
	}
	if month != "" {
		query.Add("month", month)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetDailyActivitySummary(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/daily-activity/user/me/stats/summary", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) IncrementDailyActivity(ctx context.Context, userID, email, sessionID string, payload dto.DailyActivityIncrementRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/progress/daily-activity/increment", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetCurrentWeeklyLeaderboard(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error) {
	path := "/api/v1/leaderboards/weekly/current"
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetCurrentMonthlyLeaderboard(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error) {
	path := "/api/v1/leaderboards/monthly/current"
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetWeeklyLeaderboardHistory(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error) {
	path := "/api/v1/leaderboards/weekly/history"
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetMonthlyLeaderboardHistory(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error) {
	path := "/api/v1/leaderboards/monthly/history"
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetUserLeaderboardHistory(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/leaderboards/user/me/history", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetWeekLeaderboard(ctx context.Context, weekKey string, limit *int, offset int) (*types.HTTPResponse, error) {
	if weekKey == "" {
		return nil, fmt.Errorf("week key is required")
	}
	path := "/api/v1/leaderboards/week/" + weekKey
	query := url.Values{}
	if limit != nil && *limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetMonthLeaderboard(ctx context.Context, monthKey string, limit *int, offset int) (*types.HTTPResponse, error) {
	if monthKey == "" {
		return nil, fmt.Errorf("month key is required")
	}
	path := "/api/v1/leaderboards/month/" + monthKey
	query := url.Values{}
	if limit != nil && *limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

// Course enrollments
func (c *LessonServiceClient) ListMyEnrollments(ctx context.Context, userID, email, sessionID string, status string, limit, offset int) (*types.HTTPResponse, error) {
	path := "/api/course-enrollments/me"
	q := url.Values{}
	if status != "" {
		q.Add("status", status)
	}
	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		q.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) EnrollCourse(ctx context.Context, userID, email, sessionID string, payload dto.CourseEnrollmentCreate) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/course-enrollments", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) GetEnrollment(ctx context.Context, enrollmentID, userID, email, sessionID string) (*types.HTTPResponse, error) {
	path := "/api/course-enrollments/" + enrollmentID
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) UpdateEnrollment(ctx context.Context, enrollmentID, userID, email, sessionID string, payload dto.CourseEnrollmentUpdate) (*types.HTTPResponse, error) {
	path := "/api/course-enrollments/" + enrollmentID
	return c.doRequest(ctx, http.MethodPut, path, payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *LessonServiceClient) CancelEnrollment(ctx context.Context, enrollmentID, userID, email, sessionID string) (*types.HTTPResponse, error) {
	path := "/api/course-enrollments/" + enrollmentID + "/cancel"
	return c.doRequest(ctx, http.MethodPost, path, nil, internalAuthHeaders(userID, email, sessionID))
}

// Course lessons (no auth required on service side)
func (c *LessonServiceClient) ListCourseLessonsByCourseID(ctx context.Context, courseID string) (*types.HTTPResponse, error) {
	path := "/api/course-lessons/by-course/" + courseID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) CreateCourseLesson(ctx context.Context, payload dto.CourseLessonCreate) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/course-lessons", payload, nil)
}

func (c *LessonServiceClient) UpdateCourseLesson(ctx context.Context, rowID string, payload dto.CourseLessonUpdate) (*types.HTTPResponse, error) {
	path := "/api/course-lessons/" + rowID
	return c.doRequest(ctx, http.MethodPut, path, payload, nil)
}

func (c *LessonServiceClient) DeleteCourseLesson(ctx context.Context, rowID string) (*types.HTTPResponse, error) {
	path := "/api/course-lessons/" + rowID
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func (c *LessonServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("lesson service base URL is not configured")
	}

	endpoint := c.baseURL + path

	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &types.HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
	}, nil
}
