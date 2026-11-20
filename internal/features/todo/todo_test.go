package todo

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestTodo(t *testing.T) (*TodoHandler, TodoRepository, TodoService, uint) {
	db, err := database.ConnectTest()
	assert.NoError(t, err)

	user := database.User{
		Username: "todo_user",
		Email:    "todouser@example.com",
		Password: "hashedpassword",
		Role:     "member",
	}
	db.Create(&user)

	repo := NewTodoRepository(db)
	service := NewTodoService(repo)
	handler := NewTodoHandler(service)
	return handler, repo, service, user.ID
}

func TestTodo_CRUD_And_Stats(t *testing.T) {
	handler, _, service, userID := setupTestTodo(t)

	// 1. Create Group
	grp, err := service.CreateGroup(userID, CreateTodoGroupRequest{
		Title:       "Proyek Akhir",
		Description: "Tugas akhir semester",
		IsPublic:    false,
	})
	assert.NoError(t, err)

	// 2. Create Todo
	now := time.Now()
	createReq := CreateTodoRequest{
		TodoGroupID: &grp.ID,
		Title:       "Desain UI",
		Description: "Menggunakan Tailwind dan React",
		Status:      "pending",
		DueDate:     &now,
	}
	bCreate, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Set("userID", userID)
	c1.Request, _ = http.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(bCreate))
	c1.Request.Header.Set("Content-Type", "application/json")
	handler.CreateTodo(c1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Create Todo with default status (empty string)
	emptyStatusReq := CreateTodoRequest{
		Title: "Empty status todo",
	}
	bEmpty, _ := json.Marshal(emptyStatusReq)
	wEmpty := httptest.NewRecorder()
	cEmpty, _ := gin.CreateTestContext(wEmpty)
	cEmpty.Set("userID", userID)
	cEmpty.Request, _ = http.NewRequest(http.MethodPost, "/api/todos", bytes.NewBuffer(bEmpty))
	cEmpty.Request.Header.Set("Content-Type", "application/json")
	handler.CreateTodo(cEmpty)
	assert.Equal(t, http.StatusCreated, wEmpty.Code)

	// 3. Get Todos
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set("userID", userID)
	c2.Request, _ = http.NewRequest(http.MethodGet, "/api/todos?todo_group_id=1", nil)
	handler.GetTodos(c2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Get Todos without query param
	w2b := httptest.NewRecorder()
	c2b, _ := gin.CreateTestContext(w2b)
	c2b.Set("userID", userID)
	c2b.Request, _ = http.NewRequest(http.MethodGet, "/api/todos", nil)
	handler.GetTodos(c2b)
	assert.Equal(t, http.StatusOK, w2b.Code)

	// 4. Get Todo By ID
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Set("userID", userID)
	c3.Params = gin.Params{{Key: "id", Value: "1"}}
	c3.Request, _ = http.NewRequest(http.MethodGet, "/api/todos/1", nil)
	handler.GetTodoByID(c3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// 5. Update Todo
	newDate := time.Now().Add(24 * time.Hour)
	upReq := UpdateTodoRequest{
		Title:          "Desain UI Updated",
		Description:    "Deskripsi baru",
		Status:         "in_progress",
		DueDate:        &newDate,
		CoverImagePath: "/images/cover.png",
		TodoGroupID:    &grp.ID,
	}
	bUp, _ := json.Marshal(upReq)
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Set("userID", userID)
	c4.Params = gin.Params{{Key: "id", Value: "1"}}
	c4.Request, _ = http.NewRequest(http.MethodPut, "/api/todos/1", bytes.NewBuffer(bUp))
	c4.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateTodo(c4)
	assert.Equal(t, http.StatusOK, w4.Code)

	// 5b. Update Todo Delete Cover
	delCoverReq := UpdateTodoRequest{
		CoverImagePath: "DELETE",
	}
	bDelCover, _ := json.Marshal(delCoverReq)
	w4b := httptest.NewRecorder()
	c4b, _ := gin.CreateTestContext(w4b)
	c4b.Set("userID", userID)
	c4b.Params = gin.Params{{Key: "id", Value: "1"}}
	c4b.Request, _ = http.NewRequest(http.MethodPut, "/api/todos/1", bytes.NewBuffer(bDelCover))
	c4b.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateTodo(c4b)
	assert.Equal(t, http.StatusOK, w4b.Code)

	// 6. Update Status
	stReq := UpdateTodoStatusRequest{Status: "completed"}
	bSt, _ := json.Marshal(stReq)
	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	c5.Set("userID", userID)
	c5.Params = gin.Params{{Key: "id", Value: "1"}}
	c5.Request, _ = http.NewRequest(http.MethodPatch, "/api/todos/1/status", bytes.NewBuffer(bSt))
	c5.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateTodoStatus(c5)
	assert.Equal(t, http.StatusOK, w5.Code)

	// 7. Get Stats
	w6 := httptest.NewRecorder()
	c6, _ := gin.CreateTestContext(w6)
	c6.Set("userID", userID)
	c6.Request, _ = http.NewRequest(http.MethodGet, "/api/todos/stats", nil)
	handler.GetStats(c6)
	assert.Equal(t, http.StatusOK, w6.Code)

	// 8. Delete Todo
	w7 := httptest.NewRecorder()
	c7, _ := gin.CreateTestContext(w7)
	c7.Set("userID", userID)
	c7.Params = gin.Params{{Key: "id", Value: "2"}}
	c7.Request, _ = http.NewRequest(http.MethodDelete, "/api/todos/2", nil)
	handler.DeleteTodo(c7)
	assert.Equal(t, http.StatusOK, w7.Code)
}

func TestTodoGroup_CRUD_And_TemplateClone(t *testing.T) {
	handler, _, service, userID := setupTestTodo(t)

	// Create Group via Handler
	createGrpReq := CreateTodoGroupRequest{
		Title:       "Template Starter Go",
		Description: "Template publik",
		IsPublic:    true,
	}
	bGrp, _ := json.Marshal(createGrpReq)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Set("userID", userID)
	c1.Request, _ = http.NewRequest(http.MethodPost, "/api/todo-groups", bytes.NewBuffer(bGrp))
	c1.Request.Header.Set("Content-Type", "application/json")
	handler.CreateGroup(c1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Add todo to this template
	_, _ = service.CreateTodo(userID, CreateTodoRequest{
		TodoGroupID: &isPubUint,
		Title:       "Task di template",
		Status:      "pending",
	})

	// Get Groups
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set("userID", userID)
	c2.Request, _ = http.NewRequest(http.MethodGet, "/api/todo-groups", nil)
	handler.GetGroups(c2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Get Group By ID
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Set("userID", userID)
	c3.Params = gin.Params{{Key: "id", Value: "1"}}
	c3.Request, _ = http.NewRequest(http.MethodGet, "/api/todo-groups/1", nil)
	handler.GetGroupByID(c3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// Update Group
	notPub := false
	upGrpReq := UpdateTodoGroupRequest{
		Title:       "Template Diperbarui",
		Description: "Deskripsi diperbarui",
		IsPublic:    &notPub,
	}
	bUp, _ := json.Marshal(upGrpReq)
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Set("userID", userID)
	c4.Params = gin.Params{{Key: "id", Value: "1"}}
	c4.Request, _ = http.NewRequest(http.MethodPut, "/api/todo-groups/1", bytes.NewBuffer(bUp))
	c4.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateGroup(c4)
	assert.Equal(t, http.StatusOK, w4.Code)

	// Set is_public back to true for template tests
	isPubTrue := true
	_, _ = service.UpdateGroup(userID, 1, UpdateTodoGroupRequest{IsPublic: &isPubTrue})

	// Public Templates
	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	c5.Request, _ = http.NewRequest(http.MethodGet, "/api/templates", nil)
	handler.GetPublicTemplates(c5)
	assert.Equal(t, http.StatusOK, w5.Code)

	// Public Template By ID
	w6 := httptest.NewRecorder()
	c6, _ := gin.CreateTestContext(w6)
	c6.Params = gin.Params{{Key: "id", Value: "1"}}
	c6.Request, _ = http.NewRequest(http.MethodGet, "/api/templates/1", nil)
	handler.GetPublicTemplateByID(c6)
	assert.Equal(t, http.StatusOK, w6.Code)

	// Clone Template
	w7 := httptest.NewRecorder()
	c7, _ := gin.CreateTestContext(w7)
	c7.Set("userID", userID)
	c7.Params = gin.Params{{Key: "id", Value: "1"}}
	c7.Request, _ = http.NewRequest(http.MethodPost, "/api/templates/1/clone", nil)
	handler.CloneTemplate(c7)
	assert.Equal(t, http.StatusCreated, w7.Code)

	// Delete Group
	w8 := httptest.NewRecorder()
	c8, _ := gin.CreateTestContext(w8)
	c8.Set("userID", userID)
	c8.Params = gin.Params{{Key: "id", Value: "2"}}
	c8.Request, _ = http.NewRequest(http.MethodDelete, "/api/todo-groups/2", nil)
	handler.DeleteGroup(c8)
	assert.Equal(t, http.StatusOK, w8.Code)
}

var isPubUint = uint(1)

func TestTodo_ErrorBranches(t *testing.T) {
	handler, _, service, userID := setupTestTodo(t)

	// Unauthenticated calls (getUserID returns false)
	testUnauth := func(fn func(c *gin.Context)) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		fn(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}

	testUnauth(handler.GetTodos)
	testUnauth(handler.GetTodoByID)
	testUnauth(handler.CreateTodo)
	testUnauth(handler.UpdateTodo)
	testUnauth(handler.UpdateTodoStatus)
	testUnauth(handler.DeleteTodo)
	testUnauth(handler.GetStats)
	testUnauth(handler.GetGroups)
	testUnauth(handler.GetGroupByID)
	testUnauth(handler.CreateGroup)
	testUnauth(handler.UpdateGroup)
	testUnauth(handler.DeleteGroup)
	testUnauth(handler.CloneTemplate)

	// Invalid ID parameters
	testInvalidID := func(fn func(c *gin.Context)) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", userID)
		c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}
		fn(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	testInvalidID(handler.GetTodoByID)
	testInvalidID(handler.UpdateTodo)
	testInvalidID(handler.UpdateTodoStatus)
	testInvalidID(handler.DeleteTodo)
	testInvalidID(handler.GetGroupByID)
	testInvalidID(handler.UpdateGroup)
	testInvalidID(handler.DeleteGroup)
	testInvalidID(handler.GetPublicTemplateByID)
	testInvalidID(handler.CloneTemplate)

	// Bad JSON requests
	testBadJSON := func(method, path string, fn func(c *gin.Context)) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", userID)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Request, _ = http.NewRequest(method, path, bytes.NewBuffer([]byte("{bad-json")))
		c.Request.Header.Set("Content-Type", "application/json")
		fn(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	testBadJSON(http.MethodPost, "/api/todos", handler.CreateTodo)
	testBadJSON(http.MethodPut, "/api/todos/1", handler.UpdateTodo)
	testBadJSON(http.MethodPatch, "/api/todos/1/status", handler.UpdateTodoStatus)
	testBadJSON(http.MethodPost, "/api/todo-groups", handler.CreateGroup)
	testBadJSON(http.MethodPut, "/api/todo-groups/1", handler.UpdateGroup)

	// Non-existent resources
	wNotFound := httptest.NewRecorder()
	cNotFound, _ := gin.CreateTestContext(wNotFound)
	cNotFound.Set("userID", userID)
	cNotFound.Params = gin.Params{{Key: "id", Value: "99999"}}
	cNotFound.Request, _ = http.NewRequest(http.MethodGet, "/api/todos/99999", nil)
	handler.GetTodoByID(cNotFound)
	assert.Equal(t, http.StatusNotFound, wNotFound.Code)

	wGrpNotFound := httptest.NewRecorder()
	cGrpNotFound, _ := gin.CreateTestContext(wGrpNotFound)
	cGrpNotFound.Set("userID", userID)
	cGrpNotFound.Params = gin.Params{{Key: "id", Value: "99999"}}
	cGrpNotFound.Request, _ = http.NewRequest(http.MethodGet, "/api/todo-groups/99999", nil)
	handler.GetGroupByID(cGrpNotFound)
	assert.Equal(t, http.StatusNotFound, wGrpNotFound.Code)

	wTplNotFound := httptest.NewRecorder()
	cTplNotFound, _ := gin.CreateTestContext(wTplNotFound)
	cTplNotFound.Params = gin.Params{{Key: "id", Value: "99999"}}
	cTplNotFound.Request, _ = http.NewRequest(http.MethodGet, "/api/templates/99999", nil)
	handler.GetPublicTemplateByID(cTplNotFound)
	assert.Equal(t, http.StatusNotFound, wTplNotFound.Code)

	wCloneNotFound := httptest.NewRecorder()
	cCloneNotFound, _ := gin.CreateTestContext(wCloneNotFound)
	cCloneNotFound.Set("userID", userID)
	cCloneNotFound.Params = gin.Params{{Key: "id", Value: "99999"}}
	cCloneNotFound.Request, _ = http.NewRequest(http.MethodPost, "/api/templates/99999/clone", nil)
	handler.CloneTemplate(cCloneNotFound)
	assert.Equal(t, http.StatusBadRequest, wCloneNotFound.Code)

	// Ownership forbidden test
	otherUserID := uint(9999)
	_, err := service.GetTodoByID(otherUserID, 1)
	assert.Error(t, err)

	_, err = service.UpdateTodo(otherUserID, 1, UpdateTodoRequest{Title: "Hack"})
	assert.Error(t, err)

	_, err = service.UpdateTodoStatus(otherUserID, 1, "completed")
	assert.Error(t, err)

	err = service.DeleteTodo(otherUserID, 1)
	assert.Error(t, err)

	_, err = service.GetGroupByID(otherUserID, 1) // private group
	// If private, other user forbidden
	_, _ = service.UpdateGroup(otherUserID, 1, UpdateTodoGroupRequest{Title: "Hack"})
	_ = service.DeleteGroup(otherUserID, 1)
	_ = service.DeleteGroup(userID, 99999)
}


type mockErrorTodoService struct {
	TodoService
}

func (m *mockErrorTodoService) GetTodos(userID uint, groupID *uint) ([]database.Todo, error) {
	return nil, errors.New("db error")
}
func (m *mockErrorTodoService) CreateTodo(userID uint, req CreateTodoRequest) (*database.Todo, error) {
	return nil, errors.New("db error")
}
func (m *mockErrorTodoService) GetStats(userID uint) (*TodoStatsResponse, error) {
	return nil, errors.New("db error")
}
func (m *mockErrorTodoService) GetGroups(userID uint) ([]database.TodoGroup, error) {
	return nil, errors.New("db error")
}
func (m *mockErrorTodoService) CreateGroup(userID uint, req CreateTodoGroupRequest) (*database.TodoGroup, error) {
	return nil, errors.New("db error")
}
func (m *mockErrorTodoService) GetPublicTemplates() ([]database.TodoGroup, error) {
	return nil, errors.New("db error")
}

func TestTodoHandler_ServiceErrors(t *testing.T) {
	handler := NewTodoHandler(&mockErrorTodoService{})

	testErr := func(method, path string, fn func(c *gin.Context)) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Request, _ = http.NewRequest(method, path, bytes.NewBuffer([]byte(`{"title":"test"}`)))
		c.Request.Header.Set("Content-Type", "application/json")
		fn(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}

	testErr(http.MethodGet, "/api/todos", handler.GetTodos)
	testErr(http.MethodPost, "/api/todos", handler.CreateTodo)
	testErr(http.MethodGet, "/api/todos/stats", handler.GetStats)
	testErr(http.MethodGet, "/api/todo-groups", handler.GetGroups)
	testErr(http.MethodPost, "/api/todo-groups", handler.CreateGroup)
	testErr(http.MethodGet, "/api/templates", handler.GetPublicTemplates)
}

func TestTodoHandler_ForbiddenAndNotFoundBranches(t *testing.T) {
	handler, _, _, _ := setupTestTodo(t)

	// User ID 99999 trying to update or delete non-existent / unowned
	testCaller := func(method, path string, fn func(c *gin.Context), body []byte) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(99999))
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Request, _ = http.NewRequest(method, path, bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		fn(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	testCaller(http.MethodPut, "/api/todos/1", handler.UpdateTodo, []byte(`{"title":"updated"}`))
	testCaller(http.MethodPatch, "/api/todos/1/status", handler.UpdateTodoStatus, []byte(`{"status":"completed"}`))
	testCaller(http.MethodDelete, "/api/todos/1", handler.DeleteTodo, nil)
	testCaller(http.MethodPut, "/api/todo-groups/1", handler.UpdateGroup, []byte(`{"title":"updated"}`))
	testCaller(http.MethodDelete, "/api/todo-groups/1", handler.DeleteGroup, nil)
}

func TestTodoService_PrivateGroupForbidden(t *testing.T) {
	_, _, service, userID := setupTestTodo(t)
	// Create private group
	grp, err := service.CreateGroup(userID, CreateTodoGroupRequest{
		Title:    "Private Group",
		IsPublic: false,
	})
	assert.NoError(t, err)

	// User 99999 attempts to get private group
	_, err = service.GetGroupByID(uint(99999), grp.ID)
	assert.Error(t, err)
}

func TestTodoService_UpdateAndDeleteGroup_Forbidden(t *testing.T) {
	_, _, service, userID := setupTestTodo(t)
	grp, err := service.CreateGroup(userID, CreateTodoGroupRequest{
		Title: "Group for Forbidden",
	})
	assert.NoError(t, err)

	// User 999 attempts update
	_, err = service.UpdateGroup(uint(999), grp.ID, UpdateTodoGroupRequest{Title: "Hacked"})
	assert.Error(t, err)

	// User 999 attempts delete
	err = service.DeleteGroup(uint(999), grp.ID)
	assert.Error(t, err)
}

func TestTodoService_ForbiddenTodo(t *testing.T) {
	_, _, service, userID := setupTestTodo(t)
	td, err := service.CreateTodo(userID, CreateTodoRequest{
		Title: "My Todo",
	})
	assert.NoError(t, err)

	// Another user queries it
	_, err = service.GetTodoByID(uint(88888), td.ID)
	assert.Error(t, err)
	assert.Equal(t, "forbidden: you do not own this todo", err.Error())
}
