# feature-extraction — Trích đặc trưng thị giác dùng chung

Tính các phép đo thị giác một lần, trước khi giao đánh giá cho worker.

Chi tiết gốc: [docs/step-03-style-criteria-profiles.md](../../docs/step-03-style-criteria-profiles.md), nguyên tắc "Tính các đặc trưng thị giác dùng chung một lần nếu có thể"

Trạng thái: **chưa có code**.

## Nhiệm vụ

- Đọc ảnh và tính các đặc trưng mà nhiều tiêu chí cùng cần.
- Đưa kết quả cho [evaluator-workers](../evaluator-workers/README.md) dưới dạng dữ liệu chỉ đọc.
- Chạy **trước** phần đánh giá, để các worker độc lập có thể chạy song song.

## Ví dụ cần tách riêng

Đường chân trời và góc nghiêng là dữ liệu thô. Tiêu chí `horizon_level` của natural và cinematic dùng chung phép đo này nhưng **chấm khác nhau** vì mục tiêu và trọng số khác nhau. Nếu mỗi worker tự đo lại, cùng một đặc trưng sẽ bị tính nhiều lần và có thể không nhất quán.

## Đầu ra dự kiến

Mỗi phép đo kèm:

- Tên và đơn vị/giá trị thang đo.
- Cách đo hoặc model nào đã sinh ra nó.
- Cờ cho biết phép đo có đủ tin cậy không — ảnh thiếu dữ liệu thì worker phải trả `insufficient_evidence`, không đoán.

## Quy tắc

- **Không chấm điểm và không diễn giải theo style ở đây.** Phần này chỉ đo; việc suy ra tốt/xấu thuộc về worker trong ngữ cảnh một style profile.
- Worker chỉ đọc kết quả trích xuất, không sửa.
- Không có dữ liệu thì báo thiếu rõ ràng; không thay bằng giá trị mặc định trông hợp lý.
- Ảnh gốc luôn được giữ nguyên; mọi phép đo chạy trên bản đọc, không ghi đè file của người dùng.

## Kiểm chứng

- Cùng một ảnh chạy hai lần có ra cùng phép đo không.
- Bao nhiêu phép đo được dùng bởi hơn một tiêu chí — đó là phần đáng tách lên đây.
- Thời gian trích xuất chung so với để từng worker tự đo.

## Câu hỏi mở

- Tính tất cả đặc trưng ngay từ đầu, hay chỉ tính đặc trưng mà rubric đang chọn cần?
- Có cache kết quả theo ảnh + phiên bản bộ đo không?
- Phần nào chạy bằng code đo trực tiếp, phần nào phải gọi model qua [model-provider](../model-provider/README.md)?
