# Bước 1 — Nhận diện thể loại ảnh

## Mục tiêu

Nhận một ảnh đầu vào và xác định thể loại chính của ảnh để các bước sau chọn đúng bộ tiêu chí đánh giá. Bước này chỉ phân loại thể loại ảnh; chưa đánh giá ảnh đẹp hay chưa và chưa xác định style mục tiêu.

## Định nghĩa thể loại

Dùng một **thể loại chính** để định tuyến sang hồ sơ đánh giá, kèm **tag phụ** để giữ lại ngữ cảnh khi ảnh thuộc nhiều nhóm.

Ví dụ: ảnh có người đứng giữa núi có thể có:

- Thể loại chính: portrait nếu người là chủ thể chính.
- Tag phụ: landscape, outdoor.

Trong dự án này, portrait nghĩa là ảnh tập trung vào người, không có nghĩa là ảnh được chụp theo chiều dọc.

## Danh sách nhãn ban đầu

- portrait — người là chủ thể chính.
- landscape — cảnh thiên nhiên hoặc phong cảnh là chủ thể chính.
- architecture — công trình hoặc kiến trúc là chủ thể chính.
- street — cảnh đường phố hoặc đời sống đô thị là chủ thể chính.
- food — món ăn hoặc đồ uống là chủ thể chính.
- product — sản phẩm là chủ thể chính.
- animal — động vật là chủ thể chính.
- other — không phù hợp với các nhãn trên hoặc chưa xác định được.

Danh sách này là taxonomy ban đầu của lab; cần kiểm tra trên ảnh mẫu và điều chỉnh nếu các nhãn thường xuyên bị nhầm lẫn.

## Phương pháp ban đầu

Dùng một model thị giác đã được huấn luyện sẵn để phân loại ảnh theo danh sách nhãn cố định. Chưa tự huấn luyện model ở bước này.

Yêu cầu model chọn thể loại phù hợp nhất trong danh sách đã định nghĩa, thay vì trả lời mở về mọi nội dung có trong ảnh. Có thể lưu một vài nhãn đứng đầu để nhận biết trường hợp ảnh giao thoa giữa các thể loại.

Điểm do model trả về được xem là điểm xếp hạng cho tới khi được kiểm chứng; không mặc định đó là xác suất chính xác. Nếu model không đủ chắc chắn hoặc các lựa chọn đứng đầu quá gần nhau, kết quả phải có trạng thái needs_review thay vì ép chọn một nhãn.

## Hợp đồng đầu vào/đầu ra

### Input

- Một ảnh.

### Output

    {
      "primary_genre": "landscape",
      "secondary_tags": ["outdoor"],
      "alternatives": [
        {"genre": "architecture", "score": 0.31},
        {"genre": "portrait", "score": 0.12}
      ],
      "status": "classified"
    }

score là điểm model dùng để xếp hạng nhãn, không được hiểu là xác suất đã hiệu chỉnh. status có thể là classified hoặc needs_review.

## Quy tắc xử lý

1. Kiểm tra ảnh có đọc được và có nội dung hình ảnh hợp lệ.
2. Gửi ảnh cùng danh sách nhãn cố định tới bộ phân loại.
3. Chọn nhãn chính dựa trên chủ thể nổi bật nhất của ảnh.
4. Lưu một số nhãn gần nhất làm thông tin chẩn đoán; chỉ dùng tag phụ khi chúng có ích cho các bước sau.
5. Nếu kết quả mơ hồ, trả needs_review; không tự chuyển trường hợp mơ hồ thành other nếu chưa có lý do.
6. Không thực hiện đánh giá style, điểm thẩm mỹ hoặc chỉnh sửa ảnh trong bước này.

## Đánh giá chất lượng bước 1

Tạo một tập ảnh thử nghiệm và gán nhãn thể loại chính thủ công. So sánh dự đoán với nhãn tham chiếu, đặc biệt với các ảnh giao thoa như chân dung ngoài trời, kiến trúc trong cảnh phố và người trong phong cảnh.

Ghi lại:

- Tỷ lệ phân loại đúng.
- Các cặp nhãn thường bị nhầm.
- Tỷ lệ ảnh cần xem lại.
- Thời gian xử lý và lỗi khi model không trả kết quả hợp lệ.

## Tiêu chí hoàn thành

- Có taxonomy ban đầu và định nghĩa rõ từng nhãn.
- Với một ảnh hợp lệ, hệ thống trả thể loại chính theo schema thống nhất.
- Ảnh mơ hồ có thể được đánh dấu needs_review.
- Có tập ảnh mẫu và kết quả so sánh để quyết định có cần đổi taxonomy hoặc model hay không.

## Tài liệu tham khảo

- [Hugging Face — Image classification](https://huggingface.co/docs/transformers/tasks/image_classification)
- [Hugging Face — Zero-shot image classification](https://huggingface.co/tasks/zero-shot-image-classification)
- [MIT Places2 — Scene recognition dataset](https://places2.csail.mit.edu/download.html)