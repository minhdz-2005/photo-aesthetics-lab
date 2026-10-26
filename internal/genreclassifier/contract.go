package genreclassifier

import "encoding/json"

// Result is the step 1 contract, verbatim from docs/step-01. Field order here
// fixes the byte order of the golden file.
type Result struct {
	PrimaryGenre  string        `json:"primary_genre"`
	SecondaryTags []string      `json:"secondary_tags"`
	Alternatives  []Alternative `json:"alternatives"`
	Status        string        `json:"status"`
}

func ResultOf(d Decision) Result {
	return Result{
		PrimaryGenre:  d.PrimaryGenre,
		SecondaryTags: d.SecondaryTags,
		Alternatives:  d.Alternatives,
		Status:        d.Status,
	}
}

// Bytes appends one newline so a piped result stays one valid JSON document.
func (r Result) Bytes() ([]byte, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
