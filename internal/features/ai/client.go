package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AiClientInterface interface {
	CheckAccess() error
	ChatCompletion(prompt string, count int, searchSnippets []string) ([]GeneratedTodoItem, error)
}

type openAiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAiRequest struct {
	Model       string          `json:"model"`
	Messages    []openAiMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
}

type openAiChoice struct {
	Message openAiMessage `json:"message"`
}

type openAiResponse struct {
	Choices []openAiChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type OpenAiClient struct {
	Endpoint   string
	Key        string
	Model      string
	HttpClient *http.Client
}

func NewOpenAiClient(endpoint, key, model string) *OpenAiClient {
	return &OpenAiClient{
		Endpoint: strings.TrimRight(endpoint, "/"),
		Key:      key,
		Model:    model,
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OpenAiClient) CheckAccess() error {
	if strings.TrimSpace(c.Key) == "" {
		return errors.New("OpenAI API key is not configured")
	}
	url := fmt.Sprintf("%s/models", c.Endpoint)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Key))

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(respBytes))
	}
	return nil
}

func (c *OpenAiClient) ChatCompletion(prompt string, count int, searchSnippets []string) ([]GeneratedTodoItem, error) {
	if count <= 0 {
		count = 3
	}

	systemInstruction := fmt.Sprintf(`Kamu adalah asisten perencana tugas produktif.
Hasilkan tepat %d item todo list dalam format JSON murni TANPA markdown block, TANPA kata pembuka/penutup.
Format JSON wajib:
[
  {"title": "Judul Todo", "description": "Deskripsi rinci markdown", "status": "pending", "source_url": "URL sumber jika ada"}
]`, count)

	var userPromptBuilder strings.Builder
	userPromptBuilder.WriteString(prompt)

	if len(searchSnippets) > 0 {
		userPromptBuilder.WriteString("\n\nKonteks Penelusuran Mesin Pencari / Search Engine:\n")
		for i, snippet := range searchSnippets {
			userPromptBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, snippet))
		}
		userPromptBuilder.WriteString("\nGunakan fakta dan rujukan di atas untuk merancang item todo yang relevan dan sertakan source_url tautan rujukannya.")
	}

	reqPayload := openAiRequest{
		Model: c.Model,
		Messages: []openAiMessage{
			{Role: "system", Content: systemInstruction},
			{Role: "user", Content: userPromptBuilder.String()},
		},
		Temperature: 0.7,
	}

	bodyBytes, _ := json.Marshal(reqPayload)
	url := fmt.Sprintf("%s/chat/completions", c.Endpoint)

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Key))

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var aiResp openAiResponse
	if err := json.Unmarshal(respBytes, &aiResp); err != nil {
		return nil, err
	}

	if len(aiResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI API")
	}

	rawContent := strings.TrimSpace(aiResp.Choices[0].Message.Content)
	// Strip markdown code fences if model wrapped in ```json ... ```
	if strings.HasPrefix(rawContent, "```") {
		lines := strings.Split(rawContent, "\n")
		if len(lines) >= 2 {
			rawContent = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	rawContent = strings.TrimSpace(rawContent)

	var items []GeneratedTodoItem
	if err := json.Unmarshal([]byte(rawContent), &items); err != nil {
		return nil, fmt.Errorf("failed to parse AI response into todo list: %w", err)
	}

	return items, nil
}
