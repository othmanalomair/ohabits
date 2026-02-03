package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ohabits/internal/config"
)

// AIService handles OpenRouter AI API calls
type AIService struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewAIService creates a new AI service instance
func NewAIService(cfg *config.Config) *AIService {
	return &AIService{
		apiKey:  cfg.OpenRouterAPIKey,
		model:   cfg.OpenRouterModel,
		baseURL: "https://openrouter.ai/api/v1/chat/completions",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// OpenRouter request/response types
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// SuggestTitlesRequest is the request for title suggestions
type SuggestTitlesRequest struct {
	Content string `json:"content"`
}

// SuggestTitlesResponse is the response with title suggestions
type SuggestTitlesResponse struct {
	Titles []string `json:"titles"`
}

// FormatMarkdownRequest is the request for markdown formatting
type FormatMarkdownRequest struct {
	Content string `json:"content"`
}

// FormatMarkdownResponse is the response with formatted markdown
type FormatMarkdownResponse struct {
	FormattedContent string `json:"formattedContent"`
}

// CustomPromptRequest is the request for custom AI prompts
type CustomPromptRequest struct {
	Content string `json:"content"`
	Prompt  string `json:"prompt"`
}

// CustomPromptResponse is the response from custom prompts
type CustomPromptResponse struct {
	Result string `json:"result"`
}

// sendRequest sends a request to OpenRouter
func (s *AIService) sendRequest(systemPrompt, userMessage string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("OpenRouter API key not configured")
	}

	reqBody := chatRequest{
		Model: s.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("HTTP-Referer", "https://ohabits.com")
	req.Header.Set("X-Title", "ohabits Blog Assistant")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// SuggestTitles generates 3 title suggestions for blog content
func (s *AIService) SuggestTitles(content string) (*SuggestTitlesResponse, error) {
	systemPrompt := `أنت مساعد كتابة باللغة العربية. مهمتك هي اقتراح 3 عناوين جذابة لمدونة.

القواعد:
- اقترح 3 عناوين فقط
- كل عنوان يجب أن يكون قصيراً وجذاباً (أقل من 60 حرف)
- استخدم اللغة العربية الفصحى السهلة
- أعد العناوين كل واحد في سطر منفصل
- لا تضف أرقام أو رموز قبل العناوين`

	userMessage := fmt.Sprintf("اقترح 3 عناوين لهذا المحتوى:\n\n%s", truncateForAI(content, 2000))

	response, err := s.sendRequest(systemPrompt, userMessage)
	if err != nil {
		return nil, err
	}

	// Parse response into titles
	lines := strings.Split(strings.TrimSpace(response), "\n")
	var titles []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove any leading numbers, dashes, asterisks
		line = regexp.MustCompile(`^[\d\-\*\.\)]+\s*`).ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line != "" {
			titles = append(titles, line)
		}
	}

	// Ensure we have at least 1 title
	if len(titles) == 0 {
		titles = append(titles, response)
	}

	// Limit to 3 titles
	if len(titles) > 3 {
		titles = titles[:3]
	}

	return &SuggestTitlesResponse{Titles: titles}, nil
}

// FormatMarkdown formats content as proper markdown
func (s *AIService) FormatMarkdown(content string) (*FormatMarkdownResponse, error) {
	// Extract existing images to preserve them
	images := extractImages(content)

	systemPrompt := `أنت مساعد تنسيق Markdown باللغة العربية. مهمتك هي تنسيق النص كـ Markdown صحيح.

القواعد المهمة جداً:
- لا تعدل أبداً روابط الصور الموجودة
- الصور بصيغة: ![وصف](/uploads/blog/{userId}/{filename}) أو ![](blog-img-{uuid})
- احتفظ بالصور في مكانها الأصلي بدون تغيير
- أضف عناوين (# و ##) للأقسام الرئيسية
- استخدم النقاط والقوائم المرقمة حيث مناسب
- أضف **نص عريض** للنقاط المهمة
- أضف *نص مائل* للتوكيد
- حافظ على المعنى الأصلي للنص
- لا تضف معلومات جديدة
- أعد النص المنسق فقط بدون شرح`

	userMessage := fmt.Sprintf("نسق هذا النص كـ Markdown:\n\n%s", content)

	response, err := s.sendRequest(systemPrompt, userMessage)
	if err != nil {
		return nil, err
	}

	// Verify all images are preserved
	formattedContent := verifyImagesPreserved(response, images)

	return &FormatMarkdownResponse{FormattedContent: formattedContent}, nil
}

// CustomPrompt executes a custom AI prompt on content
func (s *AIService) CustomPrompt(content, prompt string) (*CustomPromptResponse, error) {
	// Extract existing images to preserve them
	images := extractImages(content)

	systemPrompt := `أنت مساعد كتابة ذكي باللغة العربية. نفذ طلب المستخدم على المحتوى المعطى.

القواعد المهمة جداً:
- لا تعدل أبداً روابط الصور الموجودة
- الصور بصيغة: ![وصف](/uploads/blog/{userId}/{filename}) أو ![](blog-img-{uuid})
- احتفظ بالصور في مكانها الأصلي بدون تغيير
- نفذ الطلب بدقة
- أعد النتيجة فقط بدون شرح إضافي`

	userMessage := fmt.Sprintf("الطلب: %s\n\nالمحتوى:\n%s", prompt, content)

	response, err := s.sendRequest(systemPrompt, userMessage)
	if err != nil {
		return nil, err
	}

	// Verify all images are preserved
	result := verifyImagesPreserved(response, images)

	return &CustomPromptResponse{Result: result}, nil
}

// Helper functions

func truncateForAI(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

// extractImages extracts all image markdown syntax from content
func extractImages(content string) []string {
	var images []string

	// Match ![...](...) pattern
	imgRegex := regexp.MustCompile(`!\[[^\]]*\]\([^)]+\)`)
	matches := imgRegex.FindAllString(content, -1)
	images = append(images, matches...)

	return images
}

// verifyImagesPreserved ensures all original images are in the formatted content
func verifyImagesPreserved(formatted string, originalImages []string) string {
	result := formatted

	for _, img := range originalImages {
		if !strings.Contains(result, img) {
			// Image was lost, append it at the end
			result = result + "\n\n" + img
		}
	}

	return result
}

// IsConfigured returns true if the AI service has an API key configured
func (s *AIService) IsConfigured() bool {
	return s.apiKey != ""
}
