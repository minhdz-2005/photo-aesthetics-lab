# style-profiles — Rubric theo từng style

Chứa profile của mỗi style: chọn tiêu chí nào từ [criterion-registry](../criterion-registry/README.md), với trọng số và mức mục tiêu nào.

Chi tiết gốc: [docs/step-03-style-criteria-profiles.md](../../docs/step-03-style-criteria-profiles.md), mục "Style profile" và "Ví dụ rubric cho ảnh phong cảnh"

Trạng thái: **chưa có code**.

## Nhiệm vụ

Với một thể loại và một style, trả rubric gồm:

- Thể loại áp dụng.
- Tiêu chí được bật/tắt.
- Trọng số của từng tiêu chí.
- Mức mục tiêu hoặc khoảng chấp nhận được.
- Ưu tiên phong cách dùng để diễn giải kết quả.
- Phiên bản profile.

Một tiêu chí dùng chung có thể xuất hiện ở nhiều profile; trọng số và cách diễn giải có thể khác. Ví dụ `horizon_level` dùng cho cả natural, cinematic và minimalist.

## Hợp đồng

**Input:** thể loại chính từ [genre-classifier](../genre-classifier/README.md), danh sách style ứng viên từ [style-analyzer](../style-analyzer/README.md), thư viện criterion profiles, thư viện style profiles.

**Output:** danh sách rubric đã chọn cho từng style ứng viên — **chưa chứa điểm đánh giá ảnh**.

```json
{
  "profiles": [
    {
      "style_id": "natural",
      "criteria": [
        { "criterion_id": "horizon_level", "weight": 0.2 },
        { "criterion_id": "exposure_detail", "weight": 0.3 },
        { "criterion_id": "color_naturalness", "weight": 0.3 }
      ],
      "profile_version": "1.0"
    }
  ]
}
```

## Rubric khởi đầu cho ảnh phong cảnh

Các mức dưới đây là ví dụ khởi đầu, **cần kiểm chứng với người đánh giá**.

| Tiêu chí | natural | cinematic | minimalist |
| --- | --- | --- | --- |
| Horizon level | trọng số cao, tránh nghiêng ngoài ý muốn | dùng chung, trọng số trung bình | dùng chung nếu có đường chân trời |
| Exposure detail | trọng số cao, giữ chi tiết trời và vùng tối | — | — |
| Color naturalness | trọng số cao, tránh đẩy màu quá mức | — | — |
| Subject/environment balance | trọng số trung bình | — | — |
| Tonal separation | — | trọng số cao, phân tách sáng/tối | — |
| Color palette coherence | — | trọng số cao, bảng màu thống nhất | — |
| Subject emphasis | — | trọng số cao, chủ thể nổi bật | — |
| Negative space | — | — | trọng số cao, khoảng trống hỗ trợ chủ thể |
| Visual clutter | — | — | trọng số cao, yếu tố phụ có gây phân tán |
| Subject clarity | — | — | trọng số cao, chủ thể dễ nhận ra |

## Quy tắc

- Không áp chuẩn thẩm mỹ phổ quát: profile mô tả mục tiêu của chính style đó.
- Bước này chỉ **chọn rubric**; việc chấm ảnh do [evaluator-workers](../evaluator-workers/README.md) thực hiện.
- Đổi trọng số hoặc bật/tắt tiêu chí là tăng `profile_version`.

## Kiểm chứng

- Kiểm tra tổng trọng số và tính nhất quán trong từng profile.
- Nhờ người hiểu nhiếp ảnh xem rubric có hợp với từng style không.
- Xác nhận cùng một criterion profile tái sử dụng được mà không làm mất khác biệt giữa các style.

## Câu hỏi mở

- Trọng số có bắt buộc tổng bằng 1 không, và xử lý sao khi một tiêu chí `not_applicable`?
- Mức mục tiêu biểu diễn bằng thang điểm hay khoảng giá trị đo được?
- Có cho phép profile pha trộn (blend) nhiều style không?
