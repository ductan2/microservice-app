package controllers

import (
	"net/http"
	"strings"

	"bff-services/internal/services"

	"github.com/gin-gonic/gin"
)

func respondWithServiceResponse(c *gin.Context, resp *services.HTTPResponse) {
	if resp == nil {
		c.Status(http.StatusNoContent)
		return
	}

	if resp.Headers != nil {
		// Forward upstream headers except those managed by Gin automatically.
		for key, values := range resp.Headers {
			// Skip hop-by-hop headers
			if strings.EqualFold(key, "Connection") || strings.EqualFold(key, "Keep-Alive") ||
				strings.EqualFold(key, "Transfer-Encoding") || strings.EqualFold(key, "Upgrade") {
				continue
			}
			for _, value := range values {
				c.Header(key, value)
			}
		}
	}

	if resp.IsBodyEmpty() {
		c.Status(resp.StatusCode)
		return
	}

	contentType := "application/json"
	if resp.Headers != nil {
		if ct := resp.Headers.Get("Content-Type"); ct != "" {
			contentType = ct
		}
	}

	c.Data(resp.StatusCode, contentType, resp.Body)
}
