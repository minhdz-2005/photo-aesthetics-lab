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

// LoadTaxonomy reads, renders and validates a taxonomy file. Broken files fail
// here rather than at score time, where they would only look like odd results.
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
	if err := tax.validate(); err != nil {
		return nil, err
	}
	return &tax, nil
}

// validate rejects taxonomies that would otherwise misbehave at score time.
// seenID and seenPrompt are global, not per group: the provider maps results
// back by id over the union of both groups, and by prompt text, so both must
// be unique across the whole file.
func (t *Taxonomy) validate() error {
	switch {
	case t.Version == "":
		return InputErrorf("taxonomy %s: version is empty", t.path)
	case t.PromptTemplate == "":
		return InputErrorf("taxonomy %s: prompt_template is empty", t.path)
	case !strings.Contains(t.PromptTemplate, "{label}"):
		return InputErrorf("taxonomy %s: prompt_template must contain {label}", t.path)
	case t.FallbackGenre == "":
		return InputErrorf("taxonomy %s: fallback_genre is empty", t.path)
	}
	if len(t.Candidates(GroupGenre)) < 2 {
		return InputErrorf("taxonomy %s: need at least 2 genre candidates, got %d",
			t.path, len(t.Candidates(GroupGenre)))
	}

	seenID := map[string]bool{}
	seenPrompt := map[string]string{}
	for _, c := range t.Cands {
		if c.ID == "" || c.Label == "" {
			return InputErrorf("taxonomy %s: candidate %q has empty id or label", t.path, c.ID)
		}
		if c.Group != GroupGenre && c.Group != GroupTag {
			return InputErrorf("taxonomy %s: candidate %q has unknown group %q", t.path, c.ID, c.Group)
		}
		if c.ID == t.FallbackGenre {
			return InputErrorf("taxonomy %s: fallback_genre %q must not be a candidate", t.path, c.ID)
		}
		if seenID[c.ID] {
			return InputErrorf("taxonomy %s: duplicate id %q", t.path, c.ID)
		}
		seenID[c.ID] = true
		if other, dup := seenPrompt[c.Prompt]; dup {
			return InputErrorf("taxonomy %s: candidates %q and %q render to the same prompt %q",
				t.path, other, c.ID, c.Prompt)
		}
		seenPrompt[c.Prompt] = c.ID
	}

	th := t.Thresholds
	if th.GenreMinShareRatio < 0 || th.GenreMinMarginRatio < 0 || th.TagMinShareRatio < 0 {
		return InputErrorf("taxonomy %s: share and margin ratios must be >= 0", t.path)
	}
	if th.GenreAlternativesMax <= 0 {
		return InputErrorf("taxonomy %s: genre_alternatives_max must be > 0, got %d", t.path, th.GenreAlternativesMax)
	}
	if th.TagMax <= 0 {
		return InputErrorf("taxonomy %s: tag_max must be > 0, got %d", t.path, th.TagMax)
	}
	return nil
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
