# genre-classifier — Nhận diện thể loại ảnh

Bước 1 trong luồng. Nhận một ảnh và trả thể loại chính để các bước sau chọn đúng bộ tiêu chí.

Chi tiết gốc: [docs/step-01-image-genre-classification.md](../../docs/step-01-image-genre-classification.md)

Trạng thái: **chưa có code**. Bản Go đầu tiên đã bị gỡ vì cấu trúc quá cồng kềnh so với việc chỉ gọi một model; bước 1 sẽ làm lại theo cấu trúc gọn hơn. Thứ còn dùng được: nhãn, prompt template và ngưỡng trong [taxonomy/genre-v1.json](taxonomy/genre-v1.json).

## Nhiệm vụ

- Chọn **một thể loại chính** dựa trên chủ thể nổi bật nhất của ảnh.
- Giữ **tag phụ** cho ngữ cảnh khi ảnh thuộc nhiều nhóm.
- Đánh dấu **`needs_review`** khi kết quả mơ hồ thay vì ép chọn.

Bước này **không** đánh giá ảnh đẹp hay chưa, **không** xác định style, **không** chỉnh sửa ảnh.

## Danh sách nhãn

`portrait` · `landscape` · `architecture` · `street` · `food` · `product` · `animal` · `other`

`portrait` nghĩa là ảnh tập trung vào người, không phải ảnh chụp chiều dọc. Đây là taxonomy ban đầu của lab; cần kiểm tra trên ảnh mẫu và điều chỉnh nếu các nhãn hay bị nhầm.

## Phương pháp: zero-shot + taxonomy

Model zero-shot nhận ảnh cùng **danh sách mô tả nhãn đưa vào lúc chạy** và trả điểm cho từng nhãn ứng viên.

Lý do chọn:

- **Taxonomy là cấu hình, không phải trọng số.** Thêm, sửa hoặc nhập nhãn không phải train lại — đúng điều lab cần khi danh sách 8 nhãn chỉ là giả thuyết ban đầu.
- **Điểm theo từng ứng viên khớp sẵn với hợp đồng.** `alternatives` và `status` có nguồn dữ liệu thật, không phải chế thêm.
- **MVP chưa có dữ liệu gán nhãn** để fine-tune; một ảnh mỗi lượt, cần kết quả chạy được ngay.
- docs bước 1 đã tham chiếu hướng này (zero-shot image classification).

Điểm yếu đã biết trước:

Zero-shot có thiên kiến đọc **bối cảnh tổng thể** của ảnh thay vì **chủ thể chính**. Ảnh một người đứng giữa núi nhiều khả năng trả `landscape`, trong khi spec muốn `portrait`. Đây đúng là nhóm ảnh giao thoa mà mục Kiểm chứng nhắm tới, nên phải đo riêng chứ không suy đoán.

Vì rủi ro nằm ở phương pháp chứ không ở hợp đồng, bước này chạy qua [model-provider](../model-provider/README.md) để cùng một tập ảnh có thể so với classifier fine-tune mà không sửa phần điều phối.

## Hợp đồng

**Input:** một ảnh hợp lệ.

**Output:**

```json
{
  "primary_genre": "landscape",
  "secondary_tags": ["outdoor"],
  "alternatives": [
    { "genre": "architecture", "score": 0.31 },
    { "genre": "portrait", "score": 0.12 }
  ],
  "status": "classified"
}
```

`score` là **tỷ lệ trong nhóm** (`q`, làm tròn 4 chữ số), không phải xác suất đã hiệu chỉnh. `alternatives` liệt kê tối đa 3 nhãn đứng sau nhãn chính, và **không chứa `other`** vì nhãn này không tranh đua bằng điểm. `status` là `classified` hoặc `needs_review`; khi `primary_genre` là `other`, `alternatives` vẫn gồm cả nhãn dẫn đầu để người xem có chẩn đoán.

## Chọn nhãn và đặt trạng thái

`other` vẫn là nhãn trong taxonomy, nhưng **gán theo luật chứ không tranh đua bằng embedding** — mô tả kiểu "a photo of other" không có nghĩa thị giác rõ và sẽ ăn điểm ngẫu nhiên.

| Tình huống | `primary_genre` | `status` |
| --- | --- | --- |
| Có nhãn vượt ngưỡng ứng viên, và bỏ xa nhãn kế tiếp | nhãn tốt nhất | `classified` |
| Nhãn tốt nhất và nhãn hai cách nhau dưới ngưỡng margin | nhãn tốt nhất | `needs_review` |
| Không nhãn nào vượt ngưỡng ứng viên | `other` | `needs_review` |

Hai ngưỡng (ứng viên hợp lệ, margin) là **tham số chỉnh**, hiện đặt ở điểm khởi điểm `1.5×` và `0.5×` **mức đều của nhóm** (`1/N`), ghi trong [taxonomy/genre-v1.json](taxonomy/genre-v1.json); phải hiệu chỉnh trên tập ảnh mẫu ở [ai/evaluation](../evaluation/README.md). Ngưỡng không thể là điểm tuyệt đối vì một lượt gọi chấm chung hai nhóm nên chỉ còn `q` đã renormalize là có nghĩa. Ảnh rơi vào `other` mặc định vẫn cần xem lại, vì spec cấm biến trường hợp mơ hồ thành `other` khi chưa có lý do.

## Phiên bản taxonomy

Câu mô tả gửi cho model cho từng nhãn **là một phần của taxonomy**, không phải chi tiết triển khai. Đổi wording làm kết quả đổi theo, nên:

- Lưu template prompt và wording của từng nhãn dưới một `taxonomy_version`.
- Ghi `taxonomy_version` cùng model và phiên bản đã dùng vào mỗi kết quả chạy.
- Không so hai lần chạy có `taxonomy_version` khác nhau.

## Quy tắc xử lý

1. Kiểm tra ảnh đọc được và có nội dung hình ảnh hợp lệ.
2. Gửi ảnh cùng danh sách mô tả nhãn của `taxonomy_version` đang dùng tới bộ phân loại.
3. Chọn nhãn chính theo chủ thể nổi bật nhất.
4. Lưu vài nhãn gần nhất làm thông tin chẩn đoán; chỉ dùng tag phụ khi hữu ích cho bước sau.
5. Mơ hồ thì trả `needs_review`; không tự chuyển sang `other` khi chưa có lý do.
6. Không làm style, điểm thẩm mỹ hay chỉnh sửa ảnh.

## Kiểm chứng

Chạy trên tập ảnh mẫu có nhãn thủ công ở [ai/evaluation](../evaluation/README.md), ưu tiên ảnh giao thoa (chân dung ngoài trời, kiến trúc trong cảnh phố, người trong phong cảnh). Ghi lại: tỷ lệ đúng, cặp nhãn hay nhầm, tỷ lệ cần xem lại, thời gian xử lý và lỗi khi model trả kết quả không hợp lệ.

Chạy song song hai phương án trên **cùng một tập ảnh** để so:

| Phương án | Được | Mất |
| --- | --- | --- |
| Zero-shot + taxonomy | Đổi nhãn không cần train lại; có điểm cho mọi ứng viên; chạy được ngay khi chưa có dữ liệu | Yếu ở cặp chủ thể-ngữ cảnh; wording của nhãn ảnh hưởng trực tiếp tới kết quả |
| Classifier fine-tune | Hợp với nhãn đặc thù dự án; bắt được ý "chủ thể chính" nếu dữ liệu gán nhãn đúng | Cần tập gán nhãn lớn và quy trình train; mỗi lần đổi taxonomy là train lại |

## Câu hỏi mở

- Ngưỡng ứng viên hợp lệ và ngưỡng margin đặt ở bao nhiêu sau khi chạy tập mẫu?
- Mô tả nhãn nên là từ đơn, cụm ngắn hay câu đầy đủ; có cần thử nhiều template rồi lấy trung bình?
- Tag phụ có cần điểm riêng như `alternatives` không?
- Giữ nguyên 8 nhãn hay tách/nhập nhãn nào sau khi chạy tập mẫu?
- Nhóm ảnh giao thoa chiếm bao nhiêu phần trăm tập mẫu để kết luận đủ mạnh?
