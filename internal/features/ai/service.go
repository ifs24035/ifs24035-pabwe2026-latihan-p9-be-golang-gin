package ai

import (
	"errors"
	"fmt"
)

type AiService interface {
	GenerateTodos(req GenerateTodoRequest) (*GenerateTodoResponse, error)
}

type aiService struct {
	client       AiClientInterface
	searchClient SearchEngineClient
}

func NewAiService(client AiClientInterface, searchClient SearchEngineClient) AiService {
	return &aiService{
		client:       client,
		searchClient: searchClient,
	}
}

func (s *aiService) GenerateTodos(req GenerateTodoRequest) (*GenerateTodoResponse, error) {
	if req.Prompt == "" {
		return nil, errors.New("prompt is required")
	}

	// 1. Cek koneksi OpenAI terlebih dahulu
	if err := s.client.CheckAccess(); err != nil {
		return nil, fmt.Errorf("gagal mengakses OpenAI: %w", err)
	}

	count := req.Count
	if count <= 0 {
		count = 3
	}

	// 2. Jika di centang web baru cari di web
	var searchSnippets []string
	if req.IncludeWebSearch && s.searchClient != nil {
		snippets, err := s.searchClient.Search(req.Prompt)
		if err == nil && len(snippets) > 0 {
			searchSnippets = snippets
		}
	}

	// 3. Barulah buat daftar todo
	items, err := s.client.ChatCompletion(req.Prompt, count, searchSnippets)
	if err != nil {
		return nil, err
	}

	res := &GenerateTodoResponse{
		Items: items,
	}
	if req.IncludeWebSearch && len(searchSnippets) > 0 {
		res.SearchSources = searchSnippets
	}

	return res, nil
}
