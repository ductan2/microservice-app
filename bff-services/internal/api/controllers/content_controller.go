package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bff-services/internal/api/dto"
	"bff-services/internal/config"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxImageUploadSize = 50 << 20 // 50 MiB

// ContentController proxies content operations to the content service.
type ContentController struct {
	contentService services.ContentService
}

// NewContentController constructs a new ContentController.
func NewContentController(contentService services.ContentService) *ContentController {
	return &ContentController{contentService: contentService}
}

func (c *ContentController) ProxyGraphQL(ctx *gin.Context) {
	token := getOptionalBearerToken(ctx)
	contentType := ctx.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		c.proxyMultipart(ctx, token)
		return
	}

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

func (c *ContentController) UploadImages(ctx *gin.Context) {
	if c.contentService == nil {
		utils.Fail(ctx, "Content service unavailable", http.StatusServiceUnavailable, "content service not configured")
		return
	}

	userID, _, _, ok := middleware.GetUserContextFromMiddleware(ctx)
	if !ok {
		return
	}

	if err := ctx.Request.ParseMultipartForm(maxImageUploadSize); err != nil {
		utils.Fail(ctx, "Invalid multipart payload", http.StatusBadRequest, err.Error())
		return
	}
	form := ctx.Request.MultipartForm
	if form == nil {
		utils.Fail(ctx, "Invalid multipart payload", http.StatusBadRequest, "multipart form missing")
		return
	}
	defer form.RemoveAll()

	files := form.File["images"]
	if len(files) == 0 {
		files = form.File["files"]
	}
	if len(files) == 0 {
		utils.Fail(ctx, "No images provided", http.StatusBadRequest, "at least one image is required")
		return
	}

	folderID := strings.TrimSpace(ctx.PostForm("folderId"))
	if folderID == "" {
		folderID = strings.TrimSpace(ctx.PostForm("folder_id"))
	}
	if folderID != "" {
		if _, err := uuid.Parse(folderID); err != nil {
			utils.Fail(ctx, "Invalid folderId", http.StatusBadRequest, "folderId must be a valid UUID")
			return
		}
	}

	token := getOptionalBearerToken(ctx)
	if token == "" {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "missing bearer token")
		return
	}

	resp, err := c.contentService.UploadMediaBatch(ctx.Request.Context(), token, services.MediaBatchUploadOptions{
		Files:      files,
		Kind:       "IMAGE",
		UploadedBy: userID,
		FolderID:   folderID,
	})
	if err != nil {
		utils.Fail(ctx, "Unable to upload images", http.StatusBadGateway, err.Error())
		return
	}

	if resp.StatusCode >= http.StatusBadRequest {
		utils.Fail(ctx, "Content service error", resp.StatusCode, string(resp.Body))
		return
	}

	var gqlResp struct {
		Data struct {
			UploadMediaBatch []map[string]interface{} `json:"uploadMediaBatch"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(resp.Body, &gqlResp); err != nil {
		utils.Fail(ctx, "Invalid response from content service", http.StatusBadGateway, err.Error())
		return
	}
	if len(gqlResp.Errors) > 0 {
		utils.Fail(ctx, "Content service rejected upload", http.StatusBadGateway, gqlResp.Errors[0].Message)
		return
	}

	items := gqlResp.Data.UploadMediaBatch
	if items == nil {
		items = []map[string]interface{}{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": items,
		"count": len(items),
	})
}

func (c *ContentController) proxyMultipart(ctx *gin.Context, token string) {
	targetURL := fmt.Sprintf("%s/graphql", config.GetContentServiceURL())

	proxyReq, err := http.NewRequestWithContext(
		ctx.Request.Context(),
		"POST",
		targetURL,
		ctx.Request.Body,
	)
	if err != nil {
		utils.Fail(ctx, "Unable to create proxy request", http.StatusInternalServerError, err.Error())
		return
	}

	proxyReq.Header = ctx.Request.Header.Clone()
	if token != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		utils.Fail(ctx, "Unable to proxy request", http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		if strings.EqualFold(key, "Connection") || strings.EqualFold(key, "Keep-Alive") ||
			strings.EqualFold(key, "Transfer-Encoding") || strings.EqualFold(key, "Upgrade") {
			continue
		}
		for _, value := range values {
			ctx.Header(key, value)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Fail(ctx, "Unable to read response", http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
