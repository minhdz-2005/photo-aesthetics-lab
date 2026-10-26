//go:build integration

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	gc "github.com/minhdz-2005/photo-aesthetics-lab/internal/genreclassifier"
)

// TestLiveModelScoresEveryCandidate checks the two wire assumptions of spec §12:
// the endpoint returns a score for all candidates (not top-k) and accepts the
// base64 data-url body. Run with:
//
//	go test -tags integration ./cmd/genre-classify/ -v
func TestLiveModelScoresEveryCandidate(t *testing.T) {
	token := os.Getenv("HF_TOKEN")
	if token == "" {
		t.Skip("HF_TOKEN not set")
	}
	path := os.Getenv("GENRE_TEST_PHOTO")
	if path == "" {
		t.Skip("GENRE_TEST_PHOTO not set (path to a real .jpg or .png)")
	}

	photo, err := gc.OpenPhoto(path)
	if err != nil {
		t.Fatal(err)
	}
	tax, err := gc.LoadTaxonomy(filepath.Join("..", "..", "ai", "genre-classifier", "taxonomy", "genre-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	provider, err := gc.NewHuggingFace(gc.DefaultModel, token)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	scores, err := provider.Score(ctx, photo, tax.Cands)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	if len(scores) != len(tax.Cands) {
		t.Fatalf("got %d scores for %d candidates: the endpoint truncates, so the merged call is invalid",
			len(scores), len(tax.Cands))
	}
	for _, c := range tax.Cands {
		if _, ok := scores[c.ID]; !ok {
			t.Errorf("no score for %q", c.ID)
		}
	}

	dec, detail, err := gc.Decide(scores, tax)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("photo=%s %dx%d", photo.Path, photo.Width, photo.Height)
	for _, s := range detail.Genre {
		t.Logf("genre %-13s raw=%.4f q=%.4f", s.Candidate.ID, s.Raw, s.Q)
	}
	for _, s := range detail.Tags {
		t.Logf("tag   %-13s raw=%.4f q=%.4f", s.Candidate.ID, s.Raw, s.Q)
	}
	t.Logf("-> primary=%s status=%s rule=%s tags=%v", dec.PrimaryGenre, dec.Status, detail.Rule, dec.SecondaryTags)
}
