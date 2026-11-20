package docs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSwaggerJSON(t *testing.T) {
	// 1. Test with custom host and scheme
	data := GetSwaggerJSON("api.example.com:8443", "https")
	var doc map[string]interface{}
	err := json.Unmarshal(data, &doc)
	assert.NoError(t, err)
	assert.Equal(t, "api.example.com:8443", doc["host"])
	assert.Equal(t, []interface{}{"https"}, doc["schemes"])

	// 2. Test with empty host and scheme
	dataEmpty := GetSwaggerJSON("", "")
	var docEmpty map[string]interface{}
	errEmpty := json.Unmarshal(dataEmpty, &docEmpty)
	assert.NoError(t, errEmpty)
	assert.Equal(t, "", docEmpty["host"])

	// 3. Test invalid ApiJSON fallback
	origApiJSON := ApiJSON
	defer func() { ApiJSON = origApiJSON }()
	ApiJSON = []byte("invalid json")
	fallback := GetSwaggerJSON("api.test", "http")
	assert.Equal(t, []byte("invalid json"), fallback)
}
