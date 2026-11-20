package ai

import (
	"net/http"

	"app/internal/response"
	"github.com/gin-gonic/gin"
)

type AiHandler struct {
	service AiService
}

func NewAiHandler(service AiService) *AiHandler {
	return &AiHandler{service: service}
}

func (h *AiHandler) GenerateTodos(c *gin.Context) {
	var req GenerateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.service.GenerateTodos(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todos generated successfully", res)
}
