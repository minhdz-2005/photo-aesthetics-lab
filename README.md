# Photo Aesthetics Lab

## Giới thiệu

Photo Aesthetics Lab là dự án nghiên cứu và học Go thông qua việc xây dựng một trợ lý đánh giá ảnh có giải thích. Hệ thống nhận một ảnh, xác định thể loại và các style thị giác có thể phù hợp, chọn bộ tiêu chí tương ứng, rồi đánh giá từng tiêu chí để đưa ra nhận xét và gợi ý cải thiện.

Dự án không tìm cách định nghĩa một chuẩn “ảnh đẹp” duy nhất. Kết quả được diễn giải theo thể loại, style và tiêu chí đã chọn; người dùng có thể xem bằng chứng, so sánh các style và tự quyết định có áp dụng gợi ý hay không.

## Mục tiêu

- Học Go qua việc thiết kế một hệ thống có nhiều bước xử lý độc lập.
- Nghiên cứu cách phân loại thể loại ảnh và biểu diễn style dưới dạng nhiều đặc tính/nhãn.
- Xây dựng thư viện hồ sơ tiêu chí có thể tái sử dụng giữa nhiều style.
- Đánh giá từng tiêu chí độc lập, có bằng chứng và trạng thái không áp dụng/không đủ dữ liệu.
- Khảo sát liệu chạy evaluator workers đồng thời bằng goroutine có giảm thời gian xử lý hay không.
- Trình bày kết quả dễ hiểu và hỗ trợ chỉnh sửa ảnh ở các giai đoạn tiếp theo.

## Luồng sản phẩm dự kiến

1. Người dùng chọn một ảnh.
2. Hệ thống phân loại thể loại chính và các tag phụ.
3. Hệ thống ước lượng một hoặc nhiều style hiện có cùng các đặc tính thị giác hỗ trợ.
4. Hệ thống chọn rubric cho từng style ứng viên từ thư viện style profiles và criterion profiles.
5. Các evaluator workers chấm từng tiêu chí; các công việc độc lập có thể chạy đồng thời.
6. Hệ thống tổng hợp điểm và bằng chứng theo từng style, rồi trình bày nhận xét và gợi ý chỉnh sửa.
7. Người dùng có thể so sánh kết quả giữa các style; ảnh gốc luôn được giữ nguyên.

## Kiến trúc khái niệm

- **Genre classifier:** trả thể loại chính, tag phụ và trạng thái cần xem lại.
- **Style analyzer:** trả về nhiều style ứng viên cùng đặc tính thị giác và bằng chứng.
- **Criterion registry:** định nghĩa từng tiêu chí, dữ liệu cần, điều kiện áp dụng, thang điểm và phiên bản.
- **Style profiles:** chọn các criterion profiles phù hợp, đặt trọng số và mức mục tiêu theo style.
- **Evaluator workers:** mỗi worker đánh giá một tiêu chí trong ngữ cảnh một style profile, trả điểm, trạng thái, độ chắc chắn, bằng chứng và gợi ý.
- **Aggregator:** tổng hợp kết quả theo từng style; không gộp các style ứng viên thành một điểm duy nhất nếu không có profile pha trộn rõ ràng.
- **Go API/application layer:** điều phối các bước, quản lý tác vụ và trả kết quả có cấu trúc.

Model thị giác là thành phần có thể thay thế. Phần điều phối và rubric nên giữ độc lập với nhà cung cấp model để tiện so sánh phương pháp trong lab.

## Phạm vi MVP

- Một ảnh đầu vào mỗi lượt.
- Thể loại ban đầu: portrait, landscape, architecture, street, food, product, animal và other.
- Cho phép trả về nhiều style ứng viên.
- Bắt đầu với ảnh phong cảnh và một số style như natural, cinematic, film/vintage, minimalist và monochrome.
- Hồ sơ tiêu chí và style profile được định nghĩa rõ, có phiên bản.
- Kết quả giải thích theo từng tiêu chí; trạng thái có thể là scored, not_applicable, insufficient_evidence hoặc error.
- Giao diện prototype dùng dữ liệu mẫu trước khi nối backend.

## Ngoài phạm vi MVP

- Định nghĩa một thang điểm phổ quát cho vẻ đẹp.
- Chụp ảnh hoặc overlay khung hình realtime từ camera.
- Tự động thay đổi nội dung/đối tượng trong ảnh bằng generative editing.
- Tự huấn luyện model từ đầu.
- Tự sinh tiêu chí mới theo từng ảnh mà không qua kiểm chứng.

## Câu hỏi nghiên cứu

- Phân loại một thể loại chính kèm tag phụ có xử lý tốt các ảnh giao thoa không?
- Style nên được nhận diện bằng nhãn đơn hay tổ hợp đặc tính thị giác?
- Người đánh giá có đồng ý với bằng chứng và tiêu chí của từng style profile không?
- Các evaluator workers chạy đồng thời có cải thiện độ trễ so với chạy tuần tự không?
- Giới hạn concurrency ảnh hưởng thế nào đến độ trễ, thông lượng và bộ nhớ?

## Nguyên tắc kết quả

- Hiển thị điểm riêng từng tiêu chí trước điểm tổng.
- Không nhầm điểm model với xác suất đã hiệu chỉnh.
- Giữ trạng thái không áp dụng và không đủ bằng chứng riêng với điểm thấp.
- Cho biết nhận xét dựa trên dấu hiệu nào trong ảnh.
- Tách style hiện có khỏi style mục tiêu người dùng muốn hướng tới.
- Không áp dụng chỉnh sửa nếu người dùng chưa chọn.

## Tài liệu bước

- [Bước 1 — Nhận diện thể loại ảnh](docs/step-01-image-genre-classification.md)
- [Bước 2 — Xác định style của ảnh](docs/step-02-style-identification.md)
- [Bước 3 — Hồ sơ tiêu chí và worker đánh giá](docs/step-03-style-criteria-profiles.md)
- [Đặc tả giao diện](docs/ui-specification.md)