package genreclassifier

import (
	"math"
	"sort"
)

// Share is one candidate's score inside its own group. Raw is whatever the
// provider returned; Q is Raw divided by the group total, which cancels the
// shared softmax partition function and makes cross-group comparison invalid.
type Share struct {
	Candidate Candidate
	Raw       float64
	Q         float64
}

func groupShares(scores map[string]float64, cands []Candidate) ([]Share, error) {
	out := make([]Share, 0, len(cands))
	var sum float64
	for _, c := range cands {
		p, ok := scores[c.ID]
		if !ok {
			return nil, ProviderErrorf("model did not return a score for candidate %q", c.ID)
		}
		if math.IsNaN(p) || math.IsInf(p, 0) {
			return nil, ProviderErrorf("non-finite score %v for candidate %q", p, c.ID)
		}
		out = append(out, Share{Candidate: c, Raw: p})
		sum += p
	}
	if sum <= 0 {
		return nil, ProviderErrorf("candidate group has no score mass to renormalize")
	}
	for i := range out {
		out[i].Q = out[i].Raw / sum
	}
	// SliceStable plus a score-only predicate keeps equal scores in the order
	// the taxonomy declares them, which is what makes the tie-break deterministic.
	sort.SliceStable(out, func(i, j int) bool { return out[i].Raw > out[j].Raw })
	return out, nil
}
