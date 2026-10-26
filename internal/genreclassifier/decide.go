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

// Alternative is a diagnostic runner-up label. Score is the within-group
// share, a ranking number, not a calibrated probability.
type Alternative struct {
	Genre string  `json:"genre"`
	Score float64 `json:"score"`
}

// roundShare stabilises the emitted number: a float sum of probabilities is
// not exact, and byte-level golden tests need a value that does not wobble.
func roundShare(q float64) float64 { return math.Round(q*10000) / 10000 }

func pickTags(shares []Share, th Thresholds) []string {
	tags := []string{}
	if len(shares) == 0 {
		return tags
	}
	floor := th.TagMinShareRatio / float64(len(shares))
	for _, s := range shares {
		if len(tags) >= th.TagMax {
			break
		}
		if s.Q >= floor {
			tags = append(tags, s.Candidate.ID)
		}
	}
	return tags
}

// pickAlternatives reports the labels that lost to the chosen one. When the
// floor assigned the fallback genre nothing is excluded, so the leading
// candidates still show up — that is exactly when a reviewer needs them.
func pickAlternatives(shares []Share, max int, primary string) []Alternative {
	out := []Alternative{}
	for _, s := range shares {
		if len(out) >= max {
			break
		}
		if s.Candidate.ID == primary {
			continue
		}
		out = append(out, Alternative{Genre: s.Candidate.ID, Score: roundShare(s.Q)})
	}
	return out
}

// Decision is the step 1 verdict before it becomes contract JSON.
type Decision struct {
	PrimaryGenre  string
	Status        string
	SecondaryTags []string
	Alternatives  []Alternative
}

// Detail keeps the per-group shares so --debug and evaluation records can see
// what the thresholds looked at, plus which rule fired. It never enters the
// contract.
type Detail struct {
	Genre []Share
	Tags  []Share
	Rule  string
}

func Decide(scores map[string]float64, tax *Taxonomy) (Decision, Detail, error) {
	genreShares, err := groupShares(scores, tax.Candidates(GroupGenre))
	if err != nil {
		return Decision{}, Detail{}, err
	}

	var tagShares []Share
	if tags := tax.Candidates(GroupTag); len(tags) > 0 {
		tagShares, err = groupShares(scores, tags)
		if err != nil {
			return Decision{}, Detail{}, err
		}
	}

	call := genreCall(genreShares, tax.Thresholds, tax.FallbackGenre)
	return Decision{
		PrimaryGenre:  call.Primary,
		Status:        call.Status,
		SecondaryTags: pickTags(tagShares, tax.Thresholds),
		Alternatives:  pickAlternatives(genreShares, tax.Thresholds.GenreAlternativesMax, call.Primary),
	}, Detail{Genre: genreShares, Tags: tagShares, Rule: call.Rule}, nil
}
