package genreclassifier

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTaxonomyRejectsInvalidFiles(t *testing.T) {
	cases := []struct {
		file    string
		wantMsg string
	}{
		{"bad-no-placeholder.json", "prompt_template must contain {label}"},
		{"bad-single-genre.json", "at least 2 genre candidates"},
		{"bad-fallback-in-candidates.json", "must not be a candidate"},
		{"bad-duplicate-id.json", "duplicate id"},
		{"bad-duplicate-prompt.json", "same prompt"},
		{"bad-negative-threshold.json", "must be >= 0"},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			_, err := LoadTaxonomy(filepath.Join("testdata", "taxonomy", c.file))
			if err == nil {
				t.Fatalf("LoadTaxonomy(%s) = nil error, want failure", c.file)
			}
			if !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), c.wantMsg)
			}
			if !isInput(err) {
				t.Errorf("error %v is not an InputError (exit code would be wrong)", err)
			}
		})
	}
}

func TestLoadTaxonomyAcceptsValidFile(t *testing.T) {
	if _, err := LoadTaxonomy(filepath.Join("testdata", "taxonomy", "min-v1.json")); err != nil {
		t.Fatalf("min-v1.json must load: %v", err)
	}
}
