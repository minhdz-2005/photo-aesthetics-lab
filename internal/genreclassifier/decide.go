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

const (
	StatusClassified  = "classified"
	StatusNeedsReview = "needs_review"
)

// GenreCall is the verdict of the two genre thresholds. Rule records which
// threshold produced it, for --debug output and for evaluation records.
type GenreCall struct {
	Primary string
	Status  string
	Rule    string
}

// Thresholds are multiples of the uniform share, never absolute probabilities:
// that is the only form that still means something after renormalizing inside
// a group and after the taxonomy changes its number of labels.
func genreCall(shares []Share, th Thresholds, fallback string) GenreCall {
	uniform := 1 / float64(len(shares))
	leader := shares[0].Candidate.ID

	if shares[0].Q < th.GenreMinShareRatio*uniform {
		return GenreCall{Primary: fallback, Status: StatusNeedsReview, Rule: "floor"}
	}
	if len(shares) < 2 {
		// No runner-up to separate from; the floor above still guards.
		return GenreCall{Primary: leader, Status: StatusClassified, Rule: "single candidate"}
	}
	if shares[0].Q-shares[1].Q < th.GenreMinMarginRatio*uniform {
		return GenreCall{Primary: leader, Status: StatusNeedsReview, Rule: "margin"}
	}
	return GenreCall{Primary: leader, Status: StatusClassified, Rule: "floor+margin"}
}
