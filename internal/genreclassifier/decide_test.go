package genreclassifier

import (
	"math"
	"strings"
	"testing"
)

func genreCands(ids ...string) []Candidate {
	out := make([]Candidate, 0, len(ids))
	for _, id := range ids {
		out = append(out, Candidate{Group: GroupGenre, ID: id, Label: id, Prompt: "a photo of " + id})
	}
	return out
}

func ids(shares []Share) string {
	parts := make([]string, 0, len(shares))
	for _, s := range shares {
		parts = append(parts, s.Candidate.ID)
	}
	return strings.Join(parts, ",")
}

func TestGroupSharesRenormalizesWithinGroup(t *testing.T) {
	// 3 and 1 are exact in binary floating point, so the expected shares 0.75
	// and 0.25 can be compared with == instead of a tolerance.
	shares, err := groupShares(map[string]float64{"a": 3, "b": 1}, genreCands("a", "b"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ids(shares), "a,b"; got != want {
		t.Errorf("order = %q, want %q", got, want)
	}
	if shares[0].Q != 0.75 || shares[1].Q != 0.25 {
		t.Errorf("Q = %v, %v; want 0.75, 0.25", shares[0].Q, shares[1].Q)
	}
	if shares[0].Raw != 3 {
		t.Errorf("Raw = %v, want the untouched provider score 3", shares[0].Raw)
	}
}

func TestGroupSharesKeepsDeclarationOrderOnTies(t *testing.T) {
	shares, err := groupShares(
		map[string]float64{"c": 0.2, "a": 0.2, "b": 0.2},
		genreCands("a", "b", "c"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ids(shares), "a,b,c"; got != want {
		t.Errorf("order = %q, want declaration order %q", got, want)
	}
}

func TestGroupSharesRejectsMissingScore(t *testing.T) {
	_, err := groupShares(map[string]float64{"a": 0.3}, genreCands("a", "b"))
	if err == nil {
		t.Fatal("missing score for b must be an error")
	}
	if !strings.Contains(err.Error(), `"b"`) {
		t.Errorf("error should name the missing candidate, got %q", err.Error())
	}
	if !isProvider(err) {
		t.Errorf("error %v must be ProviderError so exit code is 2", err)
	}
}

func TestGroupSharesRejectsZeroMass(t *testing.T) {
	_, err := groupShares(map[string]float64{"a": 0, "b": 0}, genreCands("a", "b"))
	if err == nil {
		t.Fatal("a group with no score mass must not be renormalized")
	}
	if !strings.Contains(err.Error(), "no score mass") {
		t.Errorf("error = %q, want substring %q", err.Error(), "no score mass")
	}
}

func TestGroupSharesRejectsNonFiniteScore(t *testing.T) {
	_, err := groupShares(map[string]float64{"a": 0.3, "b": math.Inf(1)}, genreCands("a", "b"))
	if err == nil {
		t.Fatal("+Inf score must be an error")
	}
}

func shares(q ...float64) []Share {
	out := make([]Share, 0, len(q))
	for i, v := range q {
		out = append(out, Share{
			Candidate: Candidate{Group: GroupGenre, ID: string(rune('a' + i))},
			Raw:       v,
			Q:         v,
		})
	}
	return out
}

var minV1Thresholds = Thresholds{
	GenreMinShareRatio: 1.5, GenreMinMarginRatio: 0.5,
	GenreAlternativesMax: 3, TagMinShareRatio: 1.0, TagMax: 3,
}

func TestGenreCall(t *testing.T) {
	cases := []struct {
		name        string
		q           []float64
		wantPrimary string
		wantStatus  string
		wantRule    string
	}{
		{
			name:        "clears floor and margin",
			q:           []float64{0.55, 0.20, 0.15, 0.10},
			wantPrimary: "a", wantStatus: StatusClassified, wantRule: "floor+margin",
		},
		{
			// N=4 so uniform is 0.25 and the floor is 0.375: 0.28 misses it,
			// so the margin is never consulted.
			name:        "misses floor",
			q:           []float64{0.28, 0.26, 0.24, 0.22},
			wantPrimary: "other", wantStatus: StatusNeedsReview, wantRule: "floor",
		},
		{
			// 0.40 clears the 0.375 floor but leads by 0.05, under the 0.125 margin.
			name:        "misses margin",
			q:           []float64{0.40, 0.35, 0.15, 0.10},
			wantPrimary: "a", wantStatus: StatusNeedsReview, wantRule: "margin",
		},
		{
			name:        "exact tie",
			q:           []float64{0.5, 0.5, 0, 0},
			wantPrimary: "a", wantStatus: StatusNeedsReview, wantRule: "margin",
		},
		{
			// N=1 means the only share is 1.0, which a 1.5x floor can never pass.
			name:        "single candidate still guarded by floor",
			q:           []float64{1},
			wantPrimary: "other", wantStatus: StatusNeedsReview, wantRule: "floor",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := genreCall(shares(c.q...), minV1Thresholds, "other")
			if got.Primary != c.wantPrimary {
				t.Errorf("Primary = %q, want %q", got.Primary, c.wantPrimary)
			}
			if got.Status != c.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, c.wantStatus)
			}
			if got.Rule != c.wantRule {
				t.Errorf("Rule = %q, want %q", got.Rule, c.wantRule)
			}
		})
	}
}

func TestGenreCallSingleCandidateSkipsMargin(t *testing.T) {
	// With N=1 the only share is 1.0, so minV1's 1.5x floor can never be
	// cleared. Dropping the floor to 1.0 exercises the no-runner-up branch
	// instead of leaving it dead code.
	th := minV1Thresholds
	th.GenreMinShareRatio = 1.0
	got := genreCall(shares(1), th, "other")
	if got.Primary != "a" || got.Status != StatusClassified || got.Rule != "single candidate" {
		t.Errorf("got %+v, want a / classified / single candidate", got)
	}
}

func TestGenreCallUsesThresholdsFromTaxonomy(t *testing.T) {
	q := []float64{0.40, 0.35, 0.15, 0.10}
	if got := genreCall(shares(q...), minV1Thresholds, "other"); got.Status != StatusNeedsReview {
		t.Fatalf("baseline must be needs_review, got %q", got.Status)
	}
	loose := minV1Thresholds
	loose.GenreMinMarginRatio = 0.1
	if got := genreCall(shares(q...), loose, "other"); got.Status != StatusClassified {
		t.Errorf("lowering the margin ratio must promote the call, got %q", got.Status)
	}
}
