package auth

import (
	"bytes"
	"encoding/json"
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

func setupTestAuth(t *testing.T) (*AuthHandler, UserRepository, AuthService) {
	db, err := database.ConnectTest()
	assert.NoError(t, err)

	repo := NewUserRepository(db)
	service := NewAuthService(repo, "test-jwt-secret")
	handler := NewAuthHandler(service)
	return handler, repo, service
}

func TestAuth_RegisterAndLogin_Success(t *testing.T) {
	handler, repo, _ := setupTestAuth(t)

	// Register
	regReq := RegisterRequest{
		Username: "johndoe",
		Email:    "john@example.com",
		Password: "password123",
		Role:     "member",
	}
	body, _ := json.Marshal(regReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handler.Register(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify in DB
	user, err := repo.FindByUsernameOrEmail("johndoe")
	assert.NoError(t, err)
	assert.Equal(t, "johndoe", user.Username)

	// Login
	loginReq := LoginRequest{
		Username: "johndoe",
		Password: "password123",
	}
	bodyLogin, _ := json.Marshal(loginReq)
	wLogin := httptest.NewRecorder()
	cLogin, _ := gin.CreateTestContext(wLogin)
	cLogin.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(bodyLogin))
	cLogin.Request.Header.Set("Content-Type", "application/json")
	handler.Login(cLogin)
	assert.Equal(t, http.StatusOK, wLogin.Code)

	// Logout
	wLogout := httptest.NewRecorder()
	cLogout, _ := gin.CreateTestContext(wLogout)
	cLogout.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	handler.Logout(cLogout)
	assert.Equal(t, http.StatusOK, wLogout.Code)

	// FindAll
	users, err := repo.FindAll()
	assert.NoError(t, err)
	assert.True(t, len(users) >= 1)
}

func TestAuth_Register_Duplicate_And_Invalid(t *testing.T) {
	handler, _, _ := setupTestAuth(t)

	regReq := RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(regReq)

	// First registration
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	c1.Request.Header.Set("Content-Type", "application/json")
	handler.Register(c1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Duplicate username
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	c2.Request.Header.Set("Content-Type", "application/json")
	handler.Register(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// Duplicate email with different username
	dupEmailReq := RegisterRequest{
		Username: "alice2",
		Email:    "alice@example.com",
		Password: "password123",
	}
	bodyDup, _ := json.Marshal(dupEmailReq)
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(bodyDup))
	c3.Request.Header.Set("Content-Type", "application/json")
	handler.Register(c3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)

	// Invalid JSON payload
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer([]byte("{invalid-json")))
	c4.Request.Header.Set("Content-Type", "application/json")
	handler.Register(c4)
	assert.Equal(t, http.StatusBadRequest, w4.Code)
}

func TestAuth_Login_Failures(t *testing.T) {
	handler, _, _ := setupTestAuth(t)

	// User not found
	loginReq := LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	c1.Request.Header.Set("Content-Type", "application/json")
	handler.Login(c1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	// Invalid password
	regReq := RegisterRequest{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "password123",
	}
	bReg, _ := json.Marshal(regReq)
	wReg := httptest.NewRecorder()
	cReg, _ := gin.CreateTestContext(wReg)
	cReg.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(bReg))
	cReg.Request.Header.Set("Content-Type", "application/json")
	handler.Register(cReg)

	wrongPass := LoginRequest{
		Username: "bob",
		Password: "wrongpassword",
	}
	bWrong, _ := json.Marshal(wrongPass)
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(bWrong))
	c2.Request.Header.Set("Content-Type", "application/json")
	handler.Login(c2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// Invalid JSON
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request, _ = http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer([]byte("{invalid")))
	c3.Request.Header.Set("Content-Type", "application/json")
	handler.Login(c3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)
}

func TestAuth_Profile(t *testing.T) {
	handler, _, service := setupTestAuth(t)

	// Register user
	res, err := service.Register(RegisterRequest{
		Username: "charlie",
		Email:    "charlie@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	// GetProfile with userID
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Set("userID", res.User.ID)
	c1.Request, _ = http.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	handler.GetProfile(c1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// GetProfile missing userID
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	handler.GetProfile(c2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// GetProfile not found
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Set("userID", uint(99999))
	c3.Request, _ = http.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	handler.GetProfile(c3)
	assert.Equal(t, http.StatusNotFound, w3.Code)

	// UpdateProfile missing userID
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Request, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", nil)
	handler.UpdateProfile(c4)
	assert.Equal(t, http.StatusUnauthorized, w4.Code)

	// UpdateProfile invalid JSON
	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	c5.Set("userID", res.User.ID)
	c5.Request, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewBuffer([]byte("{bad-json")))
	c5.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateProfile(c5)
	assert.Equal(t, http.StatusBadRequest, w5.Code)

	// UpdateProfile success
	upReq := UpdateProfileRequest{
		Email:    "charlie_updated@example.com",
		Password: "newpassword123",
	}
	bUp, _ := json.Marshal(upReq)
	w6 := httptest.NewRecorder()
	c6, _ := gin.CreateTestContext(w6)
	c6.Set("userID", res.User.ID)
	c6.Request, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewBuffer(bUp))
	c6.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateProfile(c6)
	assert.Equal(t, http.StatusOK, w6.Code)

	// Register admin role
	adminReq := RegisterRequest{
		Username: "superadmin",
		Email:    "superadmin@example.com",
		Password: "password123",
		Role:     "admin",
	}
	resAdmin, err := service.Register(adminReq)
	assert.NoError(t, err)
	assert.Equal(t, "admin", resAdmin.User.Role)

	// Register with empty role (defaults to member)
	emptyRoleReq := RegisterRequest{
		Username: "defaultmember",
		Email:    "defaultmember@example.com",
		Password: "password123",
		Role:     "",
	}
	resMember, err := service.Register(emptyRoleReq)
	assert.NoError(t, err)
	assert.Equal(t, "member", resMember.User.Role)

	// Register with custom invalid role (defaults to member)
	invRoleReq := RegisterRequest{
		Username: "invmember",
		Email:    "invmember@example.com",
		Password: "password123",
		Role:     "unknown_role",
	}
	resInv, err := service.Register(invRoleReq)
	assert.NoError(t, err)
	assert.Equal(t, "member", resInv.User.Role)

	// UpdateProfile with only email
	resOnlyEmail, err := service.UpdateProfile(res.User.ID, UpdateProfileRequest{
		Email: "only_email@example.com",
	})
	assert.NoError(t, err)
	assert.Equal(t, "only_email@example.com", resOnlyEmail.Email)

	// UpdateProfile with only password
	resOnlyPass, err := service.UpdateProfile(res.User.ID, UpdateProfileRequest{
		Password: "newpassword456",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resOnlyPass)

	// UpdateProfile non-existent user
	w7 := httptest.NewRecorder()
	c7, _ := gin.CreateTestContext(w7)
	c7.Set("userID", uint(88888))
	c7.Request, _ = http.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewBuffer(bUp))
	c7.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateProfile(c7)
	assert.Equal(t, http.StatusBadRequest, w7.Code)
}
