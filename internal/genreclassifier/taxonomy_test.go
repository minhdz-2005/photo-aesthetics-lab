package genreclassifier

import (
	"path/filepath"
	"testing"
)

func TestLoadTaxonomyReadsValidFile(t *testing.T) {
	tax, err := LoadTaxonomy(filepath.Join("testdata", "taxonomy", "min-v1.json"))
	if err != nil {
		t.Fatalf("LoadTaxonomy: %v", err)
	}
	if tax.Version != "min-v1" {
		t.Errorf("Version = %q, want %q", tax.Version, "min-v1")
	}
	if got, want := len(tax.Candidates(GroupGenre)), 4; got != want {
		t.Errorf("genre candidates = %d, want %d", got, want)
	}
	if got, want := len(tax.Candidates(GroupTag)), 2; got != want {
		t.Errorf("tag candidates = %d, want %d", got, want)
	}
	if got, want := tax.Thresholds.GenreMinShareRatio, 1.5; got != want {
		t.Errorf("GenreMinShareRatio = %v, want %v", got, want)
	}
	if got, want := tax.FallbackGenre, "other"; got != want {
		t.Errorf("FallbackGenre = %q, want %q", got, want)
	}
	if got, want := tax.Prompt("outdoors"), "a photo of outdoors"; got != want {
		t.Errorf("Prompt = %q, want %q", got, want)
	}
}

// The production taxonomy is data, not code, so it needs a test of its own.
// Loading it runs the same validate() the CLI runs, so an edit that makes
// "other" compete as a candidate fails here rather than in production.
func TestProductionTaxonomyLoads(t *testing.T) {
	tax, err := LoadTaxonomy(filepath.Join("..", "..", "ai", "genre-classifier", "taxonomy", "genre-v1.json"))
	if err != nil {
		t.Fatalf("genre-v1.json must load: %v", err)
	}
	if got, want := len(tax.Candidates(GroupGenre)), 7; got != want {
		t.Errorf("genre candidates = %d, want %d", got, want)
	}
	if got, want := len(tax.Candidates(GroupTag)), 2; got != want {
		t.Errorf("tag candidates = %d, want %d", got, want)
	}
}
