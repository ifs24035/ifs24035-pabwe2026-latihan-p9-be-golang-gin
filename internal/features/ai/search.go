package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type SearchEngineClient interface {
	Search(query string) ([]string, error)
}

type searchEngineClient struct {
	endpoint   string
	httpClient *http.Client
}

func NewSearchEngineClient(endpoint string, client ...*http.Client) SearchEngineClient {
	hc := &http.Client{Timeout: 10 * time.Second}
	if len(client) > 0 && client[0] != nil {
		hc = client[0]
	}
	if endpoint == "" {
		endpoint = "https://id.wikipedia.org/w/api.php"
	}
	return &searchEngineClient{
		endpoint:   endpoint,
		httpClient: hc,
	}
}

func stripHTML(input string) string {
	re := regexp.MustCompile("<[^>]*>")
	cleaned := re.ReplaceAllString(input, "")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&#039;", "'")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	return strings.TrimSpace(cleaned)
}

func (s *searchEngineClient) buildSearchURL(query string) string {
	escapedQuery := url.QueryEscape(query)
	// Open-source Wikipedia API
	if strings.Contains(s.endpoint, "wikipedia.org") {
		sep := "?"
		if strings.Contains(s.endpoint, "?") {
			sep = "&"
		}
		return fmt.Sprintf("%s%saction=opensearch&search=%s&limit=5&namespace=0&format=json", s.endpoint, sep, escapedQuery)
	}
	// DuckDuckGo Open Search / Instant Answer API
	if strings.Contains(s.endpoint, "duckduckgo.com") {
		sep := "?"
		if strings.Contains(s.endpoint, "?") {
			sep = "&"
		}
		return fmt.Sprintf("%s%sq=%s&format=json", s.endpoint, sep, escapedQuery)
	}
	// SearXNG / Custom Endpoint
	sep := "?"
	if strings.Contains(s.endpoint, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%sq=%s", s.endpoint, sep, escapedQuery)
}

func (s *searchEngineClient) parseResults(body []byte, query string) []string {
	var results []string

	// 1. Try parsing Wikipedia OpenSearch format: [term, [titles], [descriptions], [urls]]
	var openSearch []interface{}
	if err := json.Unmarshal(body, &openSearch); err == nil && len(openSearch) >= 4 {
		titles, ok1 := openSearch[1].([]interface{})
		urls, ok2 := openSearch[3].([]interface{})
		var descs []interface{}
		if d, ok := openSearch[2].([]interface{}); ok {
			descs = d
		}
		if ok1 && ok2 {
			for i := 0; i < len(titles); i++ {
				titleStr, _ := titles[i].(string)
				urlStr := ""
				if i < len(urls) {
					urlStr, _ = urls[i].(string)
				}
				descStr := ""
				if i < len(descs) {
					descStr, _ = descs[i].(string)
				}
				descStr = stripHTML(descStr)
				if titleStr != "" {
					if descStr != "" {
						results = append(results, fmt.Sprintf("%s - %s (%s)", titleStr, descStr, urlStr))
					} else if urlStr != "" {
						results = append(results, fmt.Sprintf("%s (%s)", titleStr, urlStr))
					} else {
						results = append(results, titleStr)
					}
				}
			}
			if len(results) > 0 {
				return results
			}
		}
	}

	// 2. Try parsing Wikipedia query format: {"query": {"search": [{"title": "...", "snippet": "..."}]}}
	type wikiSearchResp struct {
		Query struct {
			Search []struct {
				Title   string `json:"title"`
				Snippet string `json:"snippet"`
			} `json:"search"`
		} `json:"query"`
	}
	var wikiResp wikiSearchResp
	if err := json.Unmarshal(body, &wikiResp); err == nil && len(wikiResp.Query.Search) > 0 {
		for _, item := range wikiResp.Query.Search {
			cleanSnippet := stripHTML(item.Snippet)
			pageURL := fmt.Sprintf("https://id.wikipedia.org/wiki/%s", url.PathEscape(item.Title))
			results = append(results, fmt.Sprintf("%s - %s (%s)", item.Title, cleanSnippet, pageURL))
		}
		return results
	}

	// 3. Try parsing SearXNG / Generic format: {"results": [{"title": "...", "snippet": "...", "url": "..."}]}
	var parsed SearchEngineResponse
	if err := json.Unmarshal(body, &parsed); err == nil && len(parsed.Results) > 0 {
		for _, item := range parsed.Results {
			snippet := strings.TrimSpace(item.Snippet)
			if snippet != "" {
				results = append(results, fmt.Sprintf("%s - %s (%s)", item.Title, snippet, item.URL))
			} else {
				results = append(results, fmt.Sprintf("%s (%s)", item.Title, item.URL))
			}
		}
		return results
	}

	// 4. Try parsing DuckDuckGo Instant Answer format: {"Heading": "...", "AbstractText": "...", "AbstractURL": "...", "RelatedTopics": [...]}
	type ddgResp struct {
		Heading       string `json:"Heading"`
		AbstractText  string `json:"AbstractText"`
		AbstractURL   string `json:"AbstractURL"`
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}
	var ddg ddgResp
	if err := json.Unmarshal(body, &ddg); err == nil {
		if ddg.Heading != "" && ddg.AbstractText != "" {
			results = append(results, fmt.Sprintf("%s - %s (%s)", ddg.Heading, ddg.AbstractText, ddg.AbstractURL))
		}
		for _, rt := range ddg.RelatedTopics {
			if rt.Text != "" && rt.FirstURL != "" {
				results = append(results, fmt.Sprintf("%s (%s)", rt.Text, rt.FirstURL))
			}
		}
		if len(results) > 0 {
			return results
		}
	}

	// 5. Fallback if plain text / simpler response returned
	trimmed := strings.TrimSpace(string(body))
	if trimmed != "" && !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		results = append(results, stripHTML(trimmed))
	}

	return results
}

func (s *searchEngineClient) Search(query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []string{}, nil
	}

	searchURL := s.buildSearchURL(query)
	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "OpenSource-AI-SearchTool/1.0 (Educational Assistant)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search engine request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search engine returned status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	results := s.parseResults(body, query)

	if len(results) == 0 {
		results = append(results, fmt.Sprintf("Informasi relevan mengenai: %s", query))
	}

	return results, nil
}
