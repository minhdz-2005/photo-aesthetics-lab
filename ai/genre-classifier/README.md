# genre-classifier — Nhận diện thể loại ảnh

Bước 1 trong luồng. Nhận một ảnh và trả thể loại chính để các bước sau chọn đúng bộ tiêu chí.

Chi tiết gốc: [docs/step-01-image-genre-classification.md](../../docs/step-01-image-genre-classification.md)

Trạng thái: **chưa có code**.

## Nhiệm vụ

- Chọn **một thể loại chính** dựa trên chủ thể nổi bật nhất của ảnh.
- Giữ **tag phụ** cho ngữ cảnh khi ảnh thuộc nhiều nhóm.
- Đánh dấu **`needs_review`** khi kết quả mơ hồ thay vì ép chọn.

Bước này **không** đánh giá ảnh đẹp hay chưa, **không** xác định style, **không** chỉnh sửa ảnh.

## Danh sách nhãn

`portrait` · `landscape` · `architecture` · `street` · `food` · `product` · `animal` · `other`

`portrait` nghĩa là ảnh tập trung vào người, không phải ảnh chụp chiều dọc. Đây là taxonomy ban đầu của lab; cần kiểm tra trên ảnh mẫu và điều chỉnh nếu các nhãn hay bị nhầm.

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

`score` là điểm xếp hạng nhãn, không phải xác suất đã hiệu chỉnh. `status` là `classified` hoặc `needs_review`.

## Quy tắc xử lý

1. Kiểm tra ảnh đọc được và có nội dung hình ảnh hợp lệ.
2. Gửi ảnh cùng danh sách nhãn cố định tới bộ phân loại.
3. Chọn nhãn chính theo chủ thể nổi bật nhất.
4. Lưu vài nhãn gần nhất làm thông tin chẩn đoán; chỉ dùng tag phụ khi hữu ích cho bước sau.
5. Mơ hồ thì trả `needs_review`; không tự chuyển sang `other` khi chưa có lý do.
6. Không làm style, điểm thẩm mỹ hay chỉnh sửa ảnh.

## Kiểm chứng

Chạy trên tập ảnh mẫu có nhãn thủ công ở [ai/evaluation](../evaluation/README.md), ưu tiên ảnh giao thoa (chân dung ngoài trời, kiến trúc trong cảnh phố, người trong phong cảnh). Ghi lại: tỷ lệ đúng, cặp nhãn hay nhầm, tỷ lệ cần xem lại, thời gian xử lý và lỗi khi model trả kết quả không hợp lệ.

## Câu hỏi mở

- Ngưỡng nào coi là "các lựa chọn đứng đầu quá gần nhau" để trả `needs_review`?
- Tag phụ có cần điểm riêng như `alternatives` không?
- Giữ nguyên 8 nhãn hay tách/nhập nhãn nào sau khi chạy tập mẫu?
