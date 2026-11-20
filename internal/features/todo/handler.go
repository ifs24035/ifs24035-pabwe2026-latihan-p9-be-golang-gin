package todo

import (
	"net/http"
	"strconv"

	"app/internal/response"
	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	service TodoService
}

func NewTodoHandler(service TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

func getUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return 0, false
	}
	return val.(uint), true
}

func (h *TodoHandler) GetTodos(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var groupID *uint
	groupParam := c.Query("todo_group_id")
	if groupParam != "" {
		if id, err := strconv.ParseUint(groupParam, 10, 32); err == nil {
			gid := uint(id)
			groupID = &gid
		}
	}

	todos, err := h.service.GetTodos(userID, groupID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todos retrieved successfully", todos)
}

func (h *TodoHandler) GetTodoByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	todo, err := h.service.GetTodoByID(userID, uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo retrieved successfully", todo)
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	todo, err := h.service.CreateTodo(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Todo created successfully", todo)
}

func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	todo, err := h.service.UpdateTodo(userID, uint(id), req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo updated successfully", todo)
}

func (h *TodoHandler) UpdateTodoStatus(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	var req UpdateTodoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	todo, err := h.service.UpdateTodoStatus(userID, uint(id), req.Status)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo status updated successfully", todo)
}

func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	if err := h.service.DeleteTodo(userID, uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo deleted successfully")
}

func (h *TodoHandler) GetStats(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	stats, err := h.service.GetStats(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo stats retrieved successfully", stats)
}

// Groups Handlers
func (h *TodoHandler) GetGroups(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	groups, err := h.service.GetGroups(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo groups retrieved successfully", groups)
}

func (h *TodoHandler) GetGroupByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	group, err := h.service.GetGroupByID(userID, uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo group retrieved successfully", group)
}

func (h *TodoHandler) CreateGroup(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req CreateTodoGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	group, err := h.service.CreateGroup(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Todo group created successfully", group)
}

func (h *TodoHandler) UpdateGroup(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	var req UpdateTodoGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	group, err := h.service.UpdateGroup(userID, uint(id), req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo group updated successfully", group)
}

func (h *TodoHandler) DeleteGroup(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	if err := h.service.DeleteGroup(userID, uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Todo group deleted successfully")
}

// Public Templates Handlers
func (h *TodoHandler) GetPublicTemplates(c *gin.Context) {
	templates, err := h.service.GetPublicTemplates()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Public templates retrieved successfully", templates)
}

func (h *TodoHandler) GetPublicTemplateByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	tpl, err := h.service.GetPublicTemplateByID(uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Public template retrieved successfully", tpl)
}

func (h *TodoHandler) CloneTemplate(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	group, err := h.service.CloneTemplate(userID, uint(id))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Template cloned successfully", group)
}
