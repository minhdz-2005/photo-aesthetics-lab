# Đặc tả giao diện — Photo Aesthetics Lab

## Mục đích

Tạo một prototype web responsive để người dùng hình dung quy trình đánh giá ảnh theo thể loại, nhiều style và các tiêu chí có giải thích. Giai đoạn prototype dùng dữ liệu mẫu; không cần tích hợp model AI hoặc backend thật.

## Người dùng chính

Người muốn hiểu ảnh của mình đang thể hiện theo phong cách nào và nhận gợi ý cải thiện có căn cứ, nhưng vẫn muốn tự quyết định có chỉnh ảnh hay không.

## Nguyên tắc UX

- Không tuyên bố AI quyết định ảnh đẹp hay xấu theo chuẩn tuyệt đối.
- Dùng ngôn ngữ gợi ý, không phán xét.
- Cho xem bằng chứng đằng sau nhận xét.
- Hỗ trợ nhiều style ứng viên cho cùng một ảnh.
- Giữ ảnh gốc rõ ràng; mọi chỉnh sửa chỉ là bản xem trước.
- Ưu tiên giao diện dễ hiểu hơn các thuật ngữ kỹ thuật của model.

## Điều hướng

Thanh điều hướng gọn gồm:

- Tên sản phẩm: Photo Aesthetics Lab.
- Mục “Đánh giá ảnh”.
- Mục “Lịch sử” có thể hiển thị trạng thái mẫu hoặc để disabled trong prototype.
- Nút/biểu tượng cài đặt không cần chức năng ở MVP.

## Màn hình 1 — Chọn ảnh

### Nội dung

- Tiêu đề: “Khám phá cách bức ảnh của bạn được thể hiện”.
- Mô tả: “Chọn ảnh để xem phân tích theo thể loại, style và tiêu chí cụ thể.”
- Khu vực kéo thả ảnh, nút “Chọn ảnh”, và trạng thái khi đã chọn.
- Cho phép thay ảnh hoặc xóa lựa chọn.
- Ghi chú quyền riêng tư ngắn: ảnh chỉ được dùng cho lượt phân tích này trong prototype.

### Hành động

- Chọn ảnh: mở file picker.
- Phân tích ảnh: chuyển sang trạng thái đang phân tích rồi hiển thị trang kết quả mẫu.

## Trạng thái đang phân tích

Hiển thị tiến trình theo các bước dễ hiểu:

1. Đang nhận diện thể loại ảnh.
2. Đang tìm các style phù hợp.
3. Đang áp dụng bộ tiêu chí.
4. Đang tổng hợp nhận xét.

Không cần hiển thị tên model, token hoặc log kỹ thuật. Có thể có nút hủy/quay lại.

## Màn hình 2 — Kết quả đánh giá

### Bố cục desktop

- Khu vực trái hoặc trung tâm: ảnh gốc lớn, có thể phóng to.
- Khu vực phải: kết quả, style candidates và các thẻ tiêu chí.
- Trên cùng: nút quay lại/chọn ảnh khác và trạng thái hoàn thành.

### Bố cục mobile

- Ảnh gốc ở đầu trang.
- Style candidates hiển thị dạng chip có thể cuộn.
- Kết quả tiêu chí xếp dọc dưới ảnh.

### Phần thông tin ảnh

- Thumbnail/ảnh gốc.
- Thể loại chính, ví dụ “Phong cảnh”.
- Tag phụ, ví dụ “Ngoài trời”, “Núi”.
- Nút “Đổi ảnh”.

### Phần style

Hiển thị nhiều style ứng viên dưới dạng thẻ/chip, ví dụ:

- Natural — phù hợp nhất.
- Cinematic — cũng có dấu hiệu phù hợp.
- Minimalist — ít phù hợp hơn.

Mỗi style có chỉ báo mức phù hợp, nhưng không trình bày như xác suất chắc chắn. Khi chọn một style, phần tiêu chí bên dưới chuyển sang rubric của style đó. Có thể có tùy chọn so sánh hai style cạnh nhau.

### Phần tổng quan style đang xem

- Tên style đang chọn.
- Mô tả ngắn về dấu hiệu nhận diện, ví dụ “Tông màu gần tự nhiên, ánh sáng dịu”.
- Trạng thái chắc chắn hoặc “Có thể pha trộn với style khác”.
- Không dùng một điểm tổng lớn duy nhất làm trọng tâm.

### Phần tiêu chí

Hiển thị các thẻ riêng cho từng tiêu chí. Mỗi thẻ gồm:

- Tên tiêu chí, ví dụ “Chi tiết vùng sáng”, “Độ cân bằng”, “Độ tự nhiên của màu”.
- Trạng thái: tốt, cần chú ý, chưa đủ dữ liệu hoặc không áp dụng.
- Điểm/nhãn đánh giá theo rubric, nếu có.
- Một câu bằng chứng cụ thể.
- Gợi ý hành động nhẹ nhàng, nếu phù hợp.
- Nút “Xem bằng chứng” mở thêm chi tiết.

Không coi not_applicable hoặc insufficient_evidence là điểm 0.

### Phần gợi ý chỉnh sửa

- Nhóm gợi ý theo tiêu chí liên quan.
- Mỗi gợi ý có nút “Xem trước” và “Bỏ qua”.
- Prototype có thể dùng slider giả lập hoặc ảnh trước/sau mẫu.
- Luôn có so sánh “Ảnh gốc / Bản xem trước”.
- Không tự ghi đè ảnh gốc.

## Thành phần giao diện

- Upload dropzone.
- Image preview.
- Genre badge và secondary tags.
- Style candidate chips/cards.
- Criterion result cards.
- Confidence/uncertainty badge.
- Evidence disclosure/accordion.
- Before/after comparison.
- Empty, loading, error và needs-review states.

## Dữ liệu mẫu cho prototype

Dùng một ảnh phong cảnh mẫu và dữ liệu phân tích minh họa:

- Thể loại chính: landscape.
- Tag phụ: outdoor, mountain, lake.
- Style ứng viên: natural, cinematic, minimalist.
- Tiêu chí: horizon_level, exposure_detail, color_naturalness, subject_environment_balance.
- Bao gồm ít nhất một tiêu chí insufficient_evidence để thể hiện trạng thái không chắc chắn.

Đánh dấu dữ liệu là mẫu/demo để không gây hiểu nhầm rằng AI đã phân tích thật.

## Trạng thái và lỗi

- Chưa có ảnh: hướng dẫn chọn ảnh.
- Ảnh không hợp lệ: giải thích ngắn và cho thử lại.
- Đang phân tích: tiến trình theo bước.
- Phân loại mơ hồ: cảnh báo nhẹ, vẫn hiển thị các ứng viên.
- Không đủ dữ liệu cho tiêu chí: hiện trạng thái riêng, không coi là lỗi toàn bộ.
- Phân tích lỗi: cho thử lại hoặc chọn ảnh khác.

## Hướng mỹ thuật giao diện

- Phong cách tối giản, hiện đại, tập trung vào ảnh.
- Nền trung tính sáng hoặc xám ấm để ảnh là trọng tâm.
- Typography rõ, phân cấp tiêu đề và nhãn nhất quán.
- Dùng màu trạng thái tiết chế; tránh xanh/đỏ như đúng/sai tuyệt đối.
- Card bo góc vừa phải, khoảng trắng rộng, tương phản đủ đọc.
- Biểu đồ nếu có chỉ dùng thanh mức hoặc thang nhỏ; không dùng radar chart phức tạp.
- Responsive cho desktop và điện thoại.

## Nội dung mẫu

Tiêu đề kết quả: “Bức ảnh có xu hướng Natural”

Mô tả: “Ảnh có màu sắc dịu và ánh sáng khá tự nhiên. Một số tiêu chí bố cục có thể phù hợp với nhiều style.”

Ví dụ thẻ: “Đường chân trời — ổn định. Đường chân trời nhìn thấy được và gần như cân bằng.”

Ví dụ trạng thái không chắc chắn: “Độ cân bằng chủ thể — chưa đủ dữ liệu. Chủ thể chính chưa được xác định rõ trong ảnh này.”

## Tiêu chí hoàn thành prototype

- Người dùng hiểu nơi bắt đầu và chọn được ảnh.
- Người dùng nhìn thấy thể loại và nhiều style ứng viên.
- Chọn style làm thay đổi rubric đang xem.
- Mỗi tiêu chí có trạng thái và bằng chứng.
- Có trạng thái mơ hồ/thiếu dữ liệu.
- Người dùng phân biệt được ảnh gốc với bản xem trước.
- Giao diện dùng tốt ở desktop và mobile.