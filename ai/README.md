# ai/ — Các thành phần AI

Chứa phần nhận diện và đánh giá ảnh, tách khỏi `frontend/`. Mỗi thư mục con tương ứng một thành phần trong mục **Kiến trúc khái niệm** của [README.md](../README.md) và được định nghĩa chi tiết trong các file bước ở [docs/](../docs).

Trạng thái: **chưa có code**. Các file trong này mới chốt phạm vi, hợp đồng dữ liệu và quy tắc; phần triển khai sẽ làm sau.

## Luồng xử lý

| Bước | Thành phần | Thư mục | Tài liệu gốc |
| --- | --- | --- | --- |
| 2 | Nhận thể loại chính và tag phụ | [genre-classifier](genre-classifier/README.md) | [Bước 1](../docs/step-01-image-genre-classification.md) |
| 3 | Ước lượng đặc tính thị giác và style ứng viên | [style-analyzer](style-analyzer/README.md) | [Bước 2](../docs/step-02-style-identification.md) |
| 4 | Chọn rubric: tiêu chí, trọng số, mức mục tiêu | [criterion-registry](criterion-registry/README.md), [style-profiles](style-profiles/README.md) | [Bước 3](../docs/step-03-style-criteria-profiles.md) |
| 5 | Trích đặc trưng dùng chung một lần | [feature-extraction](feature-extraction/README.md) | Bước 3, nguyên tắc thiết kế |
| 5 | Chấm từng tiêu chí trong ngữ cảnh một style | [evaluator-workers](evaluator-workers/README.md) | Bước 3 |
| 6 | Tổng hợp kết quả theo từng style | [aggregator](aggregator/README.md) | README, luồng sản phẩm |
| — | Lớp che model thị giác để đổi nhà cung cấp | [model-provider](model-provider/README.md) | README, kiến trúc khái niệm |
| — | Tập ảnh mẫu, nhãn tham chiếu, so sánh kết quả | [evaluation](evaluation/README.md) | Mục "Đánh giá chất lượng" của cả ba bước |

## Nguyên tắc áp dụng cho toàn bộ thư mục

- Điểm model trả về là **điểm xếp hạng**, không mặc định là xác suất đã hiệu chỉnh.
- Kết quả mơ hồ trả `needs_review` hoặc giữ nhiều ứng viên; không ép thành một nhãn.
- `not_applicable` và `insufficient_evidence` giữ riêng, không quy đổi thành điểm 0.
- Không để model tự nghĩ ra tiêu chí mới; chỉ dùng thư viện hồ sơ đã kiểm chứng.
- Tách style đang thể hiện khỏi style mục tiêu người dùng muốn hướng tới.
- Ảnh gốc luôn được giữ nguyên; không chỉnh sửa khi người dùng chưa chọn.
- Phần điều phối và rubric phải độc lập với nhà cung cấp model.
