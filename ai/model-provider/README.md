# model-provider — Lớp che model thị giác

Model thị giác là thành phần **có thể thay thế**. Phần điều phối và rubric phải độc lập với nhà cung cấp model để tiện so sánh phương pháp trong lab.

Chi tiết gốc: [README.md](../../README.md) mục "Kiến trúc khái niệm"; [docs/step-01](../../docs/step-01-image-genre-classification.md) và [docs/step-02](../../docs/step-02-style-identification.md), mục "Phương pháp ban đầu"

Trạng thái: **chưa có code**.

## Nhiệm vụ

- Nhận ảnh cùng danh sách nhãn/đặc tính hợp lệ, trả kết quả có cấu trúc theo schema của bước gọi.
- Che khác biệt API giữa các nhà cung cấp model.
- Chuẩn hoá lỗi, thời hạn và trường hợp model trả kết quả không đúng schema.
- Ghi lại model và phiên bản đã dùng, để kết quả sau này giải thích được.

## Ai dùng lớp này

| Nơi gọi | Cần gì |
| --- | --- |
| [genre-classifier](../genre-classifier/README.md) | Chọn một nhãn trong danh sách cố định, kèm vài nhãn đứng đầu |
| [style-analyzer](../style-analyzer/README.md) | Xếp hạng style và nêu đặc tính, trong danh sách được phép theo thể loại |
| [evaluator-workers](../evaluator-workers/README.md) | Đánh giá một tiêu chí trong ngữ cảnh một style profile, nếu tiêu chí đó cần model |

## Quy tắc

- **Chưa tự huấn luyện model.** Dùng model đã huấn luyện sẵn.
- Model phải bị ràng buộc vào danh sách nhãn cố định, không được trả lời mở hay tự đặt tên style mới.
- Điểm model trả về là **điểm xếp hạng** cho tới khi được kiểm chứng; không mặc định coi là xác suất.
- Kết quả không hợp lệ hoặc quá thời hạn là `error`, không đoán thay thế.
- Đổi nhà cung cấp không được làm thay đổi hợp đồng của các bước hay rubric.

## Hướng đang cân nhắc

Tài liệu bước 1 tham khảo image classification, zero-shot image classification và bộ MIT Places2. Lab có thể cần chạy song song hai cách để so:

- Classifier phân loại theo nhãn cố định.
- Zero-shot với danh sách nhãn đưa vào lúc chạy.

## Câu hỏi mở

- Bắt đầu với nhà cung cấp nào cho bước 1 và bước 2?
- Một model chung cho mọi bước, hay mỗi bước một model?
- Kết quả trả về có cần kèm log thô của model để phục vụ [evaluation](../evaluation/README.md) không?
- Chi phí và độ trễ mỗi lượt ảnh có phải giới hạn cần chốt từ đầu?
