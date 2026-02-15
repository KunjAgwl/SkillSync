package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeminiClient handles embedding requests to Gemini API
type GeminiClient struct {
	APIKey string
	Client *http.Client
}

// ===== Request / Response structs (CORRECT FORMAT) =====

type geminiEmbedRequest struct {
	Content              geminiContent `json:"content"`
	OutputDimensionality int           `json:"outputDimensionality,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

// ===== Constructor =====

func NewGemini(apiKey string) *GeminiClient {
	return &GeminiClient{
		APIKey: apiKey,
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ===== Embed function (THIS IS THE FIX) =====

func (c *GeminiClient) Embed(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("embedding input text cannot be empty")
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent?key=%s",
		c.APIKey,
	)

	payload := geminiEmbedRequest{
		Content: geminiContent{
			Parts: []geminiPart{
				{Text: text},
			},
		},
		OutputDimensionality: 768,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini embed error %d: %s", resp.StatusCode, string(b))
	}

	var result geminiEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Embedding.Values) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return result.Embedding.Values, nil
}
