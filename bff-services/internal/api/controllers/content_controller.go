package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// ContentController proxies content related operations to the content service.
type ContentController struct {
	contentService services.ContentService
}

// NewContentController constructs a new ContentController instance.
func NewContentController(contentService services.ContentService) *ContentController {
	return &ContentController{contentService: contentService}
}

// GetTopicsLevelsTags fetches topics, levels and tags metadata.
func (c *ContentController) GetTopicsLevelsTags(ctx *gin.Context) {
	token := getOptionalBearerToken(ctx)

	resp, err := c.contentService.GetTopicsLevelsTags(ctx.Request.Context(), token)
	if err != nil {
		utils.Fail(ctx, "Unable to fetch content metadata", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(ctx, resp)
}

// GetLessons fetches lessons using the provided query parameters.
func (c *ContentController) GetLessons(ctx *gin.Context) {
	token := getOptionalBearerToken(ctx)

	var params dto.LessonQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		utils.Fail(ctx, "Invalid query parameters", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := c.contentService.GetLessons(ctx.Request.Context(), token, params)
	if err != nil {
		utils.Fail(ctx, "Unable to fetch lessons", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(ctx, resp)
}

// CreateLesson proxies lesson creation to the content service.
func (c *ContentController) CreateLesson(ctx *gin.Context) {
	token, ok := requireBearerToken(ctx)
	if !ok {
		return
	}

	var payload dto.CreateLessonRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		utils.Fail(ctx, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := c.contentService.CreateLesson(ctx.Request.Context(), token, payload)
	if err != nil {
		utils.Fail(ctx, "Unable to create lesson", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(ctx, resp)
}

// GetFlashcardSets fetches flashcard sets from the content service.
func (c *ContentController) GetFlashcardSets(ctx *gin.Context) {
	token := getOptionalBearerToken(ctx)

	var params dto.FlashcardQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		utils.Fail(ctx, "Invalid query parameters", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := c.contentService.GetFlashcardSets(ctx.Request.Context(), token, params)
	if err != nil {
		utils.Fail(ctx, "Unable to fetch flashcard sets", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(ctx, resp)
}

// GetQuizzes fetches quizzes for a specific lesson.
func (c *ContentController) GetQuizzes(ctx *gin.Context) {
	token := getOptionalBearerToken(ctx)

	var params dto.QuizQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		utils.Fail(ctx, "Invalid query parameters", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := c.contentService.GetQuizzes(ctx.Request.Context(), token, params)
	if err != nil {
		utils.Fail(ctx, "Unable to fetch quizzes", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(ctx, resp)
}

// ProxyGraphQL forwards arbitrary GraphQL requests to the content service.
func (c *ContentController) ProxyGraphQL(ctx *gin.Context) {
	token := getOptionalBearerToken(ctx)

	var payload dto.GraphQLRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		utils.Fail(ctx, "Invalid GraphQL request", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := c.contentService.ExecuteGraphQL(ctx.Request.Context(), token, payload)
	if err != nil {
		utils.Fail(ctx, "Unable to execute GraphQL request", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(ctx, resp)
}
