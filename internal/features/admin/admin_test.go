package admin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"app/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockAdminService struct {
	users []UserWithStats
	todos []database.Todo
	err   error
}

func (m *mockAdminService) GetAllUsers() ([]UserWithStats, error) {
	return m.users, m.err
}

func (m *mockAdminService) GetUserTodos(targetUserID uint) ([]database.Todo, error) {
	return m.todos, m.err
}

func TestAdmin_Success(t *testing.T) {
	db, err := database.ConnectTest()
	assert.NoError(t, err)

	u := database.User{
		Username: "user1",
		Email:    "u1@example.com",
		Password: "pass",
		Role:     "member",
	}
	db.Create(&u)

	td := database.Todo{
		UserID: u.ID,
		Title:  "Task 1",
		Status: "completed",
	}
	db.Create(&td)

	service := NewAdminService(db)
	handler := NewAdminHandler(service)

	// GetAllUsers
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	handler.GetAllUsers(c1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// GetUserTodos success
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Params = gin.Params{{Key: "id", Value: "1"}}
	c2.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/users/1/todos", nil)
	handler.GetUserTodos(c2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// GetUserTodos not found
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Params = gin.Params{{Key: "id", Value: "99999"}}
	c3.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/users/99999/todos", nil)
	handler.GetUserTodos(c3)
	assert.Equal(t, http.StatusNotFound, w3.Code)

	// GetUserTodos invalid ID param
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c4.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/users/invalid/todos", nil)
	handler.GetUserTodos(c4)
	assert.Equal(t, http.StatusBadRequest, w4.Code)
}

func TestAdmin_ServiceError(t *testing.T) {
	mockSvc := &mockAdminService{
		err: errors.New("db error"),
	}
	handler := NewAdminHandler(mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/users", nil)
	handler.GetAllUsers(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
