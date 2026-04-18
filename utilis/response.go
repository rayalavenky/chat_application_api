package utilis

import "github.com/gin-gonic/gin"

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
)

func Success(c *gin.Context, httpStatus int, message string, data any) {
	c.JSON(httpStatus, gin.H{
		"data":    data,
		"message": message,
		"status":  StatusSuccess,
	})
}

func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, gin.H{
		"message": message,
		"status":  StatusFailure,
	})
}
