package genreclassifier

import (
	"context"
	"fmt"
	"io"
	"time"
)

const (
	ExitOK       = 0
	ExitInput    = 1
	ExitProvider = 2
)

// Config carries the flags. Timeout is not read by Run: the caller owns the
// context deadline, so injecting a provider that ignores ctx stays a caller bug.
type Config struct {
	PhotoPath    string
	TaxonomyPath string
	Model        string
	Timeout      time.Duration
	Debug        bool
}

// Deps is the injection seam that keeps Run testable: the provider, the photo
// opener and the two streams. Nothing else reaches the filesystem or network.
type Deps struct {
	Provider Provider
	Open     func(path string) (Photo, error)
	Stdout   io.Writer
	Stderr   io.Writer
}

func Run(ctx context.Context, cfg Config, d Deps) int {
	if err := classify(ctx, cfg, d); err != nil {
		fmt.Fprintf(d.Stderr, "genre-classify: %v\n", err)
		if isInput(err) {
			return ExitInput
		}
		return ExitProvider
	}
	return ExitOK
}

func classify(ctx context.Context, cfg Config, d Deps) error {
	photo, err := d.Open(cfg.PhotoPath)
	if err != nil {
		return err
	}
	tax, err := LoadTaxonomy(cfg.TaxonomyPath)
	if err != nil {
		return err
	}
	scores, err := d.Provider.Score(ctx, photo, tax.Cands)
	if err != nil {
		return err
	}
	dec, detail, err := Decide(scores, tax)
	if err != nil {
		return err
	}
	out, err := ResultOf(dec).Bytes()
	if err != nil {
		return err
	}
	if cfg.Debug {
		writeDebug(d.Stderr, cfg, photo, tax, detail, dec)
	}
	_, err = d.Stdout.Write(out)
	return err
}

// writeDebug records what the contract cannot carry: taxonomy_version, the
// model that scored the photo, the shares the thresholds saw, and which rule
// fired (spec §5.5, §8).
func writeDebug(w io.Writer, cfg Config, photo Photo, tax *Taxonomy, detail Detail, dec Decision) {
	fmt.Fprintf(w, "taxonomy_version=%s model=%s\n", tax.Version, cfg.Model)
	fmt.Fprintf(w, "photo=%s size=%dx%d mime=%s\n", cfg.PhotoPath, photo.Width, photo.Height, photo.MIME)
	for _, s := range detail.Genre {
		fmt.Fprintf(w, "genre %s raw=%.4f q=%.4f\n", s.Candidate.ID, s.Raw, s.Q)
	}
	for _, s := range detail.Tags {
		fmt.Fprintf(w, "tag %s raw=%.4f q=%.4f\n", s.Candidate.ID, s.Raw, s.Q)
	}
	fmt.Fprintf(w, "primary=%s status=%s rule=%s tags=%v\n",
		dec.PrimaryGenre, dec.Status, detail.Rule, dec.SecondaryTags)
}
