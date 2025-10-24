package controllers

import (
	"net/http"
	"strconv"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// QuizAttemptController exposes handlers for quiz attempt operations.
type QuizAttemptController struct {
	quizAttemptService services.QuizAttemptService
}

// NewQuizAttemptController constructs a new QuizAttemptController.
func NewQuizAttemptController(quizAttemptService services.QuizAttemptService) *QuizAttemptController {
	return &QuizAttemptController{
		quizAttemptService: quizAttemptService,
	}
}

// StartQuizAttempt handles POST /api/v1/quiz-attempts/start.
func (q *QuizAttemptController) StartQuizAttempt(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.QuizAttemptStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := q.quizAttemptService.StartQuizAttempt(c.Request.Context(), userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to start quiz attempt", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// GetQuizAttempt handles GET /api/v1/quiz-attempts/:attempt_id.
func (q *QuizAttemptController) GetQuizAttempt(c *gin.Context) {
	attemptID := c.Param("attempt_id")
	if attemptID == "" {
		utils.Fail(c, "Attempt ID is required", http.StatusBadRequest, "missing attempt_id path parameter")
		return
	}

	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := q.quizAttemptService.GetQuizAttempt(c.Request.Context(), attemptID, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch quiz attempt", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// SubmitQuizAttempt handles POST /api/v1/quiz-attempts/:attempt_id/submit.
func (q *QuizAttemptController) SubmitQuizAttempt(c *gin.Context) {
	attemptID := c.Param("attempt_id")
	if attemptID == "" {
		utils.Fail(c, "Attempt ID is required", http.StatusBadRequest, "missing attempt_id path parameter")
		return
	}

	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.QuizAttemptSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := q.quizAttemptService.SubmitQuizAttempt(c.Request.Context(), attemptID, userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to submit quiz attempt", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// GetUserQuizAttempts handles GET /api/v1/quiz-attempts/user/me/quiz/:quiz_id.
func (q *QuizAttemptController) GetUserQuizAttempts(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	quizID := c.Param("quiz_id")
	if quizID == "" {
		utils.Fail(c, "Quiz ID is required", http.StatusBadRequest, "missing quiz_id path parameter")
		return
	}

	resp, err := q.quizAttemptService.GetUserQuizAttempts(c.Request.Context(), userID, email, sessionID, quizID)
	if err != nil {
		utils.Fail(c, "Unable to fetch user quiz attempts", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// GetUserQuizHistory handles GET /api/v1/quiz-attempts/user/me/history.
func (q *QuizAttemptController) GetUserQuizHistory(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a number")
		return
	}
	if limit < 1 || limit > 200 {
		utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be between 1 and 200")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset must be a number")
		return
	}
	if offset < 0 {
		utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset cannot be negative")
		return
	}

	passedParam := c.Query("passed")
	var passed *bool
	if passedParam != "" {
		value, err := strconv.ParseBool(passedParam)
		if err != nil {
			utils.Fail(c, "Invalid passed parameter", http.StatusBadRequest, "passed must be true or false")
			return
		}
		passed = &value
	}

	resp, err := q.quizAttemptService.GetUserQuizHistory(c.Request.Context(), userID, email, sessionID, passed, limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch quiz history", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// GetQuizAttemptsByUserID handles GET /api/v1/quiz-attempts/user/:user_id.
func (q *QuizAttemptController) GetQuizAttemptsByUserID(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}
	
	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user_id path parameter")
		return
	}

	resp, err := q.quizAttemptService.GetQuizAttemptsByUserID(c.Request.Context(), targetUserID, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch quiz attempts", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// GetLessonQuizAttempts handles GET /api/v1/quiz-attempts/lesson/:lesson_id/user/me.
func (q *QuizAttemptController) GetLessonQuizAttempts(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	lessonID := c.Param("lesson_id")
	if lessonID == "" {
		utils.Fail(c, "Lesson ID is required", http.StatusBadRequest, "missing lesson_id path parameter")
		return
	}

	resp, err := q.quizAttemptService.GetLessonQuizAttempts(c.Request.Context(), lessonID, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch lesson quiz attempts", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// DeleteQuizAttempt handles DELETE /api/v1/quiz-attempts/:attempt_id.
func (q *QuizAttemptController) DeleteQuizAttempt(c *gin.Context) {
	attemptID := c.Param("attempt_id")
	if attemptID == "" {
		utils.Fail(c, "Attempt ID is required", http.StatusBadRequest, "missing attempt_id path parameter")
		return
	}

	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := q.quizAttemptService.DeleteQuizAttempt(c.Request.Context(), attemptID, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to delete quiz attempt", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

