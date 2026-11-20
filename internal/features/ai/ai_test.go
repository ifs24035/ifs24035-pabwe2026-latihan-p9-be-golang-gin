package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockAiClient struct {
	items    []GeneratedTodoItem
	err      error
	checkErr error
}

func (m *mockAiClient) CheckAccess() error {
	return m.checkErr
}

func (m *mockAiClient) ChatCompletion(prompt string, count int, searchSnippets []string) ([]GeneratedTodoItem, error) {
	return m.items, m.err
}

type mockSearchEngineClient struct {
	results []string
	err     error
}

func (m *mockSearchEngineClient) Search(query string) ([]string, error) {
	return m.results, m.err
}

func TestAiService_And_Handler(t *testing.T) {
	mockClient := &mockAiClient{
		items: []GeneratedTodoItem{
			{Title: "Belajar Gin", Description: "Rute dan Middleware", Status: "pending", SourceURL: "https://gin-gonic.com"},
			{Title: "Belajar React", Description: "Komponen dan Hooks", Status: "pending"},
		},
	}
	mockSearch := &mockSearchEngineClient{
		results: []string{"Delcom - Belajar Go Gin (https://delcom.org)"},
	}

	service := NewAiService(mockClient, mockSearch)
	handler := NewAiHandler(service)

	// Success Handler without search
	req := GenerateTodoRequest{Prompt: "Belajar Fullstack", Count: 2, IncludeWebSearch: false}
	bReq, _ := json.Marshal(req)
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request, _ = http.NewRequest(http.MethodPost, "/api/todos/ai-generate", bytes.NewBuffer(bReq))
	c1.Request.Header.Set("Content-Type", "application/json")
	handler.GenerateTodos(c1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Success with search enabled
	reqSearch := GenerateTodoRequest{Prompt: "Belajar Fullstack", Count: 2, IncludeWebSearch: true}
	bReqSearch, _ := json.Marshal(reqSearch)
	wSearch := httptest.NewRecorder()
	cSearch, _ := gin.CreateTestContext(wSearch)
	cSearch.Request, _ = http.NewRequest(http.MethodPost, "/api/todos/ai-generate", bytes.NewBuffer(bReqSearch))
	cSearch.Request.Header.Set("Content-Type", "application/json")
	handler.GenerateTodos(cSearch)
	assert.Equal(t, http.StatusOK, wSearch.Code)

	// Search returns error fallback
	mockSearch.err = errors.New("search engine down")
	resSearchErr, err := service.GenerateTodos(reqSearch)
	assert.NoError(t, err)
	assert.Empty(t, resSearchErr.SearchSources)
	mockSearch.err = nil

	// Invalid JSON
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest(http.MethodPost, "/api/todos/ai-generate", bytes.NewBuffer([]byte("{invalid")))
	c2.Request.Header.Set("Content-Type", "application/json")
	handler.GenerateTodos(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// Empty Prompt in service
	_, err = service.GenerateTodos(GenerateTodoRequest{Prompt: ""})
	assert.Error(t, err)

	// Default count in service (count <= 0)
	res, err := service.GenerateTodos(GenerateTodoRequest{Prompt: "Topik Umum", Count: 0})
	assert.NoError(t, err)
	assert.Len(t, res.Items, 2)

	// OpenAI check error before web search
	mockClient.checkErr = errors.New("cannot reach openai")
	_, err = service.GenerateTodos(req)
	assert.Error(t, err)
	mockClient.checkErr = nil

	// Client error in service / handler
	mockClient.err = errors.New("upstream timeout")
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request, _ = http.NewRequest(http.MethodPost, "/api/todos/ai-generate", bytes.NewBuffer(bReq))
	c3.Request.Header.Set("Content-Type", "application/json")
	handler.GenerateTodos(c3)
	assert.Equal(t, http.StatusInternalServerError, w3.Code)
}

func TestOpenAiClient_ChatCompletion(t *testing.T) {
	// 1. Success with raw JSON and searchSnippets
	serverOk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		if r.URL.Path == "/models" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		resp := openAiResponse{
			Choices: []openAiChoice{
				{
					Message: openAiMessage{
						Role:    "assistant",
						Content: `[{"title":"Task 1","description":"Desc 1","status":"pending","source_url":"https://example.com"}]`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverOk.Close()

	client := NewOpenAiClient(serverOk.URL, "test-key", "test-model")
	items, err := client.ChatCompletion("buat todo", 1, []string{"Snippet 1", "Snippet 2"})
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Task 1", items[0].Title)
	assert.Equal(t, "https://example.com", items[0].SourceURL)

	// CheckAccess tests
	assert.NoError(t, client.CheckAccess())
	clientNoKey := NewOpenAiClient(serverOk.URL, "", "test-model")
	assert.Error(t, clientNoKey.CheckAccess())
	clientBadUrl := NewOpenAiClient("http://127.0.0.1:59999", "test-key", "test-model")
	assert.Error(t, clientBadUrl.CheckAccess())

	serverUnauthorized := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer serverUnauthorized.Close()
	clientUnauthorized := NewOpenAiClient(serverUnauthorized.URL, "bad-key", "test-model")
	assert.Error(t, clientUnauthorized.CheckAccess())

	// 2. Success with markdown fences and no searchSnippets
	serverMarkdown := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAiResponse{
			Choices: []openAiChoice{
				{
					Message: openAiMessage{
						Role: "assistant",
						Content: "```json\n" +
							`[{"title":"Task Markdown","description":"Desc","status":"pending"}]` +
							"\n```",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverMarkdown.Close()

	clientMd := NewOpenAiClient(serverMarkdown.URL, "test-key", "test-model")
	itemsMd, err := clientMd.ChatCompletion("buat todo", -1, nil)
	assert.NoError(t, err)
	assert.Len(t, itemsMd, 1)
	assert.Equal(t, "Task Markdown", itemsMd[0].Title)

	// 3. Upstream HTTP Error (status 500)
	serverErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer serverErr.Close()

	clientErr := NewOpenAiClient(serverErr.URL, "test-key", "test-model")
	_, err = clientErr.ChatCompletion("buat todo", 1, nil)
	assert.Error(t, err)

	// 4. Upstream Empty choices
	serverEmpty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAiResponse{Choices: []openAiChoice{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverEmpty.Close()

	clientEmpty := NewOpenAiClient(serverEmpty.URL, "test-key", "test-model")
	_, err = clientEmpty.ChatCompletion("buat todo", 1, nil)
	assert.Error(t, err)

	// 5. Invalid JSON response from AI
	serverBadJson := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAiResponse{
			Choices: []openAiChoice{
				{
					Message: openAiMessage{
						Role:    "assistant",
						Content: `Not a json string at all`,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverBadJson.Close()

	clientBadJson := NewOpenAiClient(serverBadJson.URL, "test-key", "test-model")
	_, err = clientBadJson.ChatCompletion("buat todo", 1, nil)
	assert.Error(t, err)

	// 6. Network failure / invalid URL
	clientNetFail := NewOpenAiClient("http://127.0.0.1:54399", "test-key", "test-model")
	_, err = clientNetFail.ChatCompletion("buat todo", 1, nil)
	assert.Error(t, err)

	// 7. Malformed unmarshal response
	serverGarbage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<!DOCTYPE html><html>error</html>"))
	}))
	defer serverGarbage.Close()

	clientGarbage := NewOpenAiClient(serverGarbage.URL, "test-key", "test-model")
	_, err = clientGarbage.ChatCompletion("buat todo", 1, nil)
	assert.Error(t, err)
}

func TestSearchEngineClient(t *testing.T) {
	// Default client construction
	defClient := NewSearchEngineClient("")
	assert.NotNil(t, defClient)

	// Empty query returns empty
	emptyRes, err := defClient.Search("   ")
	assert.NoError(t, err)
	assert.Empty(t, emptyRes)

	// 1. Wikipedia OpenSearch parsing [term, titles, descriptions, urls]
	serverWikiOS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := `["golang", ["Golang", "Go Lang No Desc", "Only Title"], ["Bahasa pemrograman modern", "", ""], ["https://id.wikipedia.org/wiki/Golang", "https://id.wikipedia.org/wiki/Go", ""]]`
		_, _ = w.Write([]byte(raw))
	}))
	defer serverWikiOS.Close()

	clientWikiOS := NewSearchEngineClient(serverWikiOS.URL+"/wikipedia.org/api.php?custom=1", serverWikiOS.Client())
	resWikiOS, err := clientWikiOS.Search("golang")
	assert.NoError(t, err)
	assert.Len(t, resWikiOS, 3)
	assert.Contains(t, resWikiOS[0], "Bahasa pemrograman modern")
	assert.Contains(t, resWikiOS[1], "https://id.wikipedia.org/wiki/Go")
	assert.Equal(t, "Only Title", resWikiOS[2])

	// 2. Wikipedia Query parsing {"query": {"search": [...]}}
	serverWikiQuery := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := `{"query":{"search":[{"title":"React","snippet":"React is a <span>JS</span> library &amp; &quot;UI&quot; framework"}]}}`
		_, _ = w.Write([]byte(raw))
	}))
	defer serverWikiQuery.Close()

	clientWikiQuery := NewSearchEngineClient(serverWikiQuery.URL+"/wikipedia.org", serverWikiQuery.Client())
	resWikiQuery, err := clientWikiQuery.Search("React")
	assert.NoError(t, err)
	assert.Len(t, resWikiQuery, 1)
	assert.Contains(t, resWikiQuery[0], "React is a JS library & \"UI\" framework")

	// 3. DuckDuckGo parsing
	serverDDG := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := `{"Heading":"Go","AbstractText":"Open source programming language","AbstractURL":"https://duckduckgo.com/Go","RelatedTopics":[{"Text":"Go syntax","FirstURL":"https://duckduckgo.com/syntax"}]}`
		_, _ = w.Write([]byte(raw))
	}))
	defer serverDDG.Close()

	clientDDG := NewSearchEngineClient(serverDDG.URL+"/duckduckgo.com?lang=id", serverDDG.Client())
	resDDG, err := clientDDG.Search("Go")
	assert.NoError(t, err)
	assert.Len(t, resDDG, 2)
	assert.Contains(t, resDDG[0], "Open source programming language")

	// 4. Standard SearXNG / Delcom format with empty snippet and valid snippet
	serverStandard := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SearchEngineResponse{
			Results: []SearchResultItem{
				{Title: "Title 1", Snippet: "Snippet 1", URL: "https://example.com/1"},
				{Title: "Title 2", Snippet: "", URL: "https://example.com/2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverStandard.Close()

	clientStandard := NewSearchEngineClient(serverStandard.URL+"?lang=en", serverStandard.Client())
	resStandard, err := clientStandard.Search("test")
	assert.NoError(t, err)
	assert.Len(t, resStandard, 2)
	assert.Contains(t, resStandard[0], "Snippet 1")
	assert.Contains(t, resStandard[1], "(https://example.com/2)")

	// 5. Server OK but empty results -> fallback information
	serverEmpty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := SearchEngineResponse{Results: []SearchResultItem{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer serverEmpty.Close()

	clientEmpty := NewSearchEngineClient(serverEmpty.URL, serverEmpty.Client())
	resEmpty, err := clientEmpty.Search("unknown query")
	assert.NoError(t, err)
	assert.Len(t, resEmpty, 1)
	assert.Contains(t, resEmpty[0], "Informasi relevan mengenai: unknown query")

	// 6. Server Plain text body fallback
	serverNonJson := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Plain text search snippet &amp; notes"))
	}))
	defer serverNonJson.Close()

	clientNonJson := NewSearchEngineClient(serverNonJson.URL, serverNonJson.Client())
	resNonJson, err := clientNonJson.Search("plain")
	assert.NoError(t, err)
	assert.Equal(t, []string{"Plain text search snippet & notes"}, resNonJson)

	// 7. Server Status != 200
	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer serverError.Close()

	clientError := NewSearchEngineClient(serverError.URL, serverError.Client())
	_, err = clientError.Search("error query")
	assert.Error(t, err)

	// 8. Invalid URL / Request creation error (using invalid URL character)
	clientBadURL := NewSearchEngineClient("http://[fe80::1%en0]:invalid-url", &http.Client{})
	_, err = clientBadURL.Search("bad")
	assert.Error(t, err)

	// 9. Network connection failure
	clientNetFail := NewSearchEngineClient("http://127.0.0.1:54398", &http.Client{})
	_, err = clientNetFail.Search("fail")
	assert.Error(t, err)
}
