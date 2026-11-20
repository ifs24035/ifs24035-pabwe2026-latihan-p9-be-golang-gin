package ai

type GenerateTodoRequest struct {
	Prompt           string `json:"prompt" binding:"required"`
	Count            int    `json:"count"`
	IncludeWebSearch bool   `json:"include_web_search"`
}

type GeneratedTodoItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	SourceURL   string `json:"source_url,omitempty"`
}

type GenerateTodoResponse struct {
	Items         []GeneratedTodoItem `json:"items"`
	SearchSources []string            `json:"search_sources,omitempty"`
}

type SearchResultItem struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

type SearchEngineResponse struct {
	Results []SearchResultItem `json:"results"`
}
