# Thiết kế pipeline — genre-classifier (bước 1)

Ngày: 2026-10-25
Trạng thái: **chờ duyệt**. Chưa có code Go nào được viết cho tới khi bản này được duyệt.

Nguồn hợp đồng: [docs/step-01-image-genre-classification.md](../../../docs/step-01-image-genre-classification.md) và [ai/genre-classifier/README.md](../../../ai/genre-classifier/README.md).
Bản thiết kế này **không** thay đổi hợp đồng bước 1; nó chỉ định cách thực thi hợp đồng đó bằng Go.

---

## 1. Phạm vi

**Trong phạm vi:** một package Go + CLI nhận một ảnh, gọi model zero-shot, trả JSON đúng hợp đồng bước 1.

**Ngoài phạm vi:** điều phối các bước, xác định style (bước 2), chấm điểm thẩm mỹ (bước 4), tập ảnh mẫu và đo độ chính xác ([ai/evaluation](../../../ai/evaluation/README.md)), lưu kết quả hàng loạt.

**Câu hỏi thiết kế trung tâm:** pipeline nào biến một điểm số không hiệu chỉnh được từ model thành `primary_genre` + `status` một cách tất định và kiểm chứng được mà không gọi model?

---

## 2. Ba quyết định nền đã chốt trước khi vào pipeline

| Quyết định | Chọn | Hệ quả |
| --- | --- | --- |
| Provider model | Hugging Face Inference API (host đã train sẵn) | Package không tải trọng số model, chỉ stdlib + HTTP; đổi host chỉ sửa `huggingface.go` |
| Hình thái bàn giao | Package `internal/genreclassifier` + CLI in JSON ra stdout | Logic luật nằm trong package để test và tái dùng; `main.go` chỉ nối cờ |
| Tag phụ (`secondary_tags`) | Bộ từ vựng tag **riêng**, tách khỏi danh sách thể loại | `portrait` = thể loại, `outdoor` = tag; hai nhóm điểm không so với nhau |

---

## 3. Pipeline — năm tầng

```
image.OpenPhoto(path)
      ↓  Photo{Bytes, Width, Height}
taxonomy.LoadTaxonomy(dir)
      ↓  Taxonomy{Version, PromptTemplate, FallbackGenre, Thresholds, Candidates}
provider.Score(ctx, photo, candidates)          ← MỘT lượt gọi hợp nhất
      ↓  Scores map[id]float64  (softmax trên hợp hai nhóm)
decide.Decide(scores, taxonomy)
      ↓  Decision{PrimaryGenre, SecondaryTags, Alternatives, Status, debug}
result.Result → Marshal → stdout
```

Mỗi tầng chỉ biết tầng trên mình. `decide` là hàm thuần, không I/O. `run` nối các tầng và không chứa luật nào.

### 3.1 Vì sao chỉ một lượt gọi

Thể loại và tag được chấm trong **cùng một lượt** trên hợp của hai danh sách, rồi **renormalize trong từng nhóm**:

```
q_i = p_i / Σ_{j ∈ nhóm(i)} p_j
```

Vì mọi `p_i` đều sinh ra từ cùng một partition function của softmax trên hợp, mẫu số chung bị **triệt tiêu chính xác** khi chia lại trong nhóm. Kết quả: xếp hạng nội bộ mỗi nhóm giống hệt hai lượt gọi tách riêng, nhưng giảm một nửa độ trễ và chi phí.

**Điều kiện cứng của provider:** API phải trả điểm cho **toàn bộ** danh sách ứng viên đưa lên, không phải top-k. Nếu endpoint nào chỉ trả top-k, renormalize trong nhóm sẽ sai và thiết kế này không áp dụng được. Phải kiểm chứng bằng một lượt gọi thật, không tin tài liệu.

**Hệ quả đã biết và đã chấp nhận:** sau renormalize, điểm tuyệt đối mất ý nghĩa. Mọi ngưỡng phải diễn đạt **dưới dạng bội số của mức đều `1/N`** (xem §5). Đây là cái giá của việc hợp nhất, và nó là giá đúng.

---

## 4. Tầng `image`

- Hỗ trợ **JPEG và PNG** qua stdlib (`image/jpeg`, `image/png`). WebP cần thư viện ngoài → **không** làm ở bản này; ảnh webp trả lỗi rõ ràng.
- Từ chối: file không tồn tại, không đọc được, không giải mã được, cạnh bằng 0.
- Trả nguyên `Bytes` chưa re-encode. Model CLIP resize nội bộ 224/384px; việc giảm cỡ ảnh phía client chỉ thêm một chỗ làm mờ ảnh và mất chất lượng.
- `Width`/`Height` chỉ dùng cho metadata debug, **không** dùng để suy luận. Tên file không mang thông tin gì cho mô hình.

---

## 5. Tầng `taxonomy`

### 5.1 Vị trí và lý do

`ai/genre-classifier/taxonomy/genre-v1.json`, **nạp từ đĩa lúc chạy, không `go:embed`**.

Lý do: file này là dữ liệu thuộc `taxonomy_version`, cần đứng cạnh spec của nó. Đổi ngưỡng không phải build lại — đây là điều kiện để vòng hiệu chỉnh ngưỡng ở `ai/evaluation` chạy được. Giá phải trả là mặc định đường dẫn giả định chạy từ gốc repo; lỗi thiếu file **phải in đường dẫn tuyệt đối đã resolve**.

### 5.2 Định dạng

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
    { "group": "genre", "id": "portrait",     "label": "a person as the main subject" },
    { "group": "genre", "id": "landscape",    "label": "a natural landscape as the main subject" },
    { "group": "genre", "id": "architecture", "label": "a building as the main subject" },
    { "group": "genre", "id": "street",       "label": "a candid street scene" },
    { "group": "genre", "id": "food",         "label": "food as the main subject" },
    { "group": "genre", "id": "product",      "label": "a product as the main subject" },
    { "group": "genre", "id": "animal",       "label": "an animal as the main subject" },
    { "group": "tag",   "id": "outdoor",      "label": "outdoors" },
    { "group": "tag",   "id": "indoor",       "label": "indoors" }
  ]
}
```

Bảy ứng viên thể loại + `other`. **`other` không xuất hiện trong `candidates`** — nó chỉ do luật gán (§6). Mô tả 7 ứng viên mang nghĩa thị giác rõ; "a photo of other" thì không, và sẽ ăn điểm ngẫu nhiên.

`label` là phần tử được điền vào `{label}`, **không phải** id trong hợp đồng. `id` là khóa ổn định trong JSON; đổi wording của `label` là đổi `taxonomy_version`, đổi `id` là đổi hợp đồng.

### 5.3 Ngưỡng — giá trị khởi điểm, chưa hiệu chỉnh

Dùng `N_genre = 7` → mức đều `1/7 ≈ 0.1429`:

| Ngưỡng | Hệ số | Ý nghĩa ở giá trị khởi điểm |
| --- | --- | --- |
| `genre_min_share_ratio` | 1.5 | Nhãn dẫn đầu phải đạt ≈ **0.214** tổng điểm nhóm, tức nhỉnh hơn mức đều 50% |
| `genre_min_margin_ratio` | 0.5 | Khoảng cách tới nhãn hai phải đạt ≈ **0.071**, tức nửa mức đều |
| `tag_min_share_ratio` | 1.0 | Tag phải trên mức đều của nhóm tag |
| `genre_alternatives_max` | 3 | Số nhãn chẩn đoán giữ lại |
| `tag_max` | 3 | Số tag phụ tối đa |

Các số này là **điểm neo để bắt đầu**, không phải kết quả nghiên cứu. Sau khi chạy tập mẫu, ta đổi hệ số trong file JSON mà không đụng code.

### 5.4 Tính hợp lệ lúc nạp

Chặn tại boundary, báo lỗi kèm `file:trường`:

- `version`, `prompt_template`, `fallback_genre` không rỗng; `prompt_template` phải chứa `{label}`.
- `id` duy nhất trong mỗi `group`; mỗi `id` thuộc đúng một `group`.
- `fallback_genre` **không** được nằm trong `candidates`.
- Cần ≥ 2 ứng viên `genre`; nhóm `tag` có thể rỗng (khi đó `secondary_tags` luôn `[]`).
- Mọi ngưỡng `>= 0`; hai `*_max` là số nguyên `> 0`.

### 5.5 `taxonomy_version` nằm ngoài hợp đồng

Hợp đồng bước 1 không có chỗ cho nó. Version đi vào `--debug` (stderr) và vào bản ghi của `ai/evaluation`. Đây là quyết định có ý thức: giữ hợp đồng đúng như docs, kể cả khi nó làm metadata khó truy vết hơn.

---

## 6. Tầng `decide` — thứ tự áp luật

Đầu vào: `Scores map[string]float64` + `Taxonomy`. Đầu ra: `Decision` + thông tin chẩn đoán.

| # | Bước | Chi tiết |
| --- | --- | --- |
| 1 | **Thiếu bằng chứng là lỗi** | Mỗi `Candidate` phải có điểm. Thiếu bất kỳ id nào → `error`. Không loại bớt ứng viên rồi chấm tiếp. |
| 2 | **Sắp trong nhóm** | Giảm dần theo điểm. Điểm hòa → theo thứ tự khai báo trong file taxonomy. Kết quả tất định, không phụ thuộc thứ tự duyệt map. |
| 3 | **Renormalize trong nhóm** | `q_i = p_i / Σ_{j∈nhóm} p_j`. `genre` và `tag` tính riêng. **Không bao giờ so điểm xuyên nhóm.** |
| 4 | **Floor thể loại** | `q_1 >= genre_min_share_ratio × (1/N_genre)`. Trượt → `primary_genre = fallback_genre`, `status = needs_review`, **dừng**, không xét bước 5. |
| 5 | **Margin thể loại** | `(q_1 - q_2) >= genre_min_margin_ratio × (1/N_genre)`. Trượt → `primary_genre` vẫn là nhãn dẫn đầu, `status = needs_review`. |
| 6 | **Chốt** | Qua cả 4 và 5 → `status = classified`. |
| 7 | **Tag phụ** | Lọc tag có `q >= tag_min_share_ratio × (1/N_tag)`, lấy tối đa `tag_max` theo thứ tự bước 2. Rỗng là kết quả hợp lệ. |
| 8 | **Alternatives** | Lấy `genre_alternatives_max` ứng viên `genre` có `q` cao nhất **sau nhãn đã chọn làm `primary_genre`**; mỗi phần tử `{genre, score}` với `score = q_i`. Khi bước 4 gán `other` thì không có nhãn nào bị loại, nên alternatives gồm cả nhãn dẫn đầu. **Không bao giờ chứa `other`.** |

### 6.1 Hai trường hợp biên phải nói rõ

- **Nhóm chỉ có một ứng viên:** không có `q_2` nên bước 5 được bỏ qua — không có đối thủ để tách. Floor vẫn chặn, nên không lách qua luật được.
- **Mọi điểm bằng 0** (tổng nhóm = 0): chia cho 0. Đây là **lỗi cấu hình**, không phải ca `needs_review`. `decide` trả `error` trước khi tới bước 3.

### 6.2 Ánh xạ vào hợp đồng

| Luật | `primary_genre` | `status` |
| --- | --- | --- |
| Qua floor + qua margin | nhãn dẫn đầu | `classified` |
| Qua floor, trượt margin | nhãn dẫn đầu | `needs_review` |
| Trượt floor | `other` | `needs_review` |

Ba dòng này khớp đúng bảng "Chọn nhãn và đặt trạng thái" trong `ai/genre-classifier/README.md`.

---

## 7. Tầng `provider`

### 7.1 Giao diện

```go
type Provider interface {
    Score(ctx context.Context, photo Photo, candidates []Candidate) (map[string]float64, error)
}
```

`huggingface.go` triển khai trên `openai/clip-vit-large-patch14` (cờ `--model` cho ghi đè). Token đọc từ biến môi trường `HF_TOKEN`.

Provider **không** đặt ngưỡng, **không** quyết `other` hay `needs_review` — nó chỉ vận chuyển điểm, theo đúng ranh giới đã ghi ở [ai/model-provider/README.md](../../../ai/model-provider/README.md).

### 7.2 Thời hạn và retry

- Một `context` có deadline cho cả lệnh; cờ `--timeout` mặc định **30s**. Retry nằm **trong** deadline, không kéo dài tổng thời gian.
- **Tối đa 2 lượt gọi.** Retry chỉ cho: lỗi kết nối, `429`, `5xx`. Không retry cho `4xx` khác (sai token, sai request).
- Backoff 500ms, nhân 2, có jitter. `429` kèm `Retry-After` hợp lệ thì đợi theo `Retry-After`.
- Timeout giữa chừng → `error`. **Không** hạ xuống `needs_review`: hai thứ khác nhau — `needs_review` là model trả kết quả mà mình không chốt được, timeout là mình không có kết quả.

### 7.3 Kết quả không hợp lệ

Coi là `error`, không đoán thay thế: HTTP ngoài 2xx; response thiếu ứng viên; điểm không phải số; NaN/Inf; body không parse được. Trả về kèm model id + version đã dùng để ghi vào debug metadata.

---

## 8. CLI và mã exit

```
genre-classify [flags] <anh>
  --taxonomy  ai/genre-classifier/taxonomy/genre-v1.json
  --model     openai/clip-vit-large-patch14
  --timeout   30s
  --debug     metadata ra stderr
```

Quy tắc phân định: **`0` = có kết quả hợp đồng** (kể cả `needs_review` — đây là trạng thái làm việc đúng, không phải lỗi); **`1` = đầu vào của mình sai**; **`2` = thế giới bên ngoài trả về thứ không dùng được**.

| Tình huống | Tầng | Mã exit | stdout |
| --- | --- | --- | --- |
| Ảnh thiếu/không đọc được/không giải mã được/cạnh 0 | `image` | `1` | rỗng |
| Thiếu file taxonomy, JSON hỏng, `id` trùng, `fallback_genre` nằm trong `candidates`, thiếu `{label}`, ngưỡng âm | `taxonomy` | `1` | rỗng |
| Thiếu `HF_TOKEN` | `main` | `1` | rỗng |
| Không kết nối được, HTTP ngoài 2xx, quá hạn | `provider` | `2` | rỗng |
| Response thiếu ứng viên, điểm không phải số, NaN/Inf | `provider` | `2` | rỗng |
| Thiếu điểm khi vào `decide` | `decide` | `2` | rỗng |
| Xếp hạng, renormalize, gán nhãn | — | `0` | JSON hợp đồng |

**stdout chỉ chứa JSON hợp đồng**, không kèm log. Metadata debug (`taxonomy_version`, model id + version, điểm thô, `q_i`, luật nào bắn ở bước 4/5/7) in ra stderr khi có `--debug`.

---

## 9. Kiểm chứng

Nguyên tắc phân định: **test của package chứng minh luật chạy đúng, không chứng minh nhận diện đúng.** Độ chính xác thuộc `ai/evaluation`. Trộn hai thứ vào nhau nghĩa là mỗi lần đổi ngưỡng lại phải gọi API thật để biết mình có phá gì không.

### 9.1 Năm tầng test

| Tầng | Dựng | Cover |
| --- | --- | --- |
| `decide` | Bảng thuần, không mock, không I/O | Thứ tự luật, tie-break, ca biên |
| `taxonomy` | Fixture trong `testdata/taxonomy/` | từng quy tắc ở §5.4 một ca |
| `image` | Ảnh thật nhỏ trong `testdata/images/` | hợp lệ, 0 byte, hỏng, văn bản đổi đuôi `.jpg`, cạnh 0 |
| `provider` | `httptest.Server` trả JSON dựng sẵn | 2xx/4xx/5xx, thiếu ứng viên, NaN, deadline, retry, `Retry-After` |
| `run` | fake provider bơm qua seam | mã exit, stdout nguyên văn, debug không lẫn vào stdout |

### 9.2 Seam cho CLI

```go
func run(ctx context.Context, d deps) int
type deps struct {
    provider Provider
    open     openPhotoFunc
    stdout   io.Writer
    stderr   io.Writer
}
```

`main.go` chỉ làm việc parse cờ và `os.Exit(run(...))`. Test gọi `run` với fake provider và `bytes.Buffer`.

### 9.3 Bảng `decide`

Fixture dùng taxonomy test **4 genre + 2 tag** cho số tự kiểm chứng được: `N_genre = 4` → đều `0.25`, floor `1.5×` = **0.375**, margin `0.5×` = **0.125**; `N_tag = 2` → đều `0.5`, floor tag = **0.5**.

| Ca | `q` genre (giảm dần) | `q` tag | Mong đợi |
| --- | --- | --- | --- |
| Qua cả hai | 0.55 / 0.20 / 0.15 / 0.10 | 0.70 / 0.30 | `classified`, nhãn dẫn đầu, 1 tag, 3 alternatives |
| Trượt floor | 0.28 / 0.26 / 0.24 / 0.22 | 0.51 / 0.49 | `other` + `needs_review`, **không xét margin**, alternatives vẫn có |
| Trượt margin | 0.40 / 0.35 / 0.15 / 0.10 | 0.70 / 0.30 | nhãn dẫn đầu + `needs_review` |
| Hòa nhãn đầu | 0.50 / 0.50 / ... | — | `needs_review`, `primary_genre` theo thứ tự khai báo |
| Không tag nào qua | 0.60 / ... | dưới ngưỡng | `secondary_tags: []` |
| Vượt `tag_max` | — | 4 tag cùng qua | cắt còn `tag_max`, theo điểm |
| Thiếu điểm | 3/4 có điểm | — | `error` |
| Một ứng viên genre | — | — | bỏ qua margin, floor vẫn chặn |
| Mọi điểm bằng 0 | 0 / 0 / 0 / 0 | — | `error`, không sinh NaN |

Ca cuối bắt buộc: renormalize trên tổng bằng 0 là chỗ duy nhất `decide` có thể sinh NaN thay vì lỗi.

### 9.4 Hai phép kiểm chứng riêng cho quyết định hợp nhất

Không ca nào ở §9.3 bắt được hai rủi ro này:

- **Bất biến theo tỉ lệ:** nhân **toàn bộ** điểm thô với cùng một thừa số `c > 0` thì kết quả `decide` phải không đổi — đây là phát biểu đúng của tính bất biến. Một hàm đơn điệu *phi tuyến* thì **không** giữ kết quả: nó đổi tỷ lệ giữa các `q` nên có thể đổi cả `status`. Điều duy nhất mọi hàm đơn điệu đảm bảo là thứ tự trong nhóm không đổi.
- **Độc lập xuyên nhóm:** hai đầu vào chỉ khác điểm `tag` phải cho phần `genre` (`primary_genre`, `status`, `alternatives`) **giống nhau từng byte**. Đây là cách kiểm chứng câu "không so điểm xuyên nhóm" mà không cần đọc code.

### 9.5 Golden JSON

Một ca chuẩn ghi `testdata/golden/portrait-classified.json`, so nguyên văn sau khi marshal lại (không so từng field). Lý do: nó chặn đồng thời key thừa, key thiếu, và **`secondary_tags: null` thay vì `[]`** — Go marshal slice nil thành `null`, nên `decide` phải luôn khởi tạo mảng.

### 9.6 Ranh giới network

`go test ./...` chạy offline, dưới 1 giây. Test bắn thật lên Hugging Face đặt sau `//go:build integration` và một cổng biến môi trường, chạy bằng `go test -tags integration ./...`. Vòng hiệu chỉnh ngưỡng cần lặp lại nhiều lần; phụ thuộc network thì nó sẽ thành chỗ bị bỏ qua.

`go test -race ./...` bật `-race` vì `provider` có retry và đọc `body`.

---

## 10. Bố cục file

`go.mod` ở **gốc repo**, một module duy nhất cho cả lab:

```
module github.com/minhdz-2005/photo-aesthetics-lab
go 1.27
```

Chỉ dùng stdlib → không `go.sum`, không kéo runtime model.

```
go.mod
cmd/genre-classify/
  main.go                      // parse cờ → ctx timeout → run → os.Exit
  main_test.go
internal/genreclassifier/
  pipeline.go                  // deps + run(): nối 5 tầng, không có luật
  taxonomy.go                  // Taxonomy/Candidate/Thresholds + LoadTaxonomy + validate
  image.go                     // OpenPhoto
  provider.go                  // interface Provider + kiểu dùng chung + lỗi
  huggingface.go               // HTTP, gộp ứng viên, retry, model id/version
  decide.go                    // hàm thuần §6
  contract.go                  // Result + Marshal, json tag snake_case
  *_test.go                    // một file test cho mỗi file trên
  testdata/
    taxonomy/min-v1.json                       // 4 genre + 2 tag
    taxonomy/bad-duplicate-id.json
    taxonomy/bad-fallback-in-candidates.json
    taxonomy/bad-no-placeholder.json
    taxonomy/bad-negative-threshold.json
    images/portrait.jpg
    images/zero-bytes.jpg
    images/text-named-jpg.jpg
    golden/portrait-classified.json
ai/genre-classifier/
  README.md                    // nguồn sự thật của hợp đồng
  taxonomy/genre-v1.json       // dữ liệu chạy thật, thuộc taxonomy_version
```

**Một package, nhiều file.** Năm tầng là một luồng tuyến tính, không phải năm thành phần thay thế độc lập. Test chạm được internals mà không phải xuất khẩu những thứ chỉ phục vụ test. Khi `style-analyzer` cần `provider`, promotion ra package riêng mới có lý do thực tế.

`.gitignore` bổ sung: `bin/`, `*.exe`, `*.exe~`, `*.out`, `/genre-classify`.

`frontend/` không bị đụng tới.

---

## 11. Việc tài liệu kèm theo

- `ai/genre-classifier/README.md`: đổi "Trạng thái: **chưa có code**" → trỏ tới `internal/genreclassifier` + `cmd/genre-classify`; ghi giá trị ngưỡng khởi điểm vào mục "Chọn nhãn và đặt trạng thái" thay vì "chưa chốt giá trị"; nêu `q_i` là điểm trong `alternatives`.
- `ai/model-provider/README.md`: ghi điều kiện cứng §3.1 (API phải trả đủ mọi ứng viên) vào mục "Yêu cầu giao diện".

---

## 12. Giả định phải kiểm chứng bằng chạy thật

| Giả định | Cách kiểm chứng | Nếu sai |
| --- | --- | --- |
| HF zero-shot trả điểm cho **toàn bộ** danh sách `image_labels`, không top-k | một lượt gọi integration với 9 ứng viên | bỏ hợp nhất, quay lại hai lượt gọi riêng cho genre/tag |
| `clip-vit-large-patch14` chấp nhận mô tả câu dài ("a person as the main subject") tốt hơn từ đơn | chạy §13 với 2 template trên cùng tập ảnh | đổi `prompt_template`, tăng `taxonomy_version` |
| Endpoint nhận raw image bytes làm body | một lượt gọi integration | đổi sang multipart / base64 tùy tài liệu endpoint thực tế |

Không xây tiếp trên ba giả định này khi chưa thấy kết quả thật.

## 13. Để ngỏ — nằm ngoài bản thiết kế này

- Ngưỡng (§5.3) chỉnh theo tập mẫu nào, cỡ bao nhiêu ảnh, tỷ lệ giao thoa bao nhiêu.
- Bộ từ vựng tag đầy đủ (bản này chỉ neo `outdoor`/`indoor`).
- `secondary_tags` có cần điểm riêng như `alternatives` không — hợp đồng hiện tại không cho.
- Cache kết quả theo hash ảnh, vì một ảnh một lượt gọi 30s sẽ làm vòng hiệu chỉnh rất chậm.
- Batch cho `ai/evaluation`.
- Bước 2 (`style-analyzer`) sẽ tái dùng `Provider` hay cần interface khác.
