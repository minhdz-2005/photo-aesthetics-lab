# aggregator — Tổng hợp kết quả theo từng style

Thu kết quả từ các [evaluator-workers](../evaluator-workers/README.md) và dựng thành một bản đánh giá hoàn chỉnh **cho từng style ứng viên**.

Chi tiết gốc: [README.md](../../README.md) mục "Kiến trúc khái niệm" và "Nguyên tắc kết quả"; [docs/ui-specification.md](../../docs/ui-specification.md)

Trạng thái: **chưa có code**.

## Nhiệm vụ

- Gom toàn bộ kết quả worker theo cặp (style, tiêu chí).
- Áp trọng số từ style profile để tính điểm tổng **riêng cho từng style**.
- Giữ bằng chứng và gợi ý của từng tiêu chí đi kèm.
- Đánh dấu style có nhiều tiêu chí không đủ bằng chứng, hoặc hai ứng viên điểm sát nhau.

## Hợp đồng đầu ra (dự kiến)

Mỗi style một mục, gồm:

- `style_id` và `profile_version` đã dùng.
- Điểm từng tiêu chí: `criterion_id`, `status`, `score`, `weight`, `confidence`, `evidence`, `suggestion`.
- Điểm tổng có trọng số của style đó, kèm nhãn nói rõ đây là **điểm xếp hạng**.
- Trạng thái tổng của lượt: đủ căn cứ, hay cần người dùng xem lại.

Kết quả cuối cùng được chuyển cho `frontend/` hiển thị; giao diện đang dùng dữ liệu mẫu cùng hình dạng.

## Quy tắc

- **Điểm riêng từng tiêu chí hiển thị trước điểm tổng.**
- **Không gộp các style ứng viên thành một điểm duy nhất** nếu không có profile pha trộn rõ ràng.
- Không nhầm điểm model với xác suất đã hiệu chỉnh.
- `not_applicable`, `insufficient_evidence` và `error` giữ nguyên trạng thái; không tính thành 0 và không âm thầm loại khỏi kết quả.
- Không để worker sửa điểm tổng; chỉ thành phần này được tổng hợp.
- Nhận xét phải nói được nó dựa trên dấu hiệu nào trong ảnh.
- Không áp dụng chỉnh sửa khi người dùng chưa chọn.

## Kiểm chứng

- Cùng bộ kết quả worker có luôn ra cùng điểm tổng không (không phụ thuộc thứ tự worker hoàn thành).
- Điểm tổng có phản ánh đúng trọng số của profile không khi một số tiêu chí bị loại vì không áp dụng.
- Hai style ứng viên điểm sát nhau có được đánh dấu để người dùng tự so sánh không.

## Câu hỏi mở

- Khi một phần tiêu chí không `scored`, điểm tổng tính lại trên phần còn lại hay hiển thị "chưa đủ căn cứ"?
- Có cần giữ kết quả của mọi style ứng viên trong phản hồi, hay chỉ style đang chọn?
- Điểm tổng giữa các style có so sánh trực tiếp được không, khi rubric và trọng số khác nhau?
