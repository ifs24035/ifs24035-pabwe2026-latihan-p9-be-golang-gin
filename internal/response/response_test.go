package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResponse_SuccessWithData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]any{
		"strData": "hello",
		"intData": 90,
		"objData": map[string]string{"key": "value"},
		"arrData": []string{"item1", "item2"},
	}

	Success(c, http.StatusOK, "Data fetched successfully", data)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "success", body["status"])
	assert.Equal(t, "Data fetched successfully", body["message"])
	assert.NotNil(t, body["data"])
}

func TestResponse_SuccessWithoutData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, http.StatusOK, "Operation completed without data")

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "success", body["status"])
	assert.Equal(t, "Operation completed without data", body["message"])
	_, hasData := body["data"]
	assert.False(t, hasData)
}

func TestResponse_SuccessWithNilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, http.StatusOK, "Operation completed with nil data", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "success", body["status"])
	assert.Equal(t, "Operation completed with nil data", body["message"])
	_, hasData := body["data"]
	assert.False(t, hasData)
}

func TestResponse_Fail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Fail(c, http.StatusBadRequest, "Invalid input data")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "fail", body["status"])
	assert.Equal(t, "Invalid input data", body["message"])
	_, hasData := body["data"]
	assert.False(t, hasData)
}

func TestResponse_Error(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, http.StatusInternalServerError, "Internal database error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "error", body["status"])
	assert.Equal(t, "Internal database error", body["message"])
	_, hasData := body["data"]
	assert.False(t, hasData)
}

func TestResponse_AbortFail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	AbortFail(c, http.StatusUnauthorized, "Unauthorized token")

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "fail", body["status"])
	assert.Equal(t, "Unauthorized token", body["message"])
}

func TestResponse_AbortError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	AbortError(c, http.StatusInternalServerError, "Fatal system error")

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "error", body["status"])
	assert.Equal(t, "Fatal system error", body["message"])
}
