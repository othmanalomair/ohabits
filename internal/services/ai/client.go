package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Service handles AI operations using Ollama
type Service struct {
	baseURL   string
	model     string
	client    *http.Client
	semaphore chan struct{} // Allows only one request at a time
	mu        sync.Mutex
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

// New creates a new AI service
func New(baseURL, model string) *Service {
	return &Service{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		semaphore: make(chan struct{}, 1), // Only allow 1 concurrent request
	}
}

// Generate sends a prompt to Ollama and returns the response
func (s *Service) Generate(ctx context.Context, prompt string) (string, error) {
	// Acquire semaphore (wait if another request is in progress)
	select {
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	reqBody := OllamaRequest{
		Model:  s.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("فشل تحويل الطلب: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("فشل إنشاء الطلب: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("فشل الاتصال بـ Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("خطأ من Ollama (كود %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("فشل قراءة الرد: %w", err)
	}

	return ollamaResp.Response, nil
}

// FixText improves or corrects the given text
func (s *Service) FixText(ctx context.Context, text string, action string) (string, error) {
	var prompt string

	switch action {
	case "improve":
		prompt = PromptImproveText(text)
	case "fix":
		prompt = PromptFixErrors(text)
	case "simplify":
		prompt = PromptSimplifyText(text)
	default:
		prompt = PromptImproveText(text)
	}

	return s.Generate(ctx, prompt)
}

// GenerateTitles generates suggested titles for the blog content
func (s *Service) GenerateTitles(ctx context.Context, content string) ([]string, error) {
	prompt := PromptGenerateTitles(content)

	response, err := s.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse the response to extract titles
	titles := parseTitles(response)
	return titles, nil
}

// parseTitles extracts titles from the AI response
func parseTitles(response string) []string {
	var titles []string
	lines := splitLines(response)

	for _, line := range lines {
		// Skip empty lines
		if line == "" {
			continue
		}

		// Remove common prefixes like "1.", "2.", "3.", "-", "*"
		cleaned := cleanTitleLine(line)
		if cleaned != "" {
			titles = append(titles, cleaned)
		}

		// Limit to 5 titles
		if len(titles) >= 5 {
			break
		}
	}

	return titles
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func cleanTitleLine(line string) string {
	runes := []rune(line)
	if len(runes) == 0 {
		return ""
	}

	// Skip leading whitespace
	i := 0
	for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
		i++
	}

	if i >= len(runes) {
		return ""
	}

	// Skip number prefix (1., 2., etc.)
	if runes[i] >= '0' && runes[i] <= '9' {
		for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
			i++
		}
		if i < len(runes) && (runes[i] == '.' || runes[i] == ')' || runes[i] == '-') {
			i++
		}
	}

	// Skip bullet points
	if i < len(runes) && (runes[i] == '-' || runes[i] == '*' || runes[i] == '•') {
		i++
	}

	// Skip any remaining whitespace
	for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
		i++
	}

	// Remove quotes if present
	result := string(runes[i:])
	result = trimQuotes(result)

	return result
}

func trimQuotes(s string) string {
	runes := []rune(s)
	if len(runes) >= 2 {
		// Check for quotes at start and end
		if (runes[0] == '"' && runes[len(runes)-1] == '"') ||
			(runes[0] == '\'' && runes[len(runes)-1] == '\'') ||
			(runes[0] == '«' && runes[len(runes)-1] == '»') {
			return string(runes[1 : len(runes)-1])
		}
	}
	return s
}

// IsAvailable checks if Ollama service is available
func (s *Service) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
