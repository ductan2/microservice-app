package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BaseResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, BaseResponse{
		Status: "success",
		Data:   data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, BaseResponse{
		Status: "success",
		Data:   data,
	})
}

func Fail(c *gin.Context, message string, code int, err interface{}) {
	resp := BaseResponse{
		Status:  "error",
		Message: message,
	}
	if err != nil {
		resp.Error = err
	}
	c.JSON(code, resp)
}
