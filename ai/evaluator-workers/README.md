# evaluator-workers — Worker đánh giá một tiêu chí

Mỗi worker chấm **một tiêu chí** trong ngữ cảnh **một style profile**. Một lượt đánh giá gồm nhiều worker chạy độc lập.

Chi tiết gốc: [docs/step-03-style-criteria-profiles.md](../../docs/step-03-style-criteria-profiles.md), mục "Worker đánh giá tiêu chí"

Trạng thái: **chưa có code**.

## Worker nhận vào

1. Ảnh hoặc dữ liệu thị giác đã trích xuất dùng chung từ [feature-extraction](../feature-extraction/README.md).
2. Một criterion profile từ [criterion-registry](../criterion-registry/README.md).
3. Một style profile đang được đánh giá từ [style-profiles](../style-profiles/README.md).
4. Context hủy và thời hạn của lượt đánh giá.

## Worker trả về

| Trường | Nội dung |
| --- | --- |
| `criterion_id` | Tiêu chí vừa đánh giá |
| `style_id` | Profile/style làm ngữ cảnh chấm |
| `status` | `scored` · `not_applicable` · `insufficient_evidence` · `error` |
| `score` | Điểm theo thang đã định nghĩa trong hồ sơ tiêu chí |
| `confidence` | Độ chắc chắn, kèm thông tin đã hiệu chỉnh/kiểm chứng hay chưa |
| `evidence` | Dấu hiệu quan sát hoặc đo được trong ảnh |
| `suggestion` | Gợi ý liên quan, nếu có |

## Bốn trạng thái, không gộp lại

- `scored` — đã chấm, có điểm và bằng chứng.
- `not_applicable` — tiêu chí không áp dụng cho ảnh này. Ví dụ `horizon_level` khi ảnh không có đường chân trời.
- `insufficient_evidence` — có thể áp dụng nhưng dữ liệu không đủ để kết luận.
- `error` — quá trình chấm thất bại.

Ba trạng thái sau **không phải điểm 0**. Hiển thị điểm thấp khác với "không đánh giá được", và người xem phải phân biệt được.

## Quy tắc

- Worker chỉ đánh giá tiêu chí được giao. **Không tự đổi tiêu chí, trọng số hay điểm tổng.**
- Các worker không phụ thuộc dữ liệu của nhau thì chạy song song.
- Không để nhiều worker cùng sửa một biến điểm tổng — chỉ [aggregator](../aggregator/README.md) được tổng hợp.
- Cùng một tiêu chí dùng cho nhiều style thì chung phép đo thô nhưng **chấm riêng** theo từng profile.
- Phải tôn trọng context hủy/thời hạn: lượt đánh giá bị hủy thì dừng và dọn việc đang làm.
- Luôn có `evidence`; một điểm không kèm dấu hiệu quan sát được là kết quả không đạt.

## Kiểm chứng

- Kết quả có đúng schema và đúng một trong bốn trạng thái không.
- Cùng ảnh, cùng profile, chạy lại có cho kết quả nhất quán không.
- Chạy song song có cải thiện độ trễ so với tuần tự không; giới hạn concurrency ảnh hưởng thế nào tới độ trễ, thông lượng và bộ nhớ.

## Câu hỏi mở

- Một worker gọi model hay chỉ đọc dữ liệu đã trích xuất?
- `error` có retry không, và retry mấy lần trong thời hạn của lượt?
- Giới hạn concurrency đặt ở mức nào khi số tiêu chí × số style tăng?
