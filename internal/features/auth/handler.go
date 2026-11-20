package auth

import (
	"net/http"

	"app/internal/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.service.Register(req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.service.Login(req)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Login successful", res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, "Logged out successfully")
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID := userIDVal.(uint)
	res, err := h.service.GetProfile(userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile retrieved successfully", res)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := userIDVal.(uint)
	res, err := h.service.UpdateProfile(userID, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Profile updated successfully", res)
}
