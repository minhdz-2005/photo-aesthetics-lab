package genreclassifier

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

type fakeProvider struct {
	scores map[string]float64
	err    error
	calls  int
}

func (f *fakeProvider) Score(ctx context.Context, photo Photo, cands []Candidate) (map[string]float64, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.scores, nil
}

func goodScores() map[string]float64 {
	return map[string]float64{
		"portrait": 0.50, "landscape": 0.20, "street": 0.15, "food": 0.10,
		"outdoor": 0.03, "indoor": 0.02,
	}
}

func pipelineFixture(t *testing.T) (Config, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	return Config{
		PhotoPath:    "testdata/photos/portrait.jpg",
		TaxonomyPath: "testdata/taxonomy/min-v1.json",
		Model:        "test/model",
	}, &bytes.Buffer{}, &bytes.Buffer{}
}

func testDeps(fp Provider, out, errw io.Writer) Deps {
	return Deps{
		Provider: fp,
		Open: func(path string) (Photo, error) {
			return Photo{Path: path, Bytes: []byte{1}, MIME: "image/jpeg", Width: 4, Height: 3}, nil
		},
		Stdout: out,
		Stderr: errw,
	}
}

func TestRunWritesContractToStdout(t *testing.T) {
	cfg, out, errOut := pipelineFixture(t)
	fp := &fakeProvider{scores: goodScores()}
	code := Run(context.Background(), cfg, testDeps(fp, out, errOut))

	if code != ExitOK {
		t.Fatalf("exit = %d, want 0; stderr=%s", code, errOut.String())
	}
	if !strings.HasPrefix(out.String(), "{\n") {
		t.Errorf("stdout must be the contract JSON only, got:\n%s", out.String())
	}
	if errOut.Len() != 0 {
		t.Errorf("stderr must be empty without --debug, got:\n%s", errOut.String())
	}
	if fp.calls != 1 {
		t.Errorf("provider calls = %d, want exactly 1 merged call", fp.calls)
	}
}

func TestRunExitCodes(t *testing.T) {
	cases := []struct {
		name     string
		scores   map[string]float64
		provider error
		openErr  error
		wantCode int
	}{
		{"needs_review still exits 0", map[string]float64{
			"portrait": 0.24, "landscape": 0.23, "street": 0.22, "food": 0.21,
			"outdoor": 0.05, "indoor": 0.05,
		}, nil, nil, ExitOK},
		{"missing score exits 2", map[string]float64{"portrait": 0.5}, nil, nil, ExitProvider},
		{"provider error exits 2", nil, ProviderErrorf("HTTP 503"), nil, ExitProvider},
		{"bad photo exits 1", nil, nil, InputErrorf("photo bad: unreadable"), ExitInput},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg, out, errOut := pipelineFixture(t)
			deps := testDeps(&fakeProvider{scores: c.scores, err: c.provider}, out, errOut)
			if c.openErr != nil {
				deps.Open = func(path string) (Photo, error) { return Photo{}, c.openErr }
			}
			got := Run(context.Background(), cfg, deps)
			if got != c.wantCode {
				t.Errorf("exit = %d, want %d\nstdout:%s\nstderr:%s", got, c.wantCode, out, errOut)
			}
			if c.wantCode != ExitOK && out.Len() != 0 {
				t.Errorf("stdout must be empty on failure, got:\n%s", out.String())
			}
		})
	}
}

func TestRunReportsMissingTaxonomyFile(t *testing.T) {
	cfg, out, errOut := pipelineFixture(t)
	cfg.TaxonomyPath = "testdata/taxonomy/does-not-exist.json"
	deps := testDeps(&fakeProvider{scores: goodScores()}, out, errOut)

	if code := Run(context.Background(), cfg, deps); code != ExitInput {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut.String(), "does-not-exist.json") {
		t.Errorf("stderr should name the file, got %q", errOut.String())
	}
}

func TestRunDebugGoesToStderrOnly(t *testing.T) {
	cfg, out, errOut := pipelineFixture(t)
	cfg.Debug = true
	deps := testDeps(&fakeProvider{scores: goodScores()}, out, errOut)

	if code := Run(context.Background(), cfg, deps); code != ExitOK {
		t.Fatalf("exit = %d, stderr=%s", code, errOut)
	}
	if !strings.HasPrefix(out.String(), "{\n") {
		t.Errorf("debug output leaked into stdout:\n%s", out.String())
	}
	for _, want := range []string{"taxonomy_version=min-v1", "model=test/model", "rule=floor+margin", "q="} {
		if !strings.Contains(errOut.String(), want) {
			t.Errorf("stderr must contain %q, got:\n%s", want, errOut.String())
		}
	}
}

func TestRunStopsWhenContextExpires(t *testing.T) {
	cfg, out, errOut := pipelineFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	if code := Run(ctx, cfg, testDeps(ctxProvider{}, out, errOut)); code != ExitProvider {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "deadline") {
		t.Errorf("stderr = %q, want a deadline message", errOut.String())
	}
	if out.Len() != 0 {
		t.Errorf("stdout must stay empty, got:\n%s", out.String())
	}
}

// ctxProvider behaves like the real client: an expired context fails the call.
type ctxProvider struct{}

func (ctxProvider) Score(ctx context.Context, photo Photo, cands []Candidate) (map[string]float64, error) {
	if err := ctx.Err(); err != nil {
		return nil, ProviderErrorf("deadline while calling model: %v", err)
	}
	return goodScores(), nil
}
