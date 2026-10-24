# style-analyzer — Xác định style của ảnh

Bước 2 trong luồng. Sau khi biết thể loại, ước lượng đặc tính thị giác và một hoặc nhiều style đang thể hiện trong ảnh, để bước 3 chọn rubric phù hợp.

Chi tiết gốc: [docs/step-02-style-identification.md](../../docs/step-02-style-identification.md)

Trạng thái: **chưa có code**.

## Nhiệm vụ

- Nhận diện **đặc tính thị giác** theo danh sách kiểm soát.
- Ánh xạ tổ hợp đặc tính sang **một hoặc nhiều style ứng viên**, có xếp hạng.
- Trả **dấu hiệu hỗ trợ nhận định**, không chỉ trả tên style.

Bước này nhận diện style **đang thể hiện**. Style **mục tiêu** của người dùng là khái niệm riêng, giữ tách biệt và không xử lý ở đây. Không chấm ảnh đẹp/xấu, không chỉnh sửa ảnh.

## Ba loại nhãn cần phân biệt

| Khái niệm | Ví dụ |
| --- | --- |
| Thể loại | phong cảnh, chân dung, kiến trúc, đường phố |
| Style thể hiện | natural, cinematic, film_vintage, minimalist, editorial, monochrome |
| Đặc tính quan sát được | màu ấm/lạnh, tương phản mạnh/nhẹ, ánh sáng sáng/tối, đơn sắc |

Đặc tính là dấu hiệu; style là profile suy ra từ tổ hợp dấu hiệu. Không yêu cầu model tự nghĩ ra tên style mới.

## Taxonomy thử nghiệm

- **Style:** `natural` · `cinematic` · `film_vintage` · `minimalist` · `editorial` · `monochrome`
- **Màu:** `natural` · `warm` · `cool` · `monochrome`
- **Ánh sáng:** `bright_airy` · `balanced` · `moody` · `dramatic`
- **Tương phản:** `low` · `medium` · `high`
- **Độ bão hòa:** `muted` · `natural` · `vivid`

Tất cả chỉ là giả thuyết ban đầu cho lab; cần thử trên ảnh thật và điều chỉnh. Danh sách hợp lệ phụ thuộc thể loại — ví dụ `monochrome` không nhất thiết mở cho mọi thể loại.

## Hợp đồng

**Input:** một ảnh hợp lệ, thể loại chính từ [genre-classifier](../genre-classifier/README.md), danh sách style và đặc tính được phép dùng cho thể loại đó.

**Output:**

```json
{
  "styles": [
    { "name": "cinematic", "score": 0.68 },
    { "name": "natural", "score": 0.54 }
  ],
  "traits": {
    "color": ["cool"],
    "lighting": ["moody"],
    "contrast": "high",
    "saturation": "muted"
  },
  "evidence": ["Tông màu lạnh và tương phản cao"],
  "status": "classified"
}
```

`score` chỉ dùng để xếp hạng, không phải xác suất chính xác. `status` là `classified` hoặc `needs_review`.

## Quy tắc xử lý

1. Nhận thể loại chính từ bước 1.
2. Chọn taxonomy style phù hợp với thể loại đó.
3. Nhận diện đặc tính thị giác theo danh sách cố định.
4. Xếp hạng một hoặc nhiều style từ tổ hợp đặc tính.
5. Trả dấu hiệu hỗ trợ; không chỉ trả tên style.
6. Không chấm đẹp/xấu, không chỉnh sửa.
7. Giữ riêng style hiện có và style mục tiêu của người dùng.

## Kiểm chứng

Nhiều người đánh dấu đặc tính/style trên cùng tập ảnh, cho phép nhiều nhãn. So với **sự đồng thuận giữa người gán nhãn**, không coi một nhãn duy nhất là đáp án đúng. Ghi lại: mức phù hợp của top-k, độ đồng ý với đặc tính AI nêu, tỷ lệ cần xem lại, cặp style hay nhầm, tính nhất quán khi đánh giá lại cùng ảnh.

## Câu hỏi mở

- Style suy ra hoàn toàn từ đặc tính, hay model trả cả hai độc lập rồi so sánh?
- Giữ bao nhiêu style ứng viên khi điểm sát nhau — ngưỡng nào?
- Tổ hợp nhiều style (natural + hơi điện ảnh) biểu diễn thế nào trong hợp đồng?
