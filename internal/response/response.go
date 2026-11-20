package response

import (
	"github.com/gin-gonic/gin"
)

// Success sends an HTTP response with status: "success", a message, and optional data.
// If data is omitted or nil, the "data" key is excluded from the JSON output.
func Success(c *gin.Context, httpStatus int, message string, data ...any) {
	res := gin.H{
		"status":  "success",
		"message": message,
	}
	if len(data) > 0 && data[0] != nil {
		res["data"] = data[0]
	}
	c.JSON(httpStatus, res)
}

// Fail sends a 4xx response with status: "fail" and a message.
func Fail(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, gin.H{
		"status":  "fail",
		"message": message,
	})
}

// Error sends a 5xx response with status: "error" and a message.
func Error(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, gin.H{
		"status":  "error",
		"message": message,
	})
}

// AbortFail aborts the context with a 4xx response: status: "fail" and a message.
func AbortFail(c *gin.Context, httpStatus int, message string) {
	c.AbortWithStatusJSON(httpStatus, gin.H{
		"status":  "fail",
		"message": message,
	})
}

// AbortError aborts the context with a 5xx response: status: "error" and a message.
func AbortError(c *gin.Context, httpStatus int, message string) {
	c.AbortWithStatusJSON(httpStatus, gin.H{
		"status":  "error",
		"message": message,
	})
}
