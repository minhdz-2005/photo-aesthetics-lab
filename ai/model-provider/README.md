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
| [genre-classifier](../genre-classifier/README.md) | Chọn một nhãn trong danh sách truyền vào lúc chạy, kèm điểm của vài nhãn đứng đầu |
| [style-analyzer](../style-analyzer/README.md) | Xếp hạng style và nêu đặc tính, trong danh sách được phép theo thể loại |
| [evaluator-workers](../evaluator-workers/README.md) | Đánh giá một tiêu chí trong ngữ cảnh một style profile, nếu tiêu chí đó cần model |

## Quy tắc

- **Chưa tự huấn luyện model.** Dùng model đã huấn luyện sẵn.
- Model phải bị ràng buộc vào danh sách nhãn cố định, không được trả lời mở hay tự đặt tên style mới.
- Điểm model trả về là **điểm xếp hạng** cho tới khi được kiểm chứng; không mặc định coi là xác suất.
- Kết quả không hợp lệ hoặc quá thời hạn là `error`, không đoán thay thế.
- Đổi nhà cung cấp không được làm thay đổi hợp đồng của các bước hay rubric.

## Phương án model

**Bước 1 chốt hướng zero-shot + taxonomy**: mô tả nhãn truyền vào lúc chạy, không nằm trong trọng số model. Lý do và bảng so sánh với classifier fine-tune nằm ở [genre-classifier](../genre-classifier/README.md#phương-pháp-zero-shot--taxonomy).

Hệ quả với lớp này:

- Giao diện **bắt buộc** nhận danh sách ứng viên lúc gọi và trả điểm cho **từng** ứng viên. Nhà cung cấp nào không làm được vậy thì không dùng cho bước 1.
- Giữ được đường cắm classifier fine-tune để chạy so sánh trên cùng tập ảnh, nhưng không vì thế mà đưa quyết định nghiệp vụ vào đây.
- MIT Places2 (tài liệu tham khảo bước 1) là hướng dữ liệu nếu sau này cần fine-tune hoặc đối chiếu nhãn.

## Yêu cầu giao diện

Đi vào:

- Ảnh đã được bước gọi kiểm tra là hợp lệ.
- Danh sách mô tả nhãn đang hoạt động, kèm `taxonomy_version`.
- Context hủy và thời hạn của lượt gọi.

Đi ra:

- Điểm cho từng nhãn ứng viên, theo đúng thứ tự/không tự thêm nhãn ngoài danh sách. **Điều kiện cứng với bước 1:** trả đủ điểm cho **toàn bộ** ứng viên đưa lên, không phải top-k. Thiếu một ứng viên là `error`, vì renormalize trong nhóm cần tổng đủ mọi điểm (spec §3.1). Chưa kiểm chứng bằng chạy thật — xem [integration probe](../../docs/superpowers/plans/2026-10-26-genre-classifier.md#task-14-kiểm-chứng-bằng-một-lượt-gọi-thật).
- Id model và phiên bản đã dùng.
- Log thô của model khi có, phục vụ [evaluation](../evaluation/README.md).

Lớp này **không** đặt ngưỡng, không quyết `other` hay `needs_review`, không diễn giải kết quả. Đó là việc của bước gọi.

## Câu hỏi mở

- Bắt đầu với nhà cung cấp zero-shot nào cho bước 1 và bước 2?
- Một model chung cho mọi bước, hay mỗi bước một model?
- Có lưu log thô của mọi lượt gọi, hay chỉ những lượt rơi vào `needs_review`?
- Chi phí và độ trễ mỗi lượt ảnh có phải giới hạn cần chốt từ đầu?
- Bước 2 dùng model zero-shot chung hay cần model trả được cả đặc tính lẫn style trong một lượt?
