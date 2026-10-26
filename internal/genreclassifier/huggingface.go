package genreclassifier

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultHFBaseURL = "https://api-inference.huggingface.co/models/"
	DefaultModel     = "openai/clip-vit-large-patch14"

	maxResponseBytes = 1 << 20
)

// HuggingFace posts one image plus every candidate prompt and reads back one
// score per prompt. Attempts and Backoff are fields so tests can run the retry
// path without sleeping.
type HuggingFace struct {
	Model    string
	Token    string
	BaseURL  string
	HTTPC    *http.Client
	Backoff  time.Duration
	Attempts int
}

func NewHuggingFace(model, token string) (*HuggingFace, error) {
	if strings.TrimSpace(token) == "" {
		return nil, InputErrorf("HF_TOKEN is empty; export a Hugging Face API token before running")
	}
	if strings.TrimSpace(model) == "" {
		return nil, InputErrorf("model id is empty")
	}
	return &HuggingFace{
		Model:    model,
		Token:    token,
		BaseURL:  DefaultHFBaseURL,
		HTTPC:    &http.Client{},
		Backoff:  500 * time.Millisecond,
		Attempts: 2,
	}, nil
}

type hfRequest struct {
	Inputs     string       `json:"inputs"`
	Parameters hfParameters `json:"parameters"`
}

type hfParameters struct {
	Candidates []string `json:"candidate_labels"`
}

type hfResult struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

func (p *HuggingFace) Score(ctx context.Context, photo Photo, cands []Candidate) (map[string]float64, error) {
	prompts := make([]string, len(cands))
	for i, c := range cands {
		prompts[i] = c.Prompt
	}
	body, err := json.Marshal(hfRequest{
		Inputs:     "data:" + photo.MIME + ";base64," + base64.StdEncoding.EncodeToString(photo.Bytes),
		Parameters: hfParameters{Candidates: prompts},
	})
	if err != nil {
		return nil, err
	}
	delay := p.Backoff
	var lastErr error
	for attempt := 1; attempt <= p.Attempts; attempt++ {
		payload, status, retryWait, err := p.postOnce(ctx, body)
		switch {
		case err != nil:
			lastErr = ProviderErrorf("request to %s failed: %v", p.Model, err)
		case status == http.StatusOK:
			scores, derr := p.decode(payload, cands)
			if derr != nil {
				return nil, derr // a broken schema fails the same way on a second call
			}
			return scores, nil
		case !retryable(status):
			return nil, ProviderErrorf("HTTP %d from %s: %s", status, p.Model, snippet(payload))
		default:
			lastErr = ProviderErrorf("HTTP %d from %s: %s", status, p.Model, snippet(payload))
			if retryWait > 0 {
				delay = retryWait
			}
		}
		if attempt == p.Attempts {
			break
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ProviderErrorf("deadline while retrying %s: %v", p.Model, ctx.Err())
		}
		delay *= 2
	}
	return nil, lastErr
}

// retryable lists the statuses where asking again can plausibly help. 4xx other
// than 429 is a complaint about the request itself, so repeating it is a waste
// of the deadline.
func retryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (p *HuggingFace) postOnce(ctx context.Context, body []byte) ([]byte, int, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+p.Model, bytes.NewReader(body))
	if err != nil {
		return nil, 0, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := p.HTTPC.Do(req)
	if err != nil {
		return nil, 0, 0, err
	}
	defer res.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return nil, res.StatusCode, 0, err
	}
	return payload, res.StatusCode, retryAfter(res.Header.Get("Retry-After")), nil
}

// decode keys results by candidate id. A short list is fatal, not trimmed:
// per-group renormalization needs every score (spec §3.1). Scores arriving here
// are always finite — JSON cannot spell Infinity and the decoder refuses
// numbers that do not fit a float64.
func (p *HuggingFace) decode(payload []byte, cands []Candidate) (map[string]float64, error) {
	var rows []hfResult
	if err := json.Unmarshal(payload, &rows); err != nil {
		return nil, ProviderErrorf("response is not a [{label,score}] array: %v", err)
	}
	byPrompt := make(map[string]Candidate, len(cands))
	for _, c := range cands {
		byPrompt[c.Prompt] = c
	}
	scores := make(map[string]float64, len(cands))
	for _, r := range rows {
		c, ok := byPrompt[r.Label]
		if !ok {
			return nil, ProviderErrorf("model returned unknown label %q", r.Label)
		}
		scores[c.ID] = r.Score
	}
	if len(scores) != len(cands) {
		missing := []string{}
		for _, c := range cands {
			if _, ok := scores[c.ID]; !ok {
				missing = append(missing, c.ID)
			}
		}
		return nil, ProviderErrorf("model returned %d of %d candidates, missing %s",
			len(scores), len(cands), strings.Join(missing, ", "))
	}
	return scores, nil
}

func retryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	secs, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

func snippet(payload []byte) string {
	const limit = 200
	s := strings.TrimSpace(string(payload))
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
