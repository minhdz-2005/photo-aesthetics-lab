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
