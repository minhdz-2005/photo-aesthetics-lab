# evaluation — Tập ảnh mẫu và kiểm chứng kết quả

Nơi giữ dữ liệu thử nghiệm và kết quả so sánh cho cả ba bước. Đây là phần trả lời cho các **câu hỏi nghiên cứu** trong [README.md](../../README.md).

Chi tiết gốc: mục "Đánh giá chất lượng" của [bước 1](../../docs/step-01-image-genre-classification.md), [bước 2](../../docs/step-02-style-identification.md), [bước 3](../../docs/step-03-style-criteria-profiles.md)

Trạng thái: **chưa có code, chưa có tập ảnh**.

## Chứa gì

- Tập ảnh mẫu cho từng thể loại, có nhãn tham chiếu do người gán.
- Ảnh giao thoa cố tình đưa vào: chân dung ngoài trời, kiến trúc trong cảnh phố, người trong phong cảnh.
- Kết quả chạy của từng phương án model và từng phiên bản rubric.
- Ghi chú của nhiều người đánh giá về đặc tính, style và tiêu chí.

## Chỉ số cần ghi lại

**Bước 1 — thể loại** ([genre-classifier](../genre-classifier/README.md))

- Tỷ lệ phân loại đúng.
- Các cặp nhãn thường bị nhầm.
- Tỷ lệ ảnh cần xem lại.
- Thời gian xử lý và lỗi khi model không trả kết quả hợp lệ.
- Zero-shot + taxonomy so với classifier fine-tune, chạy trên **cùng tập ảnh**.
- Nhóm riêng cho ảnh giao thoa chủ thể/ngữ cảnh: một người đứng giữa núi phải ra `portrait` theo spec.

**Bước 2 — style** ([style-analyzer](../style-analyzer/README.md))

- Mức phù hợp của các style trong top-k.
- Đặc tính thị giác AI nêu có được người đánh giá đồng ý không.
- Tỷ lệ trường hợp cần xem lại.
- Những cặp style thường bị nhầm.
- Tính nhất quán khi cùng ảnh được đánh giá lại.

**Bước 3 — rubric** ([criterion-registry](../criterion-registry/README.md), [style-profiles](../style-profiles/README.md))

- Tiêu chí mơ hồ, trùng lặp hoặc không quan sát ổn định.
- Tổng trọng số và tính nhất quán trong từng style profile.
- Tiêu chí tái sử dụng có làm mất khác biệt giữa các style không.

**Hiệu năng** ([evaluator-workers](../evaluator-workers/README.md))

- Độ trễ khi chạy tuần tự so với song song.
- Ảnh hưởng của giới hạn concurrency tới độ trễ, thông lượng và bộ nhớ.

## Nguyên tắc

- So với **sự đồng thuận giữa nhiều người gán nhãn**, không coi một nhãn duy nhất là đáp án đúng — nhất là với style.
- Cho phép nhiều nhãn trên cùng một ảnh.
- Điểm model chỉ là xếp hạng cho tới khi được kiểm chứng; muốn gọi là xác suất thì phải hiệu chỉnh và đo.
- Mỗi lần đổi model, taxonomy hoặc phiên bản rubric đều lưu thành một kết quả riêng để so sánh được.
- Kết luận phải dẫn tới việc cần làm: giữ nguyên, đổi taxonomy, hay đổi model.

## Câu hỏi nghiên cứu cần trả lời

1. Phân loại một thể loại chính kèm tag phụ có xử lý tốt ảnh giao thoa không?
2. Style nên nhận diện bằng nhãn đơn hay bằng tổ hợp đặc tính thị giác?
3. Người đánh giá có đồng ý với bằng chứng và tiêu chí của từng style profile không?
4. Evaluator workers chạy đồng thời có cải thiện độ trễ so với tuần tự không?
5. Giới hạn concurrency ảnh hưởng thế nào đến độ trễ, thông lượng và bộ nhớ?
