package admin

import (
	"net/http"
	"strconv"

	"app/internal/response"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service AdminService
}

func NewAdminHandler(service AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Users retrieved successfully", users)
}

func (h *AdminHandler) GetUserTodos(c *gin.Context) {
	idParam := c.Param("id")
	targetID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	todos, err := h.service.GetUserTodos(uint(targetID))
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "User todos retrieved successfully", todos)
}
