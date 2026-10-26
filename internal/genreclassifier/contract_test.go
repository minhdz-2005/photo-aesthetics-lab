package genreclassifier

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite the golden contract file")

func classifiedDecision() Decision {
	return Decision{
		PrimaryGenre:  "portrait",
		Status:        StatusClassified,
		SecondaryTags: []string{"outdoor"},
		Alternatives: []Alternative{
			{Genre: "landscape", Score: 0.2105},
			{Genre: "street", Score: 0.1579},
			{Genre: "food", Score: 0.1053},
		},
	}
}

func TestResultMatchesGolden(t *testing.T) {
	got, err := ResultOf(classifiedDecision()).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("testdata", "golden", "portrait-classified.json")
	if *updateGolden {
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Skipf("rewrote %s; review the diff", path)
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("contract bytes differ.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestResultKeysAndEmptyArrays(t *testing.T) {
	b, err := ResultOf(Decision{
		PrimaryGenre: "other", Status: StatusNeedsReview,
		SecondaryTags: []string{}, Alternatives: []Alternative{},
	}).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	out := string(b)
	for _, key := range []string{`"primary_genre"`, `"secondary_tags"`, `"alternatives"`, `"status"`} {
		if !strings.Contains(out, key) {
			t.Errorf("missing contract key %s in %s", key, out)
		}
	}
	// The contract asks for a list, never for null.
	if strings.Contains(out, "null") {
		t.Errorf("empty lists must serialize as [] not null:\n%s", out)
	}
	if !strings.HasSuffix(out, "}\n") {
		t.Error("contract output must end with exactly one newline")
	}
}
