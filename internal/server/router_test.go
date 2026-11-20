package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"app/internal/config"
	"app/internal/database"
)

func TestSetupRouter_PureApi(t *testing.T) {
	cfg := &config.Config{
		AppEnv: "production",
		AppKey: "test-secret",
	}

	db, err := database.ConnectTest()
	assert.NoError(t, err)

	router := SetupRouter(cfg, db)

	// 1. Test 404 unmatched route
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/random-unmatched-route", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusNotFound, w1.Code)
	assert.Contains(t, w1.Body.String(), "Endpoint not found")

	// 2. Test 404 unmatched API route
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
	assert.Contains(t, w2.Body.String(), "Endpoint not found")

	// 3. Test Public templates endpoint
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/api/templates", nil)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// 4. Test Swagger UI endpoint
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	req4.RequestURI = "/swagger/index.html"
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
	assert.Contains(t, w4.Body.String(), "swagger-ui")

	// 5. Test Swagger Initializer points to api.json
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodGet, "/swagger/swagger-initializer.js", nil)
	req5.RequestURI = "/swagger/swagger-initializer.js"
	router.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusOK, w5.Code)
	assert.Contains(t, w5.Body.String(), `url: "/swagger/api.json"`)

	// 6. Test Swagger api.json endpoint with dynamic host and scheme
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest(http.MethodGet, "/swagger/api.json", nil)
	req6.RequestURI = "/swagger/api.json"
	req6.Host = "api.myapp.test:8080"
	router.ServeHTTP(w6, req6)
	assert.Equal(t, http.StatusOK, w6.Code)
	assert.Contains(t, w6.Body.String(), `"swagger": "2.0"`)
	assert.Contains(t, w6.Body.String(), `"host": "api.myapp.test:8080"`)
	assert.Contains(t, w6.Body.String(), `"http"`)

	// 7. Test Swagger api.json endpoint behind proxy (X-Forwarded-Host & X-Forwarded-Proto)
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest(http.MethodGet, "/swagger/api.json", nil)
	req7.RequestURI = "/swagger/api.json"
	req7.Header.Set("X-Forwarded-Host", "pabwe.del.ac.id")
	req7.Header.Set("X-Forwarded-Proto", "https")
	router.ServeHTTP(w7, req7)
	assert.Equal(t, http.StatusOK, w7.Code)
	assert.Contains(t, w7.Body.String(), `"host": "pabwe.del.ac.id"`)
	assert.Contains(t, w7.Body.String(), `"https"`)

	// 8. Test Swagger doc.json backward compatibility endpoint
	w8 := httptest.NewRecorder()
	req8, _ := http.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	req8.RequestURI = "/swagger/doc.json"
	router.ServeHTTP(w8, req8)
	assert.Equal(t, http.StatusOK, w8.Code)
	assert.Contains(t, w8.Body.String(), `"swagger": "2.0"`)

	// 9. Test Root api.json endpoint
	w9 := httptest.NewRecorder()
	req9, _ := http.NewRequest(http.MethodGet, "/api.json", nil)
	req9.Host = "localhost:7071"
	router.ServeHTTP(w9, req9)
	assert.Equal(t, http.StatusOK, w9.Code)
	assert.Contains(t, w9.Body.String(), `"host": "localhost:7071"`)

	// 10. Test Swagger redirects (/swagger and /swagger/)
	w10a := httptest.NewRecorder()
	req10a, _ := http.NewRequest(http.MethodGet, "/swagger", nil)
	router.ServeHTTP(w10a, req10a)
	assert.Equal(t, http.StatusMovedPermanently, w10a.Code)

	w10b := httptest.NewRecorder()
	req10b, _ := http.NewRequest(http.MethodGet, "/swagger/", nil)
	router.ServeHTTP(w10b, req10b)
	assert.Equal(t, http.StatusMovedPermanently, w10b.Code)
}
