# genre-classifier (bước 1) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Xây một package Go và CLI nhận một ảnh, gọi model zero-shot một lượt duy nhất, in ra stdout JSON đúng hợp đồng bước 1.

**Architecture:** Pipeline năm tầng chạy tuần tự — mở ảnh, nạp taxonomy, một lượt gọi model trên hợp của hai nhóm ứng viên, `decide` renormalize trong từng nhóm rồi áp ngưỡng, cuối cùng marshal hợp đồng. Toàn bộ luật nghiệp vụ nằm trong `decide`, một hàm thuần không I/O, nên kiểm chứng được bằng bảng mà không cần network.

**Tech Stack:** Go 1.27, chỉ stdlib (`encoding/json`, `net/http`, `image/jpeg`, `image/png`, `context`, `flag`), `httptest` cho test provider, Hugging Face Inference API với `openai/clip-vit-large-patch14`.

**Spec:** [docs/superpowers/specs/2026-10-25-genre-classifier-pipeline-design.md](../specs/2026-10-25-genre-classifier-pipeline-design.md)

## Global Constraints

Giá trị lấy nguyên văn từ spec. Mọi task đều chịu ràng buộc này.

- Module path `github.com/minhdz-2005/photo-aesthetics-lab`, go directive `1.27` (spec §10).
- **Chỉ stdlib.** Không thêm dependency bên thứ ba, không tạo `go.sum`, không kéo runtime model (spec §10).
- Khóa JSON của hợp đồng đúng như docs bước 1, snake_case: `primary_genre`, `secondary_tags`, `alternatives`, `status`; phần tử alternative: `genre`, `score` (spec §8).
- `status` của bước 1 chỉ là `classified` hoặc `needs_review`. `error` **không** phải giá trị hợp đồng — nó là mã exit (spec §8).
- `other` **không bao giờ** là ứng viên gửi lên model; nó chỉ do luật gán (spec §5.2, §6).
- Không bao giờ so điểm xuyên nhóm. Mọi ngưỡng diễn đạt theo bội số của mức đều `1/N` (spec §3.1, §5.3).
- `q` ghi ra hợp đồng làm tròn 4 chữ số: `math.Round(q*10000)/10000`. Lý do thực thi: tổng float không kết thúc, golden byte-level sẽ không ổn định nếu không làm tròn.
- `go test ./...` chạy offline, không gọi network (spec §9.6). Test thật đặt sau `//go:build integration`.
- stdout **chỉ** chứa JSON hợp đồng; metadata debug ra stderr (spec §8).
- Mã exit: `0` có kết quả hợp đồng (kể cả `needs_review`), `1` đầu vào phía ta sai, `2` thế giới bên ngoài trả về thứ không dùng được (spec §8).
- Code, identifier, comment, help text: tiếng Anh. Tài liệu repo: tiếng Việt.
- Commit message: imperative sentence capitalized, body giải thích **why**, **không** thêm `Co-Authored-By` trailer.
- Không sửa `frontend/`.

## File Structure

Một package `genreclassifier`, mỗi file một tầng của pipeline (spec §10). Test dùng chung package để chạm được hàm chưa export.

| Đường dẫn | Trách nhiệm |
| --- | --- |
| `go.mod` | khai báo module, go 1.27 |
| `internal/genreclassifier/provider.go` | `InputError`/`ProviderError` (Task 1), `Provider` interface (Task 9) |
| `internal/genreclassifier/taxonomy.go` | kiểu dữ liệu taxonomy, `LoadTaxonomy`, validate, render prompt |
| `internal/genreclassifier/image.go` | `Photo`, `OpenPhoto`, guard kích thước |
| `internal/genreclassifier/decide.go` | `groupShares`, `genreCall`, `pickTags`, `pickAlternatives`, `Decide` |
| `internal/genreclassifier/contract.go` | `Result`, `Alternative`, `ResultOf`, `Bytes` |
| `internal/genreclassifier/huggingface.go` | HTTP client, retry, decode response |
| `internal/genreclassifier/pipeline.go` | `Config`, `Deps`, `Run`, exit codes, debug writer |
| `internal/genreclassifier/testdata/**` | fixture taxonomy, golden JSON |
| `internal/genreclassifier/*_test.go` | một file test cho mỗi file trên |
| `cmd/genre-classify/main.go` | cờ, env token, context timeout, `os.Exit` |
| `cmd/genre-classify/integration_test.go` | probe API thật, build tag `integration` |
| `ai/genre-classifier/taxonomy/genre-v1.json` | dữ liệu chạy thật, thuộc `taxonomy_version` |

Ba chỗ plan **đi lệch** spec, có lý do, đều được ghi lại trong task tương ứng:

1. **Task 3** — spec §9.1 liệt kê fixture "cạnh bằng 0". Go encode ảnh 0x0 không đáng tin, nên guard tách thành `checkDimensions` và test trực tiếp hàm đó. Không có file nhị phân 0x0 trong repo.
2. **Task 7** — spec §9.4 khẳng định "mọi hàm đơn điệu dương" giữ nguyên kết quả `decide`. Sai về toán: renormalize chỉ bất biến với nhân vô hướng; hàm phi tuyến đổi `q` nên có thể đổi cả `status`. Plan test đúng hai tính chất sau: scale invariance và within-group order preservation.
3. **Task 2** — thêm luật "hai ứng viên không được render ra cùng một prompt", vì provider map kết quả về id **theo text prompt**. Luật này hệ quả của wire protocol, spec chưa nêu.

---

## Task 1: Go module + nạp file taxonomy

**Files:**
- Create: `go.mod`
- Create: `internal/genreclassifier/taxonomy.go`
- Create: `internal/genreclassifier/provider.go`
- Create: `internal/genreclassifier/testdata/taxonomy/min-v1.json`
- Create: `internal/genreclassifier/taxonomy_test.go`
- Modify: `.gitignore`

**Interfaces:**
- Consumes: nothing.
- Produces: `Group` (`GroupGenre`, `GroupTag`), `Candidate{Group, ID, Label, Prompt}`, `Thresholds`, `Taxonomy{Version, PromptTemplate, FallbackGenre, Thresholds, Cands, path}`, `LoadTaxonomy(path string) (*Taxonomy, error)`, `(*Taxonomy) Candidates(Group) []Candidate`, `(*Taxonomy) Prompt(label string) string`, `InputError`/`ProviderError` + `InputErrorf`/`ProviderErrorf` + `isInput`/`isProvider`.

- [ ] **Step 1: Khởi tạo module**

```bash
cd D:/photo-aesthetics-lab
go mod init github.com/minhdz-2005/photo-aesthetics-lab
```

Expected: file `go.mod` chứa `module github.com/minhdz-2005/photo-aesthetics-lab` và `go 1.27`.

- [ ] **Step 2: Bổ sung `.gitignore` cho Go**

Nối vào cuối file `.gitignore` hiện có (không sửa các dòng frontend):

```
# Go
bin/
*.exe
*.exe~
*.out
/genre-classify
```

`*.out` phủ output của `go test -coverprofile`. `/genre-classify` chặn binary build ra gốc repo.

- [ ] **Step 3: Tạo fixture taxonomy dùng cho toàn bộ test**

`internal/genreclassifier/testdata/taxonomy/min-v1.json` — 4 genre + 2 tag, chọn để số tự kiểm chứng được (`N_genre = 4` → đều `0.25`, floor `0.375`, margin `0.125`; `N_tag = 2` → đều `0.5`):

```json
{
  "version": "min-v1",
  "prompt_template": "a photo of {label}",
  "fallback_genre": "other",
  "thresholds": {
    "genre_min_share_ratio": 1.5,
    "genre_min_margin_ratio": 0.5,
    "genre_alternatives_max": 3,
    "tag_min_share_ratio": 1.0,
    "tag_max": 3
  },
  "candidates": [
    { "group": "genre", "id": "portrait", "label": "a person as the main subject" },
    { "group": "genre", "id": "landscape", "label": "a natural landscape as the main subject" },
    { "group": "genre", "id": "street", "label": "a candid street scene" },
    { "group": "genre", "id": "food", "label": "food as the main subject" },
    { "group": "tag", "id": "outdoor", "label": "outdoors" },
    { "group": "tag", "id": "indoor", "label": "indoors" }
  ]
}
```

- [ ] **Step 4: Viết test fail cho `LoadTaxonomy`**

`internal/genreclassifier/taxonomy_test.go`:

```go
package genreclassifier

import (
	"path/filepath"
	"testing"
)

func TestLoadTaxonomyReadsValidFile(t *testing.T) {
	tax, err := LoadTaxonomy(filepath.Join("testdata", "taxonomy", "min-v1.json"))
	if err != nil {
		t.Fatalf("LoadTaxonomy: %v", err)
	}
	if tax.Version != "min-v1" {
		t.Errorf("Version = %q, want %q", tax.Version, "min-v1")
	}
	if got, want := len(tax.Candidates(GroupGenre)), 4; got != want {
		t.Errorf("genre candidates = %d, want %d", got, want)
	}
	if got, want := len(tax.Candidates(GroupTag)), 2; got != want {
		t.Errorf("tag candidates = %d, want %d", got, want)
	}
	if got, want := tax.Thresholds.GenreMinShareRatio, 1.5; got != want {
		t.Errorf("GenreMinShareRatio = %v, want %v", got, want)
	}
	if got, want := tax.FallbackGenre, "other"; got != want {
		t.Errorf("FallbackGenre = %q, want %q", got, want)
	}
	if got, want := tax.Prompt("outdoors"), "a photo of outdoors"; got != want {
		t.Errorf("Prompt = %q, want %q", got, want)
	}
}
```

- [ ] **Step 5: Chạy test, xác nhận fail vì chưa có định nghĩa**

```bash
go test ./internal/genreclassifier/ -run TestLoadTaxonomyReadsValidFile -v
```

Expected: compile error `undefined: LoadTaxonomy`, `undefined: GroupGenre`.

- [ ] **Step 6: Viết `provider.go` — chỉ hai kiểu lỗi**

Hai kiểu lỗi định nghĩa sớm vì mọi tầng sau đều cần để phân loại mã exit.

```go
package genreclassifier

import (
	"errors"
	"fmt"
)

// InputError: dữ liệu phía ta sai (ảnh, taxonomy, cờ, token thiếu). Exit 1.
type InputError struct{ msg string }

func (e *InputError) Error() string { return e.msg }

// ProviderError: thế giới bên ngoài trả về thứ không dùng được. Exit 2.
type ProviderError struct{ msg string }

func (e *ProviderError) Error() string { return e.msg }

// InputErrorf and ProviderErrorf take %v, never %w: the concrete error is
// rendered into the message, and the wrapper carries its exit-code class.
func InputErrorf(format string, a ...any) error { return &InputError{fmt.Sprintf(format, a...)} }

func ProviderErrorf(format string, a ...any) error { return &ProviderError{fmt.Sprintf(format, a...)} }

func isInput(err error) bool {
	var target *InputError
	return errors.As(err, &target)
}

func isProvider(err error) bool {
	var target *ProviderError
	return errors.As(err, &target)
}
```

- [ ] **Step 7: Viết `taxonomy.go`**

```go
package genreclassifier

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Group string

const (
	GroupGenre Group = "genre"
	GroupTag   Group = "tag"
)

// Candidate is one label the model may choose from. Prompt is filled by
// LoadTaxonomy; the provider maps model results back to ID by Prompt text,
// so two candidates must never render to the same Prompt.
type Candidate struct {
	Group Group  `json:"group"`
	ID    string `json:"id"`
	Label string `json:"label"`
	Prompt string `json:"-"`
}

type Thresholds struct {
	GenreMinShareRatio   float64 `json:"genre_min_share_ratio"`
	GenreMinMarginRatio  float64 `json:"genre_min_margin_ratio"`
	GenreAlternativesMax int     `json:"genre_alternatives_max"`
	TagMinShareRatio     float64 `json:"tag_min_share_ratio"`
	TagMax               int     `json:"tag_max"`
}

type Taxonomy struct {
	Version        string      `json:"version"`
	PromptTemplate string      `json:"prompt_template"`
	FallbackGenre  string      `json:"fallback_genre"`
	Thresholds     Thresholds  `json:"thresholds"`
	Cands          []Candidate `json:"candidates"`

	path string
}

// LoadTaxonomy reads and renders candidates. Validation lands in Task 2.
func LoadTaxonomy(path string) (*Taxonomy, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, InputErrorf("taxonomy file %s: %v", abs, err)
	}
	var tax Taxonomy
	if err := json.Unmarshal(raw, &tax); err != nil {
		return nil, InputErrorf("taxonomy file %s: %v", abs, err)
	}
	tax.path = abs
	for i := range tax.Cands {
		tax.Cands[i].Prompt = tax.Prompt(tax.Cands[i].Label)
	}
	return &tax, nil
}

func (t *Taxonomy) Candidates(g Group) []Candidate {
	out := []Candidate{}
	for _, c := range t.Cands {
		if c.Group == g {
			out = append(out, c)
		}
	}
	return out
}

func (t *Taxonomy) Prompt(label string) string {
	return strings.ReplaceAll(t.PromptTemplate, "{label}", label)
}
```

- [ ] **Step 8: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -run TestLoadTaxonomy -v
go vet ./...
```

Expected: `PASS`, `go vet` im lặng.

- [ ] **Step 9: Commit**

```bash
git add go.mod .gitignore internal/genreclassifier
git commit -m "Load the step 1 genre taxonomy from disk

Taxonomy wording and thresholds move into a runtime JSON file so revising a
label does not require a rebuild. That is the condition for the threshold
tuning loop in ai/evaluation, and it keeps the 8-label hypothesis out of
Go source. Prompt is rendered at load time because the provider maps model
results back to ids by prompt text."
```

---

## Task 2: Chặn taxonomy hỏng ngay lúc nạp

**Files:**
- Create: `internal/genreclassifier/taxonomy_validation_test.go`
- Create: 6 fixture trong `internal/genreclassifier/testdata/taxonomy/`
- Modify: `internal/genreclassifier/taxonomy.go` (thêm `validate()`, gọi trong `LoadTaxonomy`)

**Interfaces:**
- Consumes: `LoadTaxonomy`, `Taxonomy.Candidates`, `Taxonomy.Prompt`, `InputErrorf` (Task 1).
- Produces: `(*Taxonomy) validate() error` — gọi từ `LoadTaxonomy`, không gọi trực tiếp từ test khác.

- [ ] **Step 1: Tạo 6 fixture hỏng**

Mỗi file là một biến thể tối thiểu của `min-v1.json`, chỉ hỏng **một** luật.

`testdata/taxonomy/bad-no-placeholder.json`:

```json
{
  "version": "bad",
  "prompt_template": "a photo",
  "fallback_genre": "other",
  "thresholds": { "genre_min_share_ratio": 1.5, "genre_min_margin_ratio": 0.5, "genre_alternatives_max": 3, "tag_min_share_ratio": 1.0, "tag_max": 3 },
  "candidates": [
    { "group": "genre", "id": "portrait", "label": "a person" },
    { "group": "genre", "id": "landscape", "label": "a landscape" }
  ]
}
```

`testdata/taxonomy/bad-single-genre.json` — giống trên nhưng `prompt_template` là `"a photo of {label}"` và `candidates` chỉ còn `{ "group": "genre", "id": "portrait", "label": "a person" }`.

`testdata/taxonomy/bad-fallback-in-candidates.json` — `prompt_template` hợp lệ, hai genre candidate `portrait` + `other`.

`testdata/taxonomy/bad-duplicate-id.json` — hai genre candidate cùng `id` `"portrait"`, label khác nhau.

`testdata/taxonomy/bad-duplicate-prompt.json` — `portrait` label `"a person"` và `street` label `"a person"`, id khác nhau.

`testdata/taxonomy/bad-negative-threshold.json` — mọi thứ hợp lệ, `"genre_min_share_ratio": -1.5`.

- [ ] **Step 2: Viết test fail**

`internal/genreclassifier/taxonomy_validation_test.go`:

```go
package genreclassifier

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTaxonomyRejectsInvalidFiles(t *testing.T) {
	cases := []struct {
		file    string
		wantMsg string
	}{
		{"bad-no-placeholder.json", "prompt_template must contain {label}"},
		{"bad-single-genre.json", "at least 2 genre candidates"},
		{"bad-fallback-in-candidates.json", "must not be a candidate"},
		{"bad-duplicate-id.json", "duplicate id"},
		{"bad-duplicate-prompt.json", "same prompt"},
		{"bad-negative-threshold.json", "must be >= 0"},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			_, err := LoadTaxonomy(filepath.Join("testdata", "taxonomy", c.file))
			if err == nil {
				t.Fatalf("LoadTaxonomy(%s) = nil error, want failure", c.file)
			}
			if !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), c.wantMsg)
			}
			if !isInput(err) {
				t.Errorf("error %v is not an InputError (exit code would be wrong)", err)
			}
		})
	}
}

func TestLoadTaxonomyAcceptsValidFile(t *testing.T) {
	if _, err := LoadTaxonomy(filepath.Join("testdata", "taxonomy", "min-v1.json")); err != nil {
		t.Fatalf("min-v1.json must load: %v", err)
	}
}
```

- [ ] **Step 3: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run TestLoadTaxonomyRejects -v
```

Expected: mọi subtest FAIL với `= nil error, want failure` (Task 1 chưa validate).

- [ ] **Step 4: Thêm `validate()` vào `taxonomy.go`**

```go
func (t *Taxonomy) validate() error {
	switch {
	case t.Version == "":
		return InputErrorf("taxonomy %s: version is empty", t.path)
	case t.PromptTemplate == "":
		return InputErrorf("taxonomy %s: prompt_template is empty", t.path)
	case !strings.Contains(t.PromptTemplate, "{label}"):
		return InputErrorf("taxonomy %s: prompt_template must contain {label}", t.path)
	case t.FallbackGenre == "":
		return InputErrorf("taxonomy %s: fallback_genre is empty", t.path)
	}
	if len(t.Candidates(GroupGenre)) < 2 {
		return InputErrorf("taxonomy %s: need at least 2 genre candidates, got %d",
			t.path, len(t.Candidates(GroupGenre)))
	}

	seenID := map[string]bool{}
	seenPrompt := map[string]string{}
	for _, c := range t.Cands {
		if c.ID == "" || c.Label == "" {
			return InputErrorf("taxonomy %s: candidate %q has empty id or label", t.path, c.ID)
		}
		if c.Group != GroupGenre && c.Group != GroupTag {
			return InputErrorf("taxonomy %s: candidate %q has unknown group %q", t.path, c.ID, c.Group)
		}
		if c.ID == t.FallbackGenre {
			return InputErrorf("taxonomy %s: fallback_genre %q must not be a candidate", t.path, c.ID)
		}
		if seenID[c.ID] {
			return InputErrorf("taxonomy %s: duplicate id %q", t.path, c.ID)
		}
		seenID[c.ID] = true
		if other, dup := seenPrompt[c.Prompt]; dup {
			return InputErrorf("taxonomy %s: candidates %q and %q render to the same prompt %q",
				t.path, other, c.ID, c.Prompt)
		}
		seenPrompt[c.Prompt] = c.ID
	}

	th := t.Thresholds
	if th.GenreMinShareRatio < 0 || th.GenreMinMarginRatio < 0 || th.TagMinShareRatio < 0 {
		return InputErrorf("taxonomy %s: share and margin ratios must be >= 0", t.path)
	}
	if th.GenreAlternativesMax <= 0 {
		return InputErrorf("taxonomy %s: genre_alternatives_max must be > 0, got %d", t.path, th.GenreAlternativesMax)
	}
	if th.TagMax <= 0 {
		return InputErrorf("taxonomy %s: tag_max must be > 0, got %d", t.path, th.TagMax)
	}
	return nil
}
```

`seenID` và `seenPrompt` là map **global**, không theo group: điểm số trả về từ provider map theo id trên hợp cả hai nhóm, nên id phải duy nhất toàn cục, và provider map ngược theo prompt text nên prompt cũng phải duy nhất toàn cục.

- [ ] **Step 5: Gọi `validate()` trong `LoadTaxonomy`**

Sửa cuối hàm `LoadTaxonomy`, đặt ngay sau vòng lặp render `Prompt` (validate cần `c.Prompt` đã điền):

```go
	if err := tax.validate(); err != nil {
		return nil, err
	}
	return &tax, nil
```

- [ ] **Step 6: Chạy toàn bộ test package**

```bash
go test ./internal/genreclassifier/ -v
```

Expected: `TestLoadTaxonomyRejectsInvalidFiles` (6 subtest) PASS, `TestLoadTaxonomyAcceptsValidFile` PASS, `TestLoadTaxonomyReadsValidFile` PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/genreclassifier
git commit -m "Reject broken taxonomy files before they reach the model

A taxonomy that renders two candidates to the same prompt makes the provider
map results back to the wrong id, and a taxonomy with one genre candidate
cannot support a margin threshold. Both fail silently at score time, so they
are caught at load instead."
```

---

## Task 3: Mở ảnh và kiểm tra ảnh hợp lệ

**Files:**
- Create: `internal/genreclassifier/image.go`
- Create: `internal/genreclassifier/image_test.go`

**Interfaces:**
- Consumes: `InputErrorf`, `isInput` (Task 1).
- Produces: `Photo{Path, Bytes, MIME, Width, Height}`, `OpenPhoto(path string) (Photo, error)`, `checkDimensions(path string, width, height int) error`.

- [ ] **Step 1: Viết test fail**

`internal/genreclassifier/image_test.go`. Ảnh hợp lệ **sinh ra trong test** bằng `image/jpeg` thay vì commit file nhị phân: dễ review, và diff không chứa binary.

```go
package genreclassifier

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeJPEG(t *testing.T, width, height int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 120, A: 255})
		}
	}
	path := filepath.Join(t.TempDir(), "photo.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOpenPhotoReadsJPEG(t *testing.T) {
	path := writeJPEG(t, 40, 30)
	photo, err := OpenPhoto(path)
	if err != nil {
		t.Fatalf("OpenPhoto: %v", err)
	}
	if photo.Width != 40 || photo.Height != 30 {
		t.Errorf("size = %dx%d, want 40x30", photo.Width, photo.Height)
	}
	if photo.MIME != "image/jpeg" {
		t.Errorf("MIME = %q, want image/jpeg", photo.MIME)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(photo.Bytes, raw) {
		t.Error("photo.Bytes must be the untouched file content")
	}
}

func TestOpenPhotoRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	textNamedJPG := filepath.Join(dir, "not-a-photo.jpg")
	if err := os.WriteFile(textNamedJPG, []byte("this is plain text, not an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(dir, "empty.jpg")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		path    string
		wantMsg string
	}{
		{"missing file", filepath.Join(dir, "nope.jpg"), "photo"},
		{"empty file", empty, "not a decodable JPEG or PNG"},
		{"text renamed to jpg", textNamedJPG, "not a decodable JPEG or PNG"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := OpenPhoto(c.path)
			if err == nil {
				t.Fatalf("OpenPhoto(%s) = nil error, want failure", c.path)
			}
			if !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), c.wantMsg)
			}
			if !isInput(err) {
				t.Errorf("error %v must be InputError so exit code is 1", err)
			}
		})
	}
}

func TestCheckDimensions(t *testing.T) {
	if err := checkDimensions("a.jpg", 0, 12); err == nil {
		t.Error("width 0 must be rejected")
	}
	if err := checkDimensions("a.jpg", 12, 0); err == nil {
		t.Error("height 0 must be rejected")
	}
	if err := checkDimensions("a.jpg", 12, 8); err != nil {
		t.Errorf("12x8 must pass, got %v", err)
	}
}
```

- [ ] **Step 2: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run 'TestOpenPhoto|TestCheckDimensions' -v
```

Expected: compile error `undefined: OpenPhoto`.

- [ ] **Step 3: Viết `image.go`**

```go
package genreclassifier

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

// Photo carries the untouched file bytes: the model resizes internally, and
// re-encoding client-side would only add a quality loss nobody asked for.
type Photo struct {
	Path   string
	Bytes  []byte
	MIME   string
	Width  int
	Height int
}

func OpenPhoto(path string) (Photo, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		abs, aerr := filepath.Abs(path)
		if aerr != nil {
			abs = path
		}
		return Photo{}, InputErrorf("photo %s: %v", abs, err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return Photo{}, InputErrorf("photo %s: not a decodable JPEG or PNG (%v)", path, err)
	}
	if err := checkDimensions(path, cfg.Width, cfg.Height); err != nil {
		return Photo{}, err
	}
	return Photo{
		Path:   path,
		Bytes:  raw,
		MIME:   mimeFor(format),
		Width:  cfg.Width,
		Height: cfg.Height,
	}, nil
}

func checkDimensions(path string, width, height int) error {
	if width <= 0 || height <= 0 {
		return InputErrorf("photo %s: zero-sized image %dx%d", path, width, height)
	}
	return nil
}

// mimeFor only ever sees the formats registered above, so a non-empty format
// is always jpeg or png. DecodeConfig rejects anything else.
func mimeFor(format string) string {
	if format == "png" {
		return "image/png"
	}
	return "image/jpeg"
}
```

- [ ] **Step 4: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -run 'TestOpenPhoto|TestCheckDimensions' -v
go vet ./...
```

Expected: 5 test PASS. WebP không có decoder trong stdlib nên rơi vào nhánh "not a decodable JPEG or PNG" — đúng như spec §4.

- [ ] **Step 5: Commit**

```bash
git add internal/genreclassifier
git commit -m "Validate a photo before sending it to the model

Decoding the config catches text files renamed to .jpg and unreadable images
before a network call is spent on them. The zero-size guard is a separate
function because Go cannot reliably encode a 0x0 image, so the rule is tested
directly rather than through a committed fixture."
```

---

## Task 4: `groupShares` — sắp xếp và renormalize trong một nhóm

**Files:**
- Create: `internal/genreclassifier/decide.go`
- Create: `internal/genreclassifier/decide_test.go`

**Interfaces:**
- Consumes: `Candidate`, `ProviderErrorf`, `isProvider` (Task 1).
- Produces: `Share{Candidate, Raw, Q}`, `groupShares(scores map[string]float64, cands []Candidate) ([]Share, error)`.

Đây là chỗ quyết định hợp nhất hai lượt gọi (spec §3.1) được hiện thực hoá. Sai một ly là toàn bộ ngưỡng lệch theo.

Hai chỗ của task này đã phải sửa so với bản plan ban đầu, vì ràng buộc của IEEE-754: `q` kiểm chứng bằng `==` nên điểm gốc phải chính xác trong binary (dùng `3` và `1`, không dùng `0.3` và `0.1` — `0.3/0.4` ra `0.7499999999999999`), và `+Inf` phải là `math.Inf(1)` chứ không phải `1e308 * 10`, vì hằng untyped overflow thì biên dịch fail chứ không thành `+Inf`.

- [ ] **Step 1: Viết test fail**

`internal/genreclassifier/decide_test.go`:

```go
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
```

- [ ] **Step 2: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run TestGroupShares -v
```

Expected: compile error `undefined: groupShares`.

- [ ] **Step 3: Viết phần đầu `decide.go`**

```go
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
```

- [ ] **Step 4: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -run TestGroupShares -v
go test ./internal/genreclassifier/
```

Expected: 5 test PASS, package vẫn xanh.

Chạy thêm để chắc tie-break không phụ thuộc thứ tự duyệt map (Go randomize map iteration):

```bash
go test ./internal/genreclassifier/ -run TestGroupSharesKeepsDeclarationOrderOnTies -count=20
```

Expected: 20 lần đều PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/genreclassifier
git commit -m "Renormalize model scores within a candidate group

One merged call scores genre and tag candidates together, so the shared
softmax denominator has to cancel before any threshold means anything.
Comparing a genre score against a tag score would be meaningless. A zero
group total is an error rather than a division producing NaN."
```

---

## Task 5: `genreCall` — ngưỡng ứng viên và ngưỡng margin

**Files:**
- Modify: `internal/genreclassifier/decide.go` (thêm `GenreCall`, `genreCall`, hằng số status)
- Modify: `internal/genreclassifier/decide_test.go` (thêm bảng test)

**Interfaces:**
- Consumes: `Share`, `Thresholds` (Task 1, 4).
- Produces: `StatusClassified`, `StatusNeedsReview`, `GenreCall{Primary, Status, Rule}`, `genreCall(shares []Share, th Thresholds, fallback string) GenreCall`.

`shares` **bắt buộc** đã sắp theo `Q` giảm dần (đầu ra của `groupShares`) và có ít nhất một phần tử — `validate()` (Task 2) đảm bảo ≥ 2 genre candidate, nên không cần chặn `len(shares) == 0`.

- [ ] **Step 1: Viết test fail**

Nối vào cuối `decide_test.go`. Bảng lấy đúng các ca ở spec §9.3, dùng `q` tự đặt để số kiểm chứng tay được:

```go
func shares(q ...float64) []Share {
	out := make([]Share, 0, len(q))
	for i, v := range q {
		out = append(out, Share{
			Candidate: Candidate{Group: GroupGenre, ID: string(rune('a' + i))},
			Raw:       v,
			Q:         v,
		})
	}
	return out
}

var minV1Thresholds = Thresholds{
	GenreMinShareRatio: 1.5, GenreMinMarginRatio: 0.5,
	GenreAlternativesMax: 3, TagMinShareRatio: 1.0, TagMax: 3,
}

func TestGenreCall(t *testing.T) {
	cases := []struct {
		name        string
		q           []float64
		wantPrimary string
		wantStatus  string
		wantRule    string
	}{
		{
			name:        "qua floor va margin",
			q:           []float64{0.55, 0.20, 0.15, 0.10},
			wantPrimary: "a", wantStatus: StatusClassified, wantRule: "floor+margin",
		},
		{
			// N=4 deu 0.25, floor 0.375: 0.28 < 0.375 → khong xet margin.
			name:        "truot floor",
			q:           []float64{0.28, 0.26, 0.24, 0.22},
			wantPrimary: "other", wantStatus: StatusNeedsReview, wantRule: "floor",
		},
		{
			// 0.40 qua floor 0.375, nhung cach 0.05 < margin 0.125.
			name:        "truot margin",
			q:           []float64{0.40, 0.35, 0.15, 0.10},
			wantPrimary: "a", wantStatus: StatusNeedsReview, wantRule: "margin",
		},
		{
			name:        "hoa nhat",
			q:           []float64{0.5, 0.5, 0, 0},
			wantPrimary: "a", wantStatus: StatusNeedsReview, wantRule: "margin",
		},
		{
			// N=1 → deu 1.0, floor 1.5 → khong the nao vuot. Margin bo qua vi khong doi thu.
			name:        "mot ung vien thi floor van chan",
			q:           []float64{1},
			wantPrimary: "other", wantStatus: StatusNeedsReview, wantRule: "floor",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := genreCall(shares(c.q...), minV1Thresholds, "other")
			if got.Primary != c.wantPrimary {
				t.Errorf("Primary = %q, want %q", got.Primary, c.wantPrimary)
			}
			if got.Status != c.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, c.wantStatus)
			}
			if got.Rule != c.wantRule {
				t.Errorf("Rule = %q, want %q", got.Rule, c.wantRule)
			}
		})
	}
}

func TestGenreCallSingleCandidateSkipsMargin(t *testing.T) {
	// voi N=1 thi q = 1.0 nen floor cua minV1 (1.5x) khong the nao vuot;
	// ha floor xuong 1.0 de duong "bo qua margin" chay thay vi thanh code chet.
	th := minV1Thresholds
	th.GenreMinShareRatio = 1.0
	got := genreCall(shares(1), th, "other")
	if got.Primary != "a" || got.Status != StatusClassified || got.Rule != "single candidate" {
		t.Errorf("got %+v, want a / classified / single candidate", got)
	}
}

func TestGenreCallUsesThresholdsFromTaxonomy(t *testing.T) {
	q := []float64{0.40, 0.35, 0.15, 0.10}
	if got := genreCall(shares(q...), minV1Thresholds, "other"); got.Status != StatusNeedsReview {
		t.Fatalf("baseline must be needs_review, got %q", got.Status)
	}
	loose := minV1Thresholds
	loose.GenreMinMarginRatio = 0.1
	if got := genreCall(shares(q...), loose, "other"); got.Status != StatusClassified {
		t.Errorf("lowering the margin ratio must promote the call, got %q", got.Status)
	}
}
```

- [ ] **Step 2: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run TestGenreCall -v
```

Expected: compile error `undefined: genreCall`.

- [ ] **Step 3: Thêm vào `decide.go`**

```go
const (
	StatusClassified   = "classified"
	StatusNeedsReview  = "needs_review"
)

// GenreCall is the verdict of the two genre thresholds. Rule records which
// threshold produced it, for --debug output and for evaluation records.
type GenreCall struct {
	Primary string
	Status  string
	Rule    string
}

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
```

Ngưỡng là **bội số của mức đều**, không phải điểm tuyệt đối — đó là cách duy nhất nó còn nghĩa sau khi renormalize và sau khi taxonomy đổi số nhãn.

- [ ] **Step 4: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -run TestGenreCall -v
```

Expected: `TestGenreCall` 5 subtest PASS, `TestGenreCallSingleCandidateSkipsMargin` PASS, `TestGenreCallUsesThresholdsFromTaxonomy` PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/genreclassifier
git commit -m "Set the genre label and status from two relative thresholds

Failing the candidate floor assigns other without consulting the margin, so
an ambiguous photo is never silently promoted. Both thresholds are multiples
of the uniform share, which survives adding or removing labels; an absolute
probability floor would have been destroyed by merging the two model calls."
```

---

## Task 6: Chọn tag phụ và nhãn chẩn đoán

**Files:**
- Modify: `internal/genreclassifier/decide.go` (thêm `Alternative`, `roundShare`, `pickTags`, `pickAlternatives`)
- Modify: `internal/genreclassifier/decide_test.go`

**Interfaces:**
- Consumes: `Share`, `Thresholds` (Task 1, 4).
- Produces: `Alternative{Genre, Score}`, `roundShare(float64) float64`, `pickTags(shares []Share, th Thresholds) []string`, `pickAlternatives(shares []Share, max int, primary string) []Alternative`.

- [ ] **Step 1: Viết test fail**

Nối vào cuối `decide_test.go`:

```go
func tagShares(pairs ...float64) []Share {
	names := []string{"outdoor", "indoor", "close_up", "night"}
	out := make([]Share, 0, len(pairs))
	for i, q := range pairs {
		out = append(out, Share{
			Candidate: Candidate{Group: GroupTag, ID: names[i]},
			Raw:       q, Q: q,
		})
	}
	return out
}

func TestPickTags(t *testing.T) {
	cases := []struct {
		name string
		th   Thresholds
		q    []float64
		want string
	}{
		// N=2, deu 0.5, TagMinShareRatio 1.0 → floor 0.5. Chi 0.6 vuot.
		{"nhom hai tag lay mot", minV1Thresholds, []float64{0.6, 0.4}, "outdoor"},
		// N=4, deu 0.25 → floor 0.25. 0.2 va 0.1 trot.
		{"bo tag duoi deu", minV1Thresholds, []float64{0.4, 0.3, 0.2, 0.1}, "outdoor,indoor"},
		// Floor = 0 nen ca 4 tag deu dat; dong nay do quy tac cat theo tag_max.
		{"cat theo tag_max", zeroFloorThresholds(), []float64{0.4, 0.3, 0.2, 0.1}, "outdoor,indoor,close_up"},
		{"nhom rong", minV1Thresholds, nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := pickTags(tagShares(c.q...), c.th)
			if strings.Join(got, ",") != c.want {
				t.Errorf("tags = %v, want %q", got, c.want)
			}
			if got == nil {
				t.Error("pickTags must return an empty slice, not nil: nil marshals to JSON null")
			}
		})
	}
}

func zeroFloorThresholds() Thresholds {
	th := minV1Thresholds
	th.TagMinShareRatio = 0
	return th
}

func TestPickAlternatives(t *testing.T) {
	s := shares(0.5263, 0.2105, 0.1579, 0.1053)

	normal := pickAlternatives(s, 3, "a")
	if got, want := len(normal), 3; got != want {
		t.Fatalf("len = %d, want %d", got, want)
	}
	if normal[0].Genre != "b" || normal[0].Score != 0.2105 {
		t.Errorf("first alternative = %+v, want b/0.2105", normal[0])
	}

	// Khi floor trot, primary la other → khong loai nao ca, nhan dan dau van con.
	fallback := pickAlternatives(shares(0.28, 0.26, 0.24, 0.22), 3, "other")
	if len(fallback) != 3 || fallback[0].Genre != "a" || fallback[0].Score != 0.28 {
		t.Errorf("fallback alternatives = %+v, want a/0.28 first", fallback)
	}

	if out := pickAlternatives(s, 3, "a"); out == nil {
		t.Error("must return empty slice, not nil")
	}
}

func TestRoundShare(t *testing.T) {
	if got := roundShare(0.5263157894736842); got != 0.5263 {
		t.Errorf("roundShare = %v, want 0.5263", got)
	}
	if got := roundShare(0.10526315789473684); got != 0.1053 {
		t.Errorf("roundShare = %v, want 0.1053", got)
	}
}
```

- [ ] **Step 2: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run 'TestPickTags|TestPickAlternatives|TestRoundShare' -v
```

Expected: compile error `undefined: pickTags`.

- [ ] **Step 3: Thêm vào `decide.go`**

```go
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
```

- [ ] **Step 4: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -run 'TestPickTags|TestPickAlternatives|TestRoundShare' -v
go test ./internal/genreclassifier/
```

Expected: PASS, package xanh.

- [ ] **Step 5: Commit**

```bash
git add internal/genreclassifier
git commit -m "Collect secondary tags and diagnostic alternatives

Tags and genres compete only inside their own group, so the tag floor is a
fraction of the tag uniform, not of the genre one. Both lists start as empty
slices because a nil marshals to JSON null and the contract says array.
Shares are rounded so the reported number does not wobble with float noise."
```

---

## Task 7: `Decide` — hàm thuần nối hết luật, cộng hai phép kiểm chứng bất biến

**Files:**
- Modify: `internal/genreclassifier/decide.go` (thêm `Decision`, `Detail`, `Decide`)
- Create: `internal/genreclassifier/decide_invariants_test.go`
- Modify: `docs/superpowers/specs/2026-10-25-genre-classifier-pipeline-design.md` (sửa §9.4)

**Interfaces:**
- Consumes: `groupShares`, `genreCall`, `pickTags`, `pickAlternatives`, `Taxonomy` (Task 1–6).
- Produces: `Decision{PrimaryGenre, SecondaryTags, Alternatives, Status}`, `Detail{Genre, Tags []Share}`, `Decide(scores map[string]float64, tax *Taxonomy) (Decision, Detail, error)`.

- [ ] **Step 1: Viết test fail cho `Decide`**

`internal/genreclassifier/decide_invariants_test.go`:

```go
package genreclassifier

import (
	"path/filepath"
	"reflect"
	"testing"
)

func loadMinV1(t *testing.T) *Taxonomy {
	t.Helper()
	tax, err := LoadTaxonomy(filepath.Join("testdata", "taxonomy", "min-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return tax
}

func TestDecideClassifiesClearPortrait(t *testing.T) {
	// Genre tong 0.95, tag tong 0.05; hop = 1.0 nhu softmax tren union.
	scores := map[string]float64{
		"portrait": 0.50, "landscape": 0.20, "street": 0.15, "food": 0.10,
		"outdoor": 0.03, "indoor": 0.02,
	}
	dec, detail, err := Decide(scores, loadMinV1(t))
	if err != nil {
		t.Fatal(err)
	}
	if dec.PrimaryGenre != "portrait" || dec.Status != StatusClassified {
		t.Errorf("got %s/%s, want portrait/classified", dec.PrimaryGenre, dec.Status)
	}
	if got, want := dec.SecondaryTags, []string{"outdoor"}; !reflect.DeepEqual(got, want) {
		t.Errorf("tags = %v, want %v", got, want)
	}
	// q = p / 0.95: 0.2105, 0.1579, 0.1053
	wantAlt := []Alternative{
		{Genre: "landscape", Score: 0.2105},
		{Genre: "street", Score: 0.1579},
		{Genre: "food", Score: 0.1053},
	}
	if !reflect.DeepEqual(dec.Alternatives, wantAlt) {
		t.Errorf("alternatives = %+v, want %+v", dec.Alternatives, wantAlt)
	}
	if len(detail.Genre) != 4 || len(detail.Tags) != 2 {
		t.Errorf("detail = %d genre, %d tag; want 4 and 2", len(detail.Genre), len(detail.Tags))
	}
}

func TestDecideAssignsOtherWhenFloorFails(t *testing.T) {
	scores := map[string]float64{
		"portrait": 0.24, "landscape": 0.23, "street": 0.22, "food": 0.21,
		"outdoor": 0.05, "indoor": 0.05,
	}
	dec, _, err := Decide(scores, loadMinV1(t))
	if err != nil {
		t.Fatal(err)
	}
	if dec.PrimaryGenre != "other" {
		t.Errorf("primary = %q, want other", dec.PrimaryGenre)
	}
	if dec.Status != StatusNeedsReview {
		t.Errorf("status = %q, want needs_review", dec.Status)
	}
	if len(dec.Alternatives) == 0 || dec.Alternatives[0].Genre != "portrait" {
		t.Errorf("diagnostics must survive a failed floor, got %+v", dec.Alternatives)
	}
	for _, a := range dec.Alternatives {
		if a.Genre == "other" {
			t.Error("other is assigned by rule, never a scored alternative")
		}
	}
}

func TestDecideRejectsMissingScore(t *testing.T) {
	scores := map[string]float64{
		"portrait": 0.5, "landscape": 0.2, "street": 0.15,
		"outdoor": 0.03, "indoor": 0.02, // food thieu
	}
	if _, _, err := Decide(scores, loadMinV1(t)); err == nil {
		t.Fatal("a missing candidate score must be an error")
	}
}

func TestDecideWithEmptyTagGroup(t *testing.T) {
	tax := loadMinV1(t)
	genreOnly := []Candidate{}
	for _, c := range tax.Cands {
		if c.Group == GroupGenre {
			genreOnly = append(genreOnly, c)
		}
	}
	tax.Cands = genreOnly

	dec, _, err := Decide(map[string]float64{
		"portrait": 0.5, "landscape": 0.2, "street": 0.15, "food": 0.1,
	}, tax)
	if err != nil {
		t.Fatal(err)
	}
	if dec.PrimaryGenre != "portrait" {
		t.Errorf("primary = %q, want portrait", dec.PrimaryGenre)
	}
	if dec.SecondaryTags == nil || len(dec.SecondaryTags) != 0 {
		t.Errorf("tags = %v, want empty non-nil slice", dec.SecondaryTags)
	}
}
```

- [ ] **Step 2: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run TestDecide -v
```

Expected: compile error `undefined: Decide`.

- [ ] **Step 3: Thêm `Decision`, `Detail`, `Decide` vào `decide.go`**

```go
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
```

- [ ] **Step 4: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -v
```

Expected: toàn bộ package PASS, gồm cả các test ở Task 1–6.

- [ ] **Step 5: Thêm hai phép kiểm chứng bất biến**

Spec §9.4 đòi hỏi hai thứ, nhưng một trong hai phát biểu là **sai về toán**. Renormalize chỉ bất biến với phép **nhân vô hướng**; một hàm đơn điệu phi tuyến (ví dụ `x²`) đổi tỷ lệ giữa các `q`, nên nó đổi cả `status` một cách chính đáng. Plan chỉ test những gì đúng:

```go
func TestDecideIsScaleInvariant(t *testing.T) {
	tax := loadMinV1(t)
	base := map[string]float64{
		"portrait": 0.50, "landscape": 0.20, "street": 0.15, "food": 0.10,
		"outdoor": 0.03, "indoor": 0.02,
	}
	want, _, err := Decide(base, tax)
	if err != nil {
		t.Fatal(err)
	}
	for _, factor := range []float64{2, 0.125, 1000} {
		scaled := make(map[string]float64, len(base))
		for id, p := range base {
			scaled[id] = p * factor
		}
		got, _, err := Decide(scaled, tax)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("factor %v changed the decision: %+v vs %+v", factor, got, want)
		}
	}
}

func TestDecideGenreOutputIgnoresTagScores(t *testing.T) {
	tax := loadMinV1(t)
	a, _, err := Decide(map[string]float64{
		"portrait": 0.5, "landscape": 0.2, "street": 0.15, "food": 0.1,
		"outdoor": 0.049, "indoor": 0.001,
	}, tax)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := Decide(map[string]float64{
		"portrait": 0.5, "landscape": 0.2, "street": 0.15, "food": 0.1,
		"outdoor": 0.001, "indoor": 0.049,
	}, tax)
	if err != nil {
		t.Fatal(err)
	}
	if a.PrimaryGenre != b.PrimaryGenre || a.Status != b.Status {
		t.Errorf("genre verdict changed with tag scores: %s/%s vs %s/%s",
			a.PrimaryGenre, a.Status, b.PrimaryGenre, b.Status)
	}
	if !reflect.DeepEqual(a.Alternatives, b.Alternatives) {
		t.Errorf("alternatives moved across groups: %+v vs %+v", a.Alternatives, b.Alternatives)
	}
}
```

Ca `ScaleInvariant` là lưới an toàn duy nhất bắt được việc **vô tình dùng điểm tuyệt đối** còn sót lại sau khi hợp nhất lượt gọi. Ca `IgnoresTagScores` kiểm chứng câu "không so điểm xuyên nhóm" mà không cần đọc code.

- [ ] **Step 6: Chạy tiếp, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -run 'TestDecideIsScaleInvariant|TestDecideGenreOutputIgnoresTagScores' -v -count=5
```

Expected: PASS cả 5 lần (`Decide` phải tất định dù map duyệt ngẫu nhiên).

- [ ] **Step 7: Sửa lại phát biểu sai trong spec**

Trong `docs/superpowers/specs/2026-10-25-genre-classifier-pipeline-design.md`, mục §9.4, thay bullet đầu tiên (bắt đầu bằng "**Bất biến theo tỉ lệ:**") bằng:

```markdown
- **Bất biến theo tỉ lệ:** nhân **toàn bộ** điểm thô với cùng một thừa số `c > 0` thì kết quả `decide` phải không đổi — đây là phát biểu đúng của tính bất biến. Một hàm đơn điệu *phi tuyến* thì **không** giữ kết quả: nó đổi tỷ lệ giữa các `q` nên có thể đổi cả `status`. Điều duy nhất mọi hàm đơn điệu đảm bảo là thứ tự trong nhóm không đổi.
```

- [ ] **Step 8: Commit**

```bash
git add internal/genreclassifier docs/superpowers/specs
git commit -m "Assemble the genre decision from per-group scores

Decide is the only place step 1 rules live: no I/O, no network, no provider
knowledge, so the whole contract can be driven from a table. The spec claimed
every monotone transform preserves the outcome, which is false for nonlinear
ones; the scale-invariance test states the property that actually holds."
```

---

## Task 8: Hợp đồng JSON và golden test

**Files:**
- Create: `internal/genreclassifier/contract.go`
- Create: `internal/genreclassifier/contract_test.go`
- Create: `internal/genreclassifier/testdata/golden/portrait-classified.json`

**Interfaces:**
- Consumes: `Decision`, `Alternative` (Task 6, 7).
- Produces: `Result`, `ResultOf(Decision) Result`, `(Result) Bytes() ([]byte, error)`.

- [ ] **Step 1: Viết golden file**

`internal/genreclassifier/testdata/golden/portrait-classified.json` — đúng định dạng `json.MarshalIndent(r, "", "  ")` cộng một ký tự xuống dòng cuối:

```json
{
  "primary_genre": "portrait",
  "secondary_tags": [
    "outdoor"
  ],
  "alternatives": [
    {
      "genre": "landscape",
      "score": 0.2105
    },
    {
      "genre": "street",
      "score": 0.1579
    },
    {
      "genre": "food",
      "score": 0.1053
    }
  ],
  "status": "classified"
}
```

- [ ] **Step 2: Viết test fail**

`internal/genreclassifier/contract_test.go`:

```go
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
	// Hop dong yeu cau mot danh sach, khong phai null.
	if strings.Contains(out, "null") {
		t.Errorf("empty lists must serialize as [] not null:\n%s", out)
	}
	if !strings.HasSuffix(out, "}\n") {
		t.Error("contract output must end with exactly one newline")
	}
}
```

- [ ] **Step 3: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run TestResult -v
```

Expected: compile error `undefined: ResultOf`.

- [ ] **Step 4: Viết `contract.go`**

```go
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
```

Không có field `confidence`, `taxonomy_version` hay `error` — hợp đồng bước 1 không có chỗ cho chúng (spec §5.5, §8).

- [ ] **Step 5: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -run TestResult -v
```

Expected: `TestResultMatchesGolden` PASS (byte bằng nhau). Nếu fail vì whitespace, chạy `go test ./internal/genreclassifier/ -run TestResultMatchesGolden -update-golden` rồi **đọc diff** để chắc không có key nào lọt vào hợp đồng — đó là lý do flag tồn tại, không phải để chấp nhận mọi khác biệt.

- [ ] **Step 6: Commit**

```bash
git add internal/genreclassifier
git commit -m "Pin the step 1 contract to a golden JSON file

Comparing bytes instead of fields catches an extra key, a missing key and the
nil-slice-to-null trap in one shot. taxonomy_version and error are deliberately
absent: the step 1 contract has no place for them, so they go to debug output
and exit codes instead."
```

---

## Task 9: `Provider` interface và một lượt gọi Hugging Face

**Files:**
- Create: `internal/genreclassifier/huggingface.go`
- Create: `internal/genreclassifier/huggingface_test.go`
- Modify: `internal/genreclassifier/provider.go` (thêm `Provider` interface)

**Interfaces:**
- Consumes: `Photo` (Task 3), `Candidate.Prompt` (Task 1), `ProviderErrorf` (Task 1).
- Produces: `Provider` interface, `DefaultHFBaseURL`, `DefaultModel`, `HuggingFace`, `NewHuggingFace(model, token string) (*HuggingFace, error)`, `(p *HuggingFace) Score(ctx, Photo, []Candidate) (map[string]float64, error)`.

`Score` trả điểm theo **id ứng viên**, không theo prompt — mọi tầng sau không cần biết wire format.

- [ ] **Step 1: Thêm interface vào `provider.go`**

```go
import "context" // thêm vào block import

// Provider scores every candidate in one call and keys the result by candidate
// id. It never sets thresholds and never decides other or needs_review: that
// is the caller's job (ai/model-provider README).
type Provider interface {
	Score(ctx context.Context, photo Photo, cands []Candidate) (map[string]float64, error)
}
```

- [ ] **Step 2: Viết test fail**

`internal/genreclassifier/huggingface_test.go`:

```go
package genreclassifier

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type captured struct {
	auth       string
	contentTyp string
	labels     []string
	inputs     string
}

func newProvider(t *testing.T, status int, body string, got *captured) (*HuggingFace, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var req struct {
			Inputs     string `json:"inputs"`
			Parameters struct {
				Candidates []string `json:"candidate_labels"`
			} `json:"parameters"`
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Errorf("request body is not the expected JSON: %v", err)
		}
		got.auth = r.Header.Get("Authorization")
		got.contentTyp = r.Header.Get("Content-Type")
		got.labels = req.Parameters.Candidates
		got.inputs = req.Inputs

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		io.WriteString(w, body)
	}))
	p, err := NewHuggingFace("test/model", "tok-123")
	if err != nil {
		t.Fatal(err)
	}
	p.BaseURL = srv.URL + "/models/"
	p.Backoff = 0
	p.Attempts = 1
	return p, srv
}

func testPhoto() Photo {
	return Photo{Path: "p.jpg", Bytes: []byte{0xFF, 0xD8, 0xFF, 0x00}, MIME: "image/jpeg", Width: 4, Height: 3}
}

func TestHFScoreMapsLabelsBackToIDs(t *testing.T) {
	cands := []Candidate{
		{Group: GroupGenre, ID: "portrait", Prompt: "a photo of a person"},
		{Group: GroupGenre, ID: "landscape", Prompt: "a photo of a landscape"},
		{Group: GroupTag, ID: "outdoor", Prompt: "a photo of outdoors"},
	}
	var got captured
	body := `[
		{"label":"a photo of a landscape","score":0.2},
		{"label":"a photo of a person","score":0.5},
		{"label":"a photo of outdoors","score":0.3}
	]`
	p, srv := newProvider(t, http.StatusOK, body, &got)
	defer srv.Close()

	scores, err := p.Score(context.Background(), testPhoto(), cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(scores) != 3 {
		t.Fatalf("scores = %v, want 3 entries keyed by id", scores)
	}
	if scores["portrait"] != 0.5 || scores["landscape"] != 0.2 || scores["outdoor"] != 0.3 {
		t.Errorf("scores = %v, want portrait 0.5, landscape 0.2, outdoor 0.3", scores)
	}
	if got.auth != "Bearer tok-123" {
		t.Errorf("Authorization = %q", got.auth)
	}
	if got.contentTyp != "application/json" {
		t.Errorf("Content-Type = %q", got.contentTyp)
	}
	// Thu tu gui len phai la thu tu taxonomy, khong phai thu tu response.
	if strings.Join(got.labels, "|") != "a photo of a person|a photo of a landscape|a photo of outdoors" {
		t.Errorf("candidate_labels = %v", got.labels)
	}
	want := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(testPhoto().Bytes)
	if got.inputs != want {
		t.Errorf("inputs = %q, want %q", got.inputs, want)
	}
}

func TestHFScoreRejectsUnusableResponses(t *testing.T) {
	cands := []Candidate{
		{Group: GroupGenre, ID: "portrait", Prompt: "a person"},
		{Group: GroupGenre, ID: "landscape", Prompt: "a landscape"},
	}
	cases := []struct {
		name    string
		status  int
		body    string
		wantMsg string
	}{
		{"missing candidate", http.StatusOK, `[{"label":"a person","score":0.9}]`, "of 2 candidates"},
		{"unknown label", http.StatusOK, `[{"label":"a person","score":0.5},{"label":"a cat","score":0.5}]`, "unknown label"},
		{"not an array", http.StatusOK, `{"error":"model is loading"}`, "not a [{label,score}] array"},
		{"infinite score", http.StatusOK, `[{"label":"a person","score":1e400},{"label":"a landscape","score":0.1}]`, "non-finite"},
		{"server error", http.StatusInternalServerError, `{"error":"boom"}`, "HTTP 500"},
		{"bad request", http.StatusBadRequest, `{"error":"invalid input"}`, "HTTP 400"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got captured
			p, srv := newProvider(t, c.status, c.body, &got)
			defer srv.Close()

			_, err := p.Score(context.Background(), testPhoto(), cands)
			if err == nil {
				t.Fatalf("want error for %s", c.name)
			}
			if !strings.Contains(err.Error(), c.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), c.wantMsg)
			}
			if !isProvider(err) {
				t.Errorf("error %v must be ProviderError so exit code is 2", err)
			}
		})
	}
}

func TestNewHuggingFaceRequiresToken(t *testing.T) {
	if _, err := NewHuggingFace("test/model", ""); err == nil {
		t.Fatal("empty token must fail")
	} else if !isInput(err) {
		t.Errorf("empty token is our mistake, not the world's: got %T", err)
	}
}
```

- [ ] **Step 3: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run 'TestHF|TestNewHuggingFace' -v
```

Expected: compile error `undefined: NewHuggingFace`.

- [ ] **Step 4: Viết `huggingface.go` (chưa retry)**

```go
package genreclassifier

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultHFBaseURL = "https://api-inference.huggingface.co/models/"
	DefaultModel     = "openai/clip-vit-large-patch14"

	maxResponseBytes = 1 << 20
)

// HuggingFace posts one image plus every candidate prompt and reads back one
// score per prompt. Attempts and Backoff are fields so tests can run the retry
// path without sleeping.
type HuggingFace struct {
	Model    string
	Token    string
	BaseURL  string
	HTTPC    *http.Client
	Backoff  time.Duration
	Attempts int
}

func NewHuggingFace(model, token string) (*HuggingFace, error) {
	if strings.TrimSpace(token) == "" {
		return nil, InputErrorf("HF_TOKEN is empty; export a Hugging Face API token before running")
	}
	if strings.TrimSpace(model) == "" {
		return nil, InputErrorf("model id is empty")
	}
	return &HuggingFace{
		Model:    model,
		Token:    token,
		BaseURL:  DefaultHFBaseURL,
		HTTPC:    &http.Client{},
		Backoff:  500 * time.Millisecond,
		Attempts: 1,
	}, nil
}

type hfRequest struct {
	Inputs     string       `json:"inputs"`
	Parameters hfParameters `json:"parameters"`
}

type hfParameters struct {
	Candidates []string `json:"candidate_labels"`
}

type hfResult struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

func (p *HuggingFace) Score(ctx context.Context, photo Photo, cands []Candidate) (map[string]float64, error) {
	prompts := make([]string, len(cands))
	for i, c := range cands {
		prompts[i] = c.Prompt
	}
	body, err := json.Marshal(hfRequest{
		Inputs: "data:" + photo.MIME + ";base64," + base64.StdEncoding.EncodeToString(photo.Bytes),
		Parameters: hfParameters{Candidates: prompts},
	})
	if err != nil {
		return nil, err
	}
	payload, status, _, err := p.postOnce(ctx, body)
	if err != nil {
		return nil, ProviderErrorf("request to %s failed: %v", p.Model, err)
	}
	if status != http.StatusOK {
		return nil, ProviderErrorf("HTTP %d from %s: %s", status, p.Model, snippet(payload))
	}
	return p.decode(payload, cands)
}

func (p *HuggingFace) postOnce(ctx context.Context, body []byte) ([]byte, int, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+p.Model, bytes.NewReader(body))
	if err != nil {
		return nil, 0, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := p.HTTPC.Do(req)
	if err != nil {
		return nil, 0, 0, err
	}
	defer res.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return nil, res.StatusCode, 0, err
	}
	return payload, res.StatusCode, retryAfter(res.Header.Get("Retry-After")), nil
}

// decode keys results by candidate id. A short list is fatal, not trimmed:
// per-group renormalization needs every score (spec §3.1).
func (p *HuggingFace) decode(payload []byte, cands []Candidate) (map[string]float64, error) {
	var rows []hfResult
	if err := json.Unmarshal(payload, &rows); err != nil {
		return nil, ProviderErrorf("response is not a [{label,score}] array: %v", err)
	}
	byPrompt := make(map[string]Candidate, len(cands))
	for _, c := range cands {
		byPrompt[c.Prompt] = c
	}
	scores := make(map[string]float64, len(cands))
	for _, r := range rows {
		c, ok := byPrompt[r.Label]
		if !ok {
			return nil, ProviderErrorf("model returned unknown label %q", r.Label)
		}
		if math.IsNaN(r.Score) || math.IsInf(r.Score, 0) {
			return nil, ProviderErrorf("non-finite score %v for candidate %q", r.Score, c.ID)
		}
		scores[c.ID] = r.Score
	}
	if len(scores) != len(cands) {
		missing := []string{}
		for _, c := range cands {
			if _, ok := scores[c.ID]; !ok {
				missing = append(missing, c.ID)
			}
		}
		return nil, ProviderErrorf("model returned %d of %d candidates, missing %s",
			len(scores), len(cands), strings.Join(missing, ", "))
	}
	return scores, nil
}

func retryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	secs, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

func snippet(payload []byte) string {
	const limit = 200
	s := strings.TrimSpace(string(payload))
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
```

`retryAfter` cần `import "strconv"`. Thêm `"infinite score"` vào test được phát hiện ở `decode`: `1e400` unmarshal thành `+Inf` (thành thật, không phải nội suy).

- [ ] **Step 5: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -run 'TestHF|TestNewHuggingFace' -v
go vet ./...
```

Expected: 3 test PASS (gồm 6 subtest của `TestHFScoreRejectsUnusableResponses`).

- [ ] **Step 6: Commit**

```bash
git add internal/genreclassifier
git commit -m "Score every candidate in one Hugging Face call

Genre and tag labels travel together so the merged softmax can be split by
renormalizing each group afterwards. A response that omits a candidate is an
error rather than a shortened list, because a partial group total makes every
share wrong instead of merely less confident."
```

---

## Task 10: Retry có trần thời gian

**Files:**
- Modify: `internal/genreclassifier/huggingface.go` (vòng lặp trong `Score`, `retryable`, `NewHuggingFace` đặt `Attempts: 2`)
- Modify: `internal/genreclassifier/huggingface_test.go`

**Interfaces:**
- Consumes: `postOnce`, `decode`, `retryAfter` (Task 9).
- Produces: hành vi retry của `Score`; `retryable(status int) bool`.

Retry phải nằm **trong** deadline của cả lệnh, không được kéo dài tổng thời gian (spec §7.2). Timeout → `error`, không hạ xuống `needs_review`.

- [ ] **Step 1: Viết test fail**

Nối vào cuối `huggingface_test.go`:

```go
type sequence struct {
	statuses []int
	headers  []map[string]string
	calls    int
}

func newSequenceProvider(t *testing.T, seq *sequence) (*HuggingFace, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := seq.calls
		seq.calls++
		if i >= len(seq.statuses) {
			i = len(seq.statuses) - 1
		}
		if h := seq.headers[i]; h != nil {
			for k, v := range h {
				w.Header().Set(k, v)
			}
		}
		status := seq.statuses[i]
		if status == http.StatusOK {
			io.WriteString(w, `[{"label":"a person","score":0.9},{"label":"a landscape","score":0.1}]`)
		} else {
			io.WriteString(w, `{"error":"temporary"}`)
		}
		w.WriteHeader(status)
	}))
	p, err := NewHuggingFace("test/model", "tok")
	if err != nil {
		t.Fatal(err)
	}
	p.BaseURL = srv.URL + "/models/"
	p.Backoff = time.Millisecond
	return p, srv
}

func retryCands() []Candidate {
	return []Candidate{
		{Group: GroupGenre, ID: "portrait", Prompt: "a person"},
		{Group: GroupGenre, ID: "landscape", Prompt: "a landscape"},
	}
}

func TestHFRetriesTransientFailures(t *testing.T) {
	cases := []struct {
		name      string
		statuses  []int
		wantCalls int
		wantOK    bool
	}{
		{"500 roi 200", []int{500, 200}, 2, true},
		{"429 roi 200", []int{429, 200}, 2, true},
		{"401 khong retry", []int{401, 200}, 1, false},
		{"400 khong retry", []int{400, 200}, 1, false},
		{"hai 500 van fail", []int{500, 503}, 2, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			seq := &sequence{statuses: c.statuses, headers: make([]map[string]string, len(c.statuses))}
			p, srv := newSequenceProvider(t, seq)
			defer srv.Close()

			_, err := p.Score(context.Background(), testPhoto(), retryCands())
			if seq.calls != c.wantCalls {
				t.Errorf("calls = %d, want %d", seq.calls, c.wantCalls)
			}
			if c.wantOK && err != nil {
				t.Errorf("want success, got %v", err)
			}
			if !c.wantOK && err == nil {
				t.Error("want error")
			}
			if err != nil && !isProvider(err) {
				t.Errorf("error %v must be ProviderError", err)
			}
		})
	}
}

func TestHFDoesNotRetryABrokenResponse(t *testing.T) {
	// 200 nhung thieu ứng viên: goi lai cung ra ket qua nhu vay, nên không retry.
	seq := &sequence{statuses: []int{200}, headers: make([]map[string]string, 1)}
	p, srv := newSequenceProvider(t, seq)
	defer srv.Close()

	if _, err := p.Score(context.Background(), testPhoto(), retryCands()); err == nil {
		t.Fatal("want error for a short response")
	} else if seq.calls != 1 {
		t.Errorf("calls = %d, want 1: schema failures must not be retried", seq.calls)
	}
}

func TestHFRespectsContextDeadline(t *testing.T) {
	blocked := make(chan struct{})
	defer close(blocked)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked // keu giu cho den khi test huy
	}))
	defer srv.Close()

	p, err := NewHuggingFace("test/model", "tok")
	if err != nil {
		t.Fatal(err)
	}
	p.BaseURL = srv.URL + "/models/"
	p.Backoff = time.Hour // retry sleep dai; deadline phai thang

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err = p.Score(ctx, testPhoto(), retryCands())
	if err == nil {
		t.Fatal("a hung server must produce an error")
	}
	if !isProvider(err) {
		t.Errorf("error %v must be ProviderError", err)
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("error = %q, want it to mention the deadline", err.Error())
	}
}

func TestHFRetriesAfterRetryAfterHeader(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `[{"label":"a person","score":0.9},{"label":"a landscape","score":0.1}]`)
	}))
	defer srv.Close()

	p, err := NewHuggingFace("test/model", "tok")
	if err != nil {
		t.Fatal(err)
	}
	p.BaseURL = srv.URL + "/models/"
	p.Backoff = time.Millisecond

	if _, err := p.Score(context.Background(), testPhoto(), retryCands()); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}
```

Test file cần `import "time"` (bên cạnh `context`, `encoding/base64`, `encoding/json`, `io`, `net/http`, `net/http/httptest`, `strings`, `testing`).

- [ ] **Step 2: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run 'TestHF(Retries|DoesNot|Respects)' -v
```

Expected: FAIL — `Attempts = 1` nên chỉ một lượt gọi, và chưa có vòng lặp retry.

- [ ] **Step 3: Sửa `NewHuggingFace`**

```go
		Attempts: 2,
```

(thay `Attempts: 1`) — "tối đa hai lượt gọi" theo spec §7.2.

- [ ] **Step 4: Thay thân `Score` bằng vòng lặp retry**

Giữ nguyên phần dựng `body`, chỉ đổi phần gọi HTTP:

```go
	delay := p.Backoff
	var lastErr error
	for attempt := 1; attempt <= p.Attempts; attempt++ {
		payload, status, retryWait, err := p.postOnce(ctx, body)
		switch {
		case err != nil:
			lastErr = ProviderErrorf("request to %s failed: %v", p.Model, err)
		case status == http.StatusOK:
			scores, derr := p.decode(payload, cands)
			if derr != nil {
				return nil, derr // schema hỏng thì goi lai cung hong
			}
			return scores, nil
		case !retryable(status):
			return nil, ProviderErrorf("HTTP %d from %s: %s", status, p.Model, snippet(payload))
		default:
			lastErr = ProviderErrorf("HTTP %d from %s: %s", status, p.Model, snippet(payload))
			if retryWait > 0 {
				delay = retryWait
			}
		}
		if attempt == p.Attempts {
			break
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ProviderErrorf("deadline while retrying %s: %v", p.Model, ctx.Err())
		}
		delay *= 2
	}
	return nil, lastErr
```

Và thêm helper:

```go
func retryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}
```

`ctx.Done()` chắn trong `select` nghĩa là retry không bao giờ vượt thời hạn của cả lệnh: hết giờ là `ProviderError`, không có `needs_review` giả.

- [ ] **Step 5: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -v
go vet ./...
```

Expected: toàn bộ package PASS, gồm 4 test retry mới.

- [ ] **Step 6: Commit**

```bash
git add internal/genreclassifier
git commit -m "Retry transient provider failures inside the deadline

One network blip should not cost a labelled photo, but a wrong token or a
malformed response should not be retried either, so only transport errors, 429
and 5xx get a second attempt. The sleep waits on ctx.Done as well, which keeps
the retry budget inside the command deadline instead of behind it."
```

---

## Task 11: `Run` — nối năm tầng và mã exit

**Files:**
- Create: `internal/genreclassifier/pipeline.go`
- Create: `internal/genreclassifier/pipeline_test.go`

**Interfaces:**
- Consumes: `Provider`, `Photo`, `OpenPhoto` (Task 3, 9), `LoadTaxonomy` (Task 1), `Decide`, `Detail` (Task 7), `ResultOf` (Task 8).
- Produces: `ExitOK`/`ExitInput`/`ExitProvider`, `Config{PhotoPath, TaxonomyPath, Model, Timeout, Debug}`, `Deps{Provider, Open, Stdout, Stderr}`, `Run(ctx context.Context, cfg Config, d Deps) int`.

- [ ] **Step 1: Viết test fail**

`internal/genreclassifier/pipeline_test.go`:

```go
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
	return f.scorePhoto(), nil
}

func (f *fakeProvider) scorePhoto() map[string]float64 { return f.scores }

func goodScores() map[string]float64 {
	return map[string]float64{
		"portrait": 0.50, "landscape": 0.20, "street": 0.15, "food": 0.10,
		"outdoor": 0.03, "indoor": 0.02,
	}
}

func pipelineFixture(t *testing.T, _ Provider) (Config, *bytes.Buffer, *bytes.Buffer) {
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
	cfg, out, errOut := pipelineFixture(t, nil)
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
		{"needs_review van la 0", map[string]float64{
			"portrait": 0.24, "landscape": 0.23, "street": 0.22, "food": 0.21,
			"outdoor": 0.05, "indoor": 0.05,
		}, nil, nil, ExitOK},
		{"thieu diem la 2", map[string]float64{"portrait": 0.5}, nil, nil, ExitProvider},
		{"provider loi la 2", nil, ProviderErrorf("HTTP 503"), nil, ExitProvider},
		{"anh loi la 1", nil, nil, InputErrorf("photo bad: unreadable"), ExitInput},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg, out, errOut := pipelineFixture(t, nil)
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
	cfg, out, errOut := pipelineFixture(t, nil)
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
	cfg, out, errOut := pipelineFixture(t, nil)
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
	cfg, out, errOut := pipelineFixture(t, nil)
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
```

- [ ] **Step 2: Chạy test, xác nhận fail**

```bash
go test ./internal/genreclassifier/ -run TestRun -v
```

Expected: compile error `undefined: Run`.

- [ ] **Step 3: Viết `pipeline.go`**

```go
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
	fmt.Fprintf(w, "primary=%s status=%s tags=%v\n", dec.PrimaryGenre, dec.Status, dec.SecondaryTags)
}
```

`Detail.Rule` đã được `Decide` set ở Task 7, nên `writeDebug` chỉ việc in nó ra — thay dòng cuối của `writeDebug` ở trên bằng:

```go
	fmt.Fprintf(w, "primary=%s status=%s rule=%s tags=%v\n",
		dec.PrimaryGenre, dec.Status, detail.Rule, dec.SecondaryTags)
```

- [ ] **Step 4: Chạy test, xác nhận pass**

```bash
go test ./internal/genreclassifier/ -v
go vet ./...
```

Expected: toàn bộ package PASS. `TestRunWritesContractToStdout` xác nhận đúng **một** lượt gọi provider — đó là bằng chứng decision hợp nhất ở Phần 2 không bị ai lén bẻ lại thành hai lượt.

- [ ] **Step 5: Commit**

```bash
git add internal/genreclassifier
git commit -m "Wire the step 1 pipeline and map failures to exit codes

needs_review is a normal result and exits 0, while a broken photo exits 1 and
an unusable model response exits 2, so a batch runner can tell its own bugs
from the world's. Debug metadata goes to stderr because stdout is the contract."
```

---

## Task 12: CLI `genre-classify`

**Files:**
- Create: `cmd/genre-classify/main.go`
- Create: `cmd/genre-classify/main_test.go`

**Interfaces:**
- Consumes: `genreclassifier.Config`, `Run`, `Deps`, `OpenPhoto`, `NewHuggingFace`, `DefaultModel`, exit constants (Task 9, 11).
- Produces: binary `genre-classify`; `parseArgs(args []string) (gc.Config, error)` and `execute(args []string, token string) int` (unexported, tested in-package).

- [ ] **Step 1: Viết test fail cho phần parse**

`cmd/genre-classify/main_test.go`:

```go
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
```

- [ ] **Step 2: Chạy test, xác nhận fail**

```bash
go test ./cmd/genre-classify/ -v
```

Expected: compile error `undefined: parseArgs`.

- [ ] **Step 3: Viết `main.go`**

```go
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
```

`gc.Config.Timeout` khai báo ở Task 11; `pipeline_test.go` không đặt nên nhận giá trị zero — vô hại vì `Run` không đọc field đó, chỉ `main` đọc để dựng context deadline.

- [ ] **Step 4: Chạy test và build**

```bash
go test ./cmd/genre-classify/ -v
go build ./...
go run ./cmd/genre-classify -h
```

Expected: 3 test PASS, build sạch, `-h` in usage ra stderr và exit `2` theo cơ chế của `flag.ErrHelp` — `execute` coi đó là input error nên in exit 1; chấp nhận được cho một CLI lab, không cần thêm branch riêng.

- [ ] **Step 5: Smoke test không cần network**

```bash
HF_TOKEN="" go run ./cmd/genre-classify README.md
echo "exit=$?"
```

Expected: in `genre-classify: HF_TOKEN is empty; ...` và `exit=1` — đúng bảng exit code ở spec §8 (token thiếu là lỗi phía ta, chặn trước khi đọc ảnh).

- [ ] **Step 6: Commit**

```bash
git add cmd
git commit -m "Add the genre-classify command

Argument handling is split from process exit so the parsing rules are testable
without spawning a subprocess. The token is required before the photo is even
read, so a missing credential fails fast with exit 1 instead of looking like a
model error."
```

---

## Task 13: File taxonomy chạy thật và tài liệu đi kèm

**Files:**
- Create: `ai/genre-classifier/taxonomy/genre-v1.json`
- Modify: `ai/genre-classifier/README.md`
- Modify: `ai/model-provider/README.md`

**Interfaces:**
- Consumes: `LoadTaxonomy` + `validate` (Task 1, 2).
- Produces: file mặc định của CLI (`--taxonomy` default ở Task 12).

- [ ] **Step 1: Viết `ai/genre-classifier/taxonomy/genre-v1.json`**

Bảy ứng viên thể loại — `other` đứng ngoài, đúng luật gán:

```json
{
  "version": "genre-v1",
  "prompt_template": "a photo of {label}",
  "fallback_genre": "other",
  "thresholds": {
    "genre_min_share_ratio": 1.5,
    "genre_min_margin_ratio": 0.5,
    "genre_alternatives_max": 3,
    "tag_min_share_ratio": 1.0,
    "tag_max": 3
  },
  "candidates": [
    { "group": "genre", "id": "portrait", "label": "a person as the main subject" },
    { "group": "genre", "id": "landscape", "label": "a natural landscape as the main subject" },
    { "group": "genre", "id": "architecture", "label": "a building as the main subject" },
    { "group": "genre", "id": "street", "label": "a candid street scene" },
    { "group": "genre", "id": "food", "label": "food as the main subject" },
    { "group": "genre", "id": "product", "label": "a product as the main subject" },
    { "group": "genre", "id": "animal", "label": "an animal as the main subject" },
    { "group": "tag", "id": "outdoor", "label": "outdoors" },
    { "group": "tag", "id": "indoor", "label": "indoors" }
  ]
}
```

Với `N_genre = 7`: mức đều `1/7 ≈ 0.1429`, floor thực tế `≈ 0.2143`, margin thực tế `≈ 0.0714`. Với `N_tag = 2`: floor tag `0.5` — tag nào không vượt quá nửa tổng điểm nhóm tag thì không được ghi.

- [ ] **Step 2: Nạp thử file thật bằng một test**

Nối vào `internal/genreclassifier/taxonomy_test.go`:

```go
func TestProductionTaxonomyLoads(t *testing.T) {
	tax, err := LoadTaxonomy(filepath.Join("..", "..", "ai", "genre-classifier", "taxonomy", "genre-v1.json"))
	if err != nil {
		t.Fatalf("genre-v1.json must load: %v", err)
	}
	if got, want := len(tax.Candidates(GroupGenre)), 7; got != want {
		t.Errorf("genre candidates = %d, want %d", got, want)
	}
	if tax.FallbackGenre == "other" {
		for _, c := range tax.Cands {
			if c.ID == "other" {
				t.Error("other must not compete as a candidate")
			}
		}
	}
}
```

```bash
go test ./internal/genreclassifier/ -run TestProductionTaxonomyLoads -v
```

Expected: PASS. Đây là test giữ cho file dữ liệu production không hỏng âm thầm — `validate()` chạy y như lúc CLI nạp.

- [ ] **Step 3: Cập nhật `ai/genre-classifier/README.md`**

Thay dòng 7:

```markdown
Trạng thái: **chưa có code**.
```

bằng:

```markdown
Trạng thái: **đang implement** theo [spec](../../docs/superpowers/specs/2026-10-25-genre-classifier-pipeline-design.md) và [plan](../../docs/superpowers/plans/2026-10-26-genre-classifier.md). Code: `internal/genreclassifier` (package) và `cmd/genre-classify` (CLI); dữ liệu: [taxonomy/genre-v1.json](taxonomy/genre-v1.json).
```

Sửa câu ở mục "Chọn nhãn và đặt trạng thái" (dòng 70). Trước:

```markdown
Hai ngưỡng (ứng viên hợp lệ, margin) là **tham số chỉnh**, chưa chốt giá trị; phải lấy từ tập ảnh mẫu ở [ai/evaluation](../evaluation/README.md).
```

Sau:

```markdown
Hai ngưỡng (ứng viên hợp lệ, margin) là **tham số chỉnh**, hiện đặt ở điểm khởi điểm `1.5×` và `0.5×` **mức đều của nhóm** (`1/N`), ghi trong [taxonomy/genre-v1.json](taxonomy/genre-v1.json); phải hiệu chỉnh trên tập ảnh mẫu ở [ai/evaluation](../evaluation/README.md). Ngưỡng không thể là điểm tuyệt đối vì một lượt gọi chấm chung hai nhóm nên chỉ còn `q` đã renormalize là có nghĩa.
```

Sửa câu ở mục "Hợp đồng" (dòng 58). Trước:

```markdown
`score` là điểm xếp hạng nhãn, không phải xác suất đã hiệu chỉnh. `status` là `classified` hoặc `needs_review`.
```

Sau:

```markdown
`score` là **tỷ lệ trong nhóm** (`q`, làm tròn 4 chữ số), không phải xác suất đã hiệu chỉnh. `alternatives` liệt kê tối đa 3 nhãn đứng sau nhãn chính, và **không chứa `other`** vì nhãn này không tranh đua bằng điểm. `status` là `classified` hoặc `needs_review`; khi `primary_genre` là `other`, `alternatives` vẫn gồm cả nhãn dẫn đầu để người xem có chẩn đoán.
```

- [ ] **Step 4: Cập nhật `ai/model-provider/README.md`**

Thay bullet cuối mục "Yêu cầu giao diện" (dòng 52). Trước:

```markdown
- Điểm cho từng nhãn ứng viên, theo đúng thứ tự/không tự thêm nhãn ngoài danh sách.
```

Sau:

```markdown
- Điểm cho từng nhãn ứng viên, theo đúng thứ tự/không tự thêm nhãn ngoài danh sách. **Điều kiện cứng với bước 1:** trả đủ điểm cho **toàn bộ** ứng viên đưa lên, không phải top-k. Thiếu một ứng viên là `error`, vì renormalize trong nhóm cần tổng đủ mọi điểm (spec §3.1). Chưa kiểm chứng bằng chạy thật — xem [integration probe](../../docs/superpowers/plans/2026-10-26-genre-classifier.md#task-14-kiểm-chứng-bằng-một-lượt-gọi-thật).
```

- [ ] **Step 5: Chạy toàn bộ và commit**

```bash
go test ./... && go vet ./...
git add ai docs internal/genreclassifier/taxonomy_test.go
git commit -m "Ship the genre-v1 taxonomy and record its thresholds

Seven competing labels plus other, which the rules assign rather than the
model scoring. Thresholds are stated as multiples of the group uniform so
they survive a change in label count, and the README now says that instead
of leaving them open. The provider requirement to score every candidate is
written into the model-provider contract because the merged call depends on it."
```

---

## Task 14: Kiểm chứng bằng một lượt gọi thật

**Files:**
- Create: `cmd/genre-classify/integration_test.go`

**Interfaces:**
- Consumes: `OpenPhoto`, `LoadTaxonomy`, `NewHuggingFace`, `Decide` (Task 1–9).
- Produces: bằng chứng cho ba giả định ở spec §12.

Hai giả định kiểm chứng bằng code ở task này; giả định thứ ba (mô tả ngắn hay dài tốt hơn) là thí nghiệm trên tập ảnh, thuộc [ai/evaluation](../../../ai/evaluation/README.md) — plan này không tự xưng làm được.

- [ ] **Step 1: Viết probe**

`cmd/genre-classify/integration_test.go`:

```go
//go:build integration

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	gc "github.com/minhdz-2005/photo-aesthetics-lab/internal/genreclassifier"
)

// TestLiveModelScoresEveryCandidate checks the two wire assumptions of spec §12:
// the endpoint returns a score for all candidates (not top-k) and accepts the
// base64 data-url body. Run with:
//
//	go test -tags integration ./cmd/genre-classify/ -v
func TestLiveModelScoresEveryCandidate(t *testing.T) {
	token := os.Getenv("HF_TOKEN")
	if token == "" {
		t.Skip("HF_TOKEN not set")
	}
	path := os.Getenv("GENRE_TEST_PHOTO")
	if path == "" {
		t.Skip("GENRE_TEST_PHOTO not set (path to a real .jpg or .png)")
	}

	photo, err := gc.OpenPhoto(path)
	if err != nil {
		t.Fatal(err)
	}
	tax, err := gc.LoadTaxonomy(filepath.Join("..", "..", "ai", "genre-classifier", "taxonomy", "genre-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	provider, err := gc.NewHuggingFace(gc.DefaultModel, token)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	scores, err := provider.Score(ctx, photo, tax.Cands)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	if len(scores) != len(tax.Cands) {
		t.Fatalf("got %d scores for %d candidates: the endpoint truncates, so the merged call is invalid",
			len(scores), len(tax.Cands))
	}
	for _, c := range tax.Cands {
		if _, ok := scores[c.ID]; !ok {
			t.Errorf("no score for %q", c.ID)
		}
	}

	dec, detail, err := gc.Decide(scores, tax)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("photo=%s %dx%d", photo.Path, photo.Width, photo.Height)
	for _, s := range detail.Genre {
		t.Logf("genre %-13s raw=%.4f q=%.4f", s.Candidate.ID, s.Raw, s.Q)
	}
	for _, s := range detail.Tags {
		t.Logf("tag   %-13s raw=%.4f q=%.4f", s.Candidate.ID, s.Raw, s.Q)
	}
	t.Logf("-> primary=%s status=%s rule=%s tags=%v", dec.PrimaryGenre, dec.Status, detail.Rule, dec.SecondaryTags)
}
```

- [ ] **Step 2: Xác nhận test mặc định vẫn offline**

```bash
go test ./...
go vet -tags integration ./...
```

Expected: `go test ./...` PASS mà không cần token (probe không biên dịch vào); `go vet -tags integration` sạch.

- [ ] **Step 3: Chạy probe với một ảnh thật**

```bash
HF_TOKEN="<token>" GENRE_TEST_PHOTO="D:/path/to/portrait-outdoor.jpg" \
  go test -tags integration ./cmd/genre-classify/ -v
```

Expected: hoặc PASS kèm log `q` từng nhãn, hoặc FAIL chỉ ra đúng một trong hai giả định sụp. **Không được coi §12 là đã chốt trước khi thấy dòng log này.**

Nếu endpoint trả top-k: dừng, quay về hai lượt gọi tách cho genre và tag, và cập nhật spec §3.1 — toàn bộ thiết kế ngưỡng dựa trên việc hợp nhất.

- [ ] **Step 4: Ghi kết quả vào tài liệu**

Sau khi chạy, thay câu "Chưa kiểm chứng bằng chạy thật" ở `ai/model-provider/README.md` (Task 13 Step 4) bằng một trong hai:

```markdown
Đã kiểm chứng bằng integration probe ngày YYYY-MM-DD: model trả đủ điểm cho 9/9 ứng viên.
```

hoặc, nếu ngược lại:

```markdown
Integration probe ngày YYYY-MM-DD cho thấy endpoint trả top-k. Lượt gọi hợp nhất vì thế không dùng được cho bước 1; xem spec §3.1.
```

- [ ] **Step 5: Commit**

```bash
git add cmd/genre-classify/integration_test.go ai/model-provider/README.md
git commit -m "Probe the live endpoint for the merged-call assumption

The whole threshold scheme rests on the model scoring every candidate, which
documentation does not guarantee. The probe is build-tagged so the default test
run stays offline and fast, and the finding gets written back into the provider
contract instead of staying in a terminal buffer."
```

---

## Definition of done

Chạy đủ danh sách này trước khi báo hoàn thành:

```bash
go vet ./...
go test -race ./...
go test ./... -count=5              # tie-break phải tất định dù map duyệt ngẫu nhiên
go build -o bin/genre-classify.exe ./cmd/genre-classify
HF_TOKEN="" ./bin/genre-classify.exe README.md; echo "exit=$?"   # -> exit=1
```

Và tự kiểm từng điều:

- [ ] `go test ./...` không gọi network, không cần token, chạy dưới 1 giây.
- [ ] `go.sum` không tồn tại; `go.mod` không có require nào ngoài stdlib.
- [ ] stdout của một lệnh thành công parse được thành đúng 4 khóa `primary_genre`, `secondary_tags`, `alternatives`, `status` — và chỉ 4 khóa đó.
- [ ] `needs_review` trả exit `0`; `other` không bao giờ xuất hiện trong `alternatives`.
- [ ] Mỗi ảnh đúng **một** lượt gọi model (`fp.calls == 1` trong `TestRunWritesContractToStdout`).
- [ ] Không file nhị phân nào trong `testdata/` (ảnh test sinh trong test).
- [ ] `frontend/` không có thay đổi nào: `git diff --name-only HEAD~N -- frontend` rỗng.
- [ ] Commit cuối không có `Co-Authored-By` trailer.
