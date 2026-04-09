package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const ollamaBase = "http://localhost:11434"

type Client struct {
	Model        string
	VoiceProfile string
	httpClient   *http.Client
}

func NewClient(model, voiceProfile string) *Client {
	return &Client{
		Model:        model,
		VoiceProfile: voiceProfile,
		httpClient:   &http.Client{Timeout: 120 * time.Second},
	}
}

// IsAvailable checks if Ollama is running locally.
func IsAvailable() bool {
	c := &http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get(ollamaBase + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

type tagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ListModels returns names of locally available Ollama models.
func ListModels() ([]string, error) {
	c := &http.Client{Timeout: 5 * time.Second}
	resp, err := c.Get(ollamaBase + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tr tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}

	models := make([]string, 0, len(tr.Models))
	for _, m := range tr.Models {
		models = append(models, m.Name)
	}
	return models, nil
}

// RecommendModel returns the model to use based on available RAM.
func RecommendModel() (model, reason string) {
	// macOS: use sysctl to infer RAM; fall back to llama3
	// For now we use a simple heuristic. Real implementation would call sysctl.
	// We default to llama3 but the onboarding can override.
	return "llama3", "default model for 8GB+ RAM systems"
}

type generateReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResp struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (c *Client) generate(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(generateReq{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ollamaBase+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned HTTP %d", resp.StatusCode)
	}

	var gr generateResp
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", err
	}

	return strings.TrimSpace(gr.Response), nil
}

// Polish rewrites a journal entry in the configured voice.
func (c *Client) Polish(ctx context.Context, rawText string) (string, error) {
	wordCount := len(strings.Fields(rawText))
	maxWords := int(float64(wordCount) * 1.5)

	prompt := fmt.Sprintf(
		"Rewrite the following journal entry in first-person prose.\n"+
			"Preserve all facts and details. Write in this voice: %s\n"+
			"Do not add events or emotions that aren't implied by the original.\n"+
			"Keep it under %d words.\n\nJournal entry:\n%s",
		c.VoiceProfile, maxWords, rawText,
	)

	return c.generate(ctx, prompt)
}

// GeneratePrompt creates a personalized writing prompt based on recent entry snippets.
func (c *Client) GeneratePrompt(ctx context.Context, recentSnippets []string) (string, error) {
	var preamble string
	if len(recentSnippets) > 0 {
		if len(recentSnippets) > 3 {
			recentSnippets = recentSnippets[:3]
		}
		preamble = "Based on these recent journal themes:\n" + strings.Join(recentSnippets, "\n---\n") + "\n\n"
	}

	prompt := preamble + "Generate a single, specific, personal journaling prompt that will help someone reflect on their day. Make it concrete and interesting, not generic. Just give the prompt, nothing else."
	return c.generate(ctx, prompt)
}
