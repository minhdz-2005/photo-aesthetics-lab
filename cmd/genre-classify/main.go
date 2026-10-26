package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	gc "github.com/minhdz-2005/photo-aesthetics-lab/internal/genreclassifier"
)

func main() {
	os.Exit(execute(os.Args[1:], os.Getenv("HF_TOKEN")))
}

func execute(args []string, token string) int {
	cfg, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "genre-classify: %v\n", err)
		return gc.ExitInput
	}
	provider, err := gc.NewHuggingFace(cfg.Model, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "genre-classify: %v\n", err)
		return gc.ExitInput
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	return gc.Run(ctx, cfg, gc.Deps{
		Provider: provider,
		Open:     gc.OpenPhoto,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
	})
}

func parseArgs(args []string) (gc.Config, error) {
	var cfg gc.Config
	fs := flag.NewFlagSet("genre-classify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	taxonomy := fs.String("taxonomy", "ai/genre-classifier/taxonomy/genre-v1.json",
		"path to the taxonomy file (relative to the repository root)")
	model := fs.String("model", gc.DefaultModel, "Hugging Face model id")
	timeout := fs.Duration("timeout", 30*time.Second, "deadline for the whole command, retries included")
	fs.BoolVar(&cfg.Debug, "debug", false, "write diagnostic metadata to stderr")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: genre-classify [-taxonomy file] [-model id] [-timeout dur] [-debug] <photo>")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return cfg, fmt.Errorf("need exactly one photo path, got %d", fs.NArg())
	}
	cfg.PhotoPath = fs.Arg(0)
	cfg.TaxonomyPath = *taxonomy
	cfg.Model = *model
	cfg.Timeout = *timeout
	return cfg, nil
}
