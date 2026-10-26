package genreclassifier

import (
	"path/filepath"
	"reflect"
	"testing"
)

func loadMinV1(t *testing.T) *Taxonomy {
	t.Helper()
	tax, err := LoadTaxonomy(filepath.Join("testdata", "taxonomy", "min-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return tax
}

func TestDecideClassifiesClearPortrait(t *testing.T) {
	// Genre mass 0.95, tag mass 0.05; the union sums to 1.0 like a softmax
	// over all candidates at once.
	scores := map[string]float64{
		"portrait": 0.50, "landscape": 0.20, "street": 0.15, "food": 0.10,
		"outdoor": 0.03, "indoor": 0.02,
	}
	dec, detail, err := Decide(scores, loadMinV1(t))
	if err != nil {
		t.Fatal(err)
	}
	if dec.PrimaryGenre != "portrait" || dec.Status != StatusClassified {
		t.Errorf("got %s/%s, want portrait/classified", dec.PrimaryGenre, dec.Status)
	}
	if got, want := dec.SecondaryTags, []string{"outdoor"}; !reflect.DeepEqual(got, want) {
		t.Errorf("tags = %v, want %v", got, want)
	}
	// q = p / 0.95: 0.2105, 0.1579, 0.1053
	wantAlt := []Alternative{
		{Genre: "landscape", Score: 0.2105},
		{Genre: "street", Score: 0.1579},
		{Genre: "food", Score: 0.1053},
	}
	if !reflect.DeepEqual(dec.Alternatives, wantAlt) {
		t.Errorf("alternatives = %+v, want %+v", dec.Alternatives, wantAlt)
	}
	if len(detail.Genre) != 4 || len(detail.Tags) != 2 {
		t.Errorf("detail = %d genre, %d tag; want 4 and 2", len(detail.Genre), len(detail.Tags))
	}
}

func TestDecideAssignsOtherWhenFloorFails(t *testing.T) {
	scores := map[string]float64{
		"portrait": 0.24, "landscape": 0.23, "street": 0.22, "food": 0.21,
		"outdoor": 0.05, "indoor": 0.05,
	}
	dec, _, err := Decide(scores, loadMinV1(t))
	if err != nil {
		t.Fatal(err)
	}
	if dec.PrimaryGenre != "other" {
		t.Errorf("primary = %q, want other", dec.PrimaryGenre)
	}
	if dec.Status != StatusNeedsReview {
		t.Errorf("status = %q, want needs_review", dec.Status)
	}
	if len(dec.Alternatives) == 0 || dec.Alternatives[0].Genre != "portrait" {
		t.Errorf("diagnostics must survive a failed floor, got %+v", dec.Alternatives)
	}
	for _, a := range dec.Alternatives {
		if a.Genre == "other" {
			t.Error("other is assigned by rule, never a scored alternative")
		}
	}
}

func TestDecideRejectsMissingScore(t *testing.T) {
	scores := map[string]float64{
		"portrait": 0.5, "landscape": 0.2, "street": 0.15,
		"outdoor": 0.03, "indoor": 0.02, // food is missing
	}
	if _, _, err := Decide(scores, loadMinV1(t)); err == nil {
		t.Fatal("a missing candidate score must be an error")
	}
}

func TestDecideWithEmptyTagGroup(t *testing.T) {
	tax := loadMinV1(t)
	genreOnly := []Candidate{}
	for _, c := range tax.Cands {
		if c.Group == GroupGenre {
			genreOnly = append(genreOnly, c)
		}
	}
	tax.Cands = genreOnly

	dec, _, err := Decide(map[string]float64{
		"portrait": 0.5, "landscape": 0.2, "street": 0.15, "food": 0.1,
	}, tax)
	if err != nil {
		t.Fatal(err)
	}
	if dec.PrimaryGenre != "portrait" {
		t.Errorf("primary = %q, want portrait", dec.PrimaryGenre)
	}
	if dec.SecondaryTags == nil || len(dec.SecondaryTags) != 0 {
		t.Errorf("tags = %v, want empty non-nil slice", dec.SecondaryTags)
	}
}

func TestDecideIsScaleInvariant(t *testing.T) {
	tax := loadMinV1(t)
	base := map[string]float64{
		"portrait": 0.50, "landscape": 0.20, "street": 0.15, "food": 0.10,
		"outdoor": 0.03, "indoor": 0.02,
	}
	want, _, err := Decide(base, tax)
	if err != nil {
		t.Fatal(err)
	}
	for _, factor := range []float64{2, 0.125, 1000} {
		scaled := make(map[string]float64, len(base))
		for id, p := range base {
			scaled[id] = p * factor
		}
		got, _, err := Decide(scaled, tax)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("factor %v changed the decision: %+v vs %+v", factor, got, want)
		}
	}
}

func TestDecideGenreOutputIgnoresTagScores(t *testing.T) {
	tax := loadMinV1(t)
	a, _, err := Decide(map[string]float64{
		"portrait": 0.5, "landscape": 0.2, "street": 0.15, "food": 0.1,
		"outdoor": 0.049, "indoor": 0.001,
	}, tax)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := Decide(map[string]float64{
		"portrait": 0.5, "landscape": 0.2, "street": 0.15, "food": 0.1,
		"outdoor": 0.001, "indoor": 0.049,
	}, tax)
	if err != nil {
		t.Fatal(err)
	}
	if a.PrimaryGenre != b.PrimaryGenre || a.Status != b.Status {
		t.Errorf("genre verdict changed with tag scores: %s/%s vs %s/%s",
			a.PrimaryGenre, a.Status, b.PrimaryGenre, b.Status)
	}
	if !reflect.DeepEqual(a.Alternatives, b.Alternatives) {
		t.Errorf("alternatives moved across groups: %+v vs %+v", a.Alternatives, b.Alternatives)
	}
}
