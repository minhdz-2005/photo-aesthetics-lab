package genreclassifier

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Group string

const (
	GroupGenre Group = "genre"
	GroupTag   Group = "tag"
)

// Candidate is one label the model may choose from. Prompt is filled by
// LoadTaxonomy; the provider maps model results back to ID by Prompt text,
// so two candidates must never render to the same Prompt.
type Candidate struct {
	Group  Group  `json:"group"`
	ID     string `json:"id"`
	Label  string `json:"label"`
	Prompt string `json:"-"`
}

type Thresholds struct {
	GenreMinShareRatio   float64 `json:"genre_min_share_ratio"`
	GenreMinMarginRatio  float64 `json:"genre_min_margin_ratio"`
	GenreAlternativesMax int     `json:"genre_alternatives_max"`
	TagMinShareRatio     float64 `json:"tag_min_share_ratio"`
	TagMax               int     `json:"tag_max"`
}

type Taxonomy struct {
	Version        string      `json:"version"`
	PromptTemplate string      `json:"prompt_template"`
	FallbackGenre  string      `json:"fallback_genre"`
	Thresholds     Thresholds  `json:"thresholds"`
	Cands          []Candidate `json:"candidates"`

	path string
}

// LoadTaxonomy reads and renders candidates. Validation lands in Task 2.
func LoadTaxonomy(path string) (*Taxonomy, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, InputErrorf("taxonomy file %s: %v", abs, err)
	}
	var tax Taxonomy
	if err := json.Unmarshal(raw, &tax); err != nil {
		return nil, InputErrorf("taxonomy file %s: %v", abs, err)
	}
	tax.path = abs
	for i := range tax.Cands {
		tax.Cands[i].Prompt = tax.Prompt(tax.Cands[i].Label)
	}
	return &tax, nil
}

func (t *Taxonomy) Candidates(g Group) []Candidate {
	out := []Candidate{}
	for _, c := range t.Cands {
		if c.Group == g {
			out = append(out, c)
		}
	}
	return out
}

func (t *Taxonomy) Prompt(label string) string {
	return strings.ReplaceAll(t.PromptTemplate, "{label}", label)
}
