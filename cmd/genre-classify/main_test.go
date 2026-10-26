package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseArgsDefaults(t *testing.T) {
	cfg, err := parseArgs([]string{"photo.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PhotoPath != "photo.jpg" {
		t.Errorf("PhotoPath = %q", cfg.PhotoPath)
	}
	if cfg.TaxonomyPath != "ai/genre-classifier/taxonomy/genre-v1.json" {
		t.Errorf("TaxonomyPath = %q, want the v1 default", cfg.TaxonomyPath)
	}
	if cfg.Model != "openai/clip-vit-large-patch14" {
		t.Errorf("Model = %q", cfg.Model)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.Debug {
		t.Error("Debug must default to false")
	}
}

func TestParseArgsFlags(t *testing.T) {
	cfg, err := parseArgs([]string{"-debug", "-timeout", "5s", "-model", "x/y", "-taxonomy", "t.json", "a.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Debug || cfg.Timeout != 5*time.Second || cfg.Model != "x/y" || cfg.TaxonomyPath != "t.json" {
		t.Errorf("flags not applied: %+v", cfg)
	}
}

func TestParseArgsRequiresExactlyOnePhoto(t *testing.T) {
	for _, args := range [][]string{{}, {"a.jpg", "b.jpg"}} {
		_, err := parseArgs(args)
		if err == nil {
			t.Errorf("parseArgs(%v) must fail", args)
		} else if !strings.Contains(err.Error(), "one photo") {
			t.Errorf("error = %q, want it to say one photo", err.Error())
		}
	}
}
