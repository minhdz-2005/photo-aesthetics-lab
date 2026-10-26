package genreclassifier

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type captured struct {
	auth       string
	contentTyp string
	labels     []string
	inputs     string
}

func newProvider(t *testing.T, status int, body string, got *captured) (*HuggingFace, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var req struct {
			Inputs     string `json:"inputs"`
			Parameters struct {
				Candidates []string `json:"candidate_labels"`
			} `json:"parameters"`
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Errorf("request body is not the expected JSON: %v", err)
		}
		got.auth = r.Header.Get("Authorization")
		got.contentTyp = r.Header.Get("Content-Type")
		got.labels = req.Parameters.Candidates
		got.inputs = req.Inputs

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		io.WriteString(w, body)
	}))
	p, err := NewHuggingFace("test/model", "tok-123")
	if err != nil {
		t.Fatal(err)
	}
	p.BaseURL = srv.URL + "/models/"
	p.Backoff = 0
	p.Attempts = 1
	return p, srv
}

func testPhoto() Photo {
	return Photo{Path: "p.jpg", Bytes: []byte{0xFF, 0xD8, 0xFF, 0x00}, MIME: "image/jpeg", Width: 4, Height: 3}
}

func TestHFScoreMapsLabelsBackToIDs(t *testing.T) {
	cands := []Candidate{
		{Group: GroupGenre, ID: "portrait", Prompt: "a photo of a person"},
		{Group: GroupGenre, ID: "landscape", Prompt: "a photo of a landscape"},
		{Group: GroupTag, ID: "outdoor", Prompt: "a photo of outdoors"},
	}
	var got captured
	body := `[
		{"label":"a photo of a landscape","score":0.2},
		{"label":"a photo of a person","score":0.5},
		{"label":"a photo of outdoors","score":0.3}
	]`
	p, srv := newProvider(t, http.StatusOK, body, &got)
	defer srv.Close()

	scores, err := p.Score(context.Background(), testPhoto(), cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(scores) != 3 {
		t.Fatalf("scores = %v, want 3 entries keyed by id", scores)
	}
	if scores["portrait"] != 0.5 || scores["landscape"] != 0.2 || scores["outdoor"] != 0.3 {
		t.Errorf("scores = %v, want portrait 0.5, landscape 0.2, outdoor 0.3", scores)
	}
	if got.auth != "Bearer tok-123" {
		t.Errorf("Authorization = %q", got.auth)
	}
	if got.contentTyp != "application/json" {
		t.Errorf("Content-Type = %q", got.contentTyp)
	}
	// The order sent is the taxonomy order, not the order of the response.
	if strings.Join(got.labels, "|") != "a photo of a person|a photo of a landscape|a photo of outdoors" {
		t.Errorf("candidate_labels = %v", got.labels)
	}
	want := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(testPhoto().Bytes)
	if got.inputs != want {
		t.Errorf("inputs = %q, want %q", got.inputs, want)
	}
}

func TestHFScoreRejectsUnusableResponses(t *testing.T) {
	cands := []Candidate{
		{Group: GroupGenre, ID: "portrait", Prompt: "a person"},
		{Group: GroupGenre, ID: "landscape", Prompt: "a landscape"},
	}
	cases := []struct {
		name    string
		status  int
		body    string
		wantMsg string
	}{
		{"missing candidate", http.StatusOK, `[{"label":"a person","score":0.9}]`, "of 2 candidates"},
		{"unknown label", http.StatusOK, `[{"label":"a person","score":0.5},{"label":"a cat","score":0.5}]`, "unknown label"},
		{"not an array", http.StatusOK, `{"error":"model is loading"}`, "not a [{label,score}] array"},
		// JSON has no Infinity literal, and Go's decoder refuses a number that
		// does not fit a float64, so a non-finite score can never reach decide.
		{"out-of-range number", http.StatusOK, `[{"label":"a person","score":1e400},{"label":"a landscape","score":0.1}]`, "not a [{label,score}] array"},
		{"server error", http.StatusInternalServerError, `{"error":"boom"}`, "HTTP 500"},
		{"bad request", http.StatusBadRequest, `{"error":"invalid input"}`, "HTTP 400"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got captured
			p, srv := newProvider(t, c.status, c.body, &got)
			defer srv.Close()

			_, err := p.Score(context.Background(), testPhoto(), cands)
			if err == nil {
				t.Fatalf("want error for %s", c.name)
			}
			if !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), c.wantMsg)
			}
			if !isProvider(err) {
				t.Errorf("error %v must be ProviderError so exit code is 2", err)
			}
		})
	}
}

func TestNewHuggingFaceRequiresToken(t *testing.T) {
	if _, err := NewHuggingFace("test/model", ""); err == nil {
		t.Fatal("empty token must fail")
	} else if !isInput(err) {
		t.Errorf("empty token is our mistake, not the world's: got %T", err)
	}
}
