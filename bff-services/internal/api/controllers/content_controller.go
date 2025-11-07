package controllers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bff-services/internal/api/dto"
	"bff-services/internal/config"
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

// ProxyGraphQL forwards arbitrary GraphQL requests to the content service.
func (c *ContentController) ProxyGraphQL(ctx *gin.Context) {
	token := getOptionalBearerToken(ctx)
	contentType := ctx.GetHeader("Content-Type")

	// Auto-detect multipart requests (file uploads)
	if strings.HasPrefix(contentType, "multipart/form-data") {
		c.proxyMultipart(ctx, token)
		return
	}

	// Handle regular JSON GraphQL
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

// Private helper method
func (c *ContentController) proxyMultipart(ctx *gin.Context, token string) {
	targetURL := fmt.Sprintf("%s/graphql", config.GetContentServiceURL())

	proxyReq, err := http.NewRequestWithContext(
		ctx.Request.Context(),
		"POST",
		targetURL,
		ctx.Request.Body,
	)
	if err != nil {
		utils.Fail(ctx, "Failed to create proxy request", http.StatusInternalServerError, err.Error())
		return
	}

	// Copy headers
	proxyReq.Header = ctx.Request.Header.Clone()
	if token != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+token)
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		utils.Fail(ctx, "Failed to proxy request", http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		if strings.EqualFold(key, "Connection") || strings.EqualFold(key, "Keep-Alive") ||
			strings.EqualFold(key, "Transfer-Encoding") || strings.EqualFold(key, "Upgrade") {
			continue
		}
		for _, value := range values {
			ctx.Header(key, value)
		}
	}

	// Copy response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Fail(ctx, "Failed to read response", http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
