# Bước 2 — Xác định style của ảnh

## Mục tiêu

Sau khi đã biết thể loại ảnh, ước lượng các đặc tính thị giác và style hiện có để chọn bộ tiêu chí đánh giá phù hợp.

Bước này nhận diện style đang thể hiện trong ảnh. Nó không xác định style người dùng muốn ảnh đạt tới; style mục tiêu là một khái niệm riêng, có thể dùng ở bước chỉnh sửa.

## Phân biệt khái niệm

- **Thể loại:** ảnh phong cảnh, chân dung, kiến trúc, đường phố, đồ ăn, sản phẩm, động vật...
- **Style thể hiện:** tự nhiên, điện ảnh, film/vintage, tối giản, editorial...
- **Đặc tính quan sát được:** màu ấm/lạnh, tương phản mạnh/nhẹ, ánh sáng sáng/tối, đơn sắc...

Không yêu cầu model tự nghĩ ra một tên style. Trước hết nhận diện đặc tính theo danh sách kiểm soát, sau đó ánh xạ các đặc tính đó sang một hoặc nhiều profile style đã định nghĩa.

## Phương pháp ban đầu

Dùng model thị giác có sẵn. Cung cấp ảnh, thể loại đã nhận diện và danh sách đặc tính/style hợp lệ. Yêu cầu model xếp hạng các style phù hợp, đồng thời trả về các dấu hiệu thị giác hỗ trợ nhận định.

Kết quả được biểu diễn theo nhiều nhãn, không bắt buộc chỉ có một style. Ví dụ một ảnh có thể đồng thời mang đặc tính tự nhiên, màu lạnh và hơi điện ảnh.

Điểm model trả về chỉ dùng để xếp hạng cho tới khi được kiểm chứng; không mặc định đó là xác suất đã hiệu chỉnh. Nếu các ứng viên gần nhau hoặc dấu hiệu không rõ, trả trạng thái needs_review hoặc giữ nhiều style ứng viên.

## Taxonomy thử nghiệm

Các nhóm ban đầu dưới đây chỉ là giả thuyết cho lab; cần thử trên ảnh thật và điều chỉnh.

### Style/profile

- natural — cách thể hiện gần tự nhiên.
- cinematic — màu sắc và tương phản có cảm giác điện ảnh.
- film_vintage — dấu hiệu màu hoặc chất ảnh gợi phong cách film/vintage.
- minimalist — hình ảnh ít yếu tố, có khoảng trống đáng kể.
- editorial — cách thể hiện giống ảnh biên tập/tạp chí.
- monochrome — ảnh đơn sắc hoặc gần đơn sắc.

### Đặc tính thị giác

- Màu: natural, warm, cool, monochrome.
- Ánh sáng: bright_airy, balanced, moody, dramatic.
- Tương phản: low, medium, high.
- Độ bão hòa: muted, natural, vivid.

Style và đặc tính thị giác không phải cùng một loại nhãn. Đặc tính là dấu hiệu; style là profile được suy ra từ một tổ hợp dấu hiệu.

## Hợp đồng đầu vào/đầu ra

### Input

- Một ảnh hợp lệ.
- Thể loại chính từ bước 1.
- Danh sách style và đặc tính được phép dùng cho thể loại đó.

### Output

    {
      "styles": [
        {"name": "cinematic", "score": 0.68},
        {"name": "natural", "score": 0.54}
      ],
      "traits": {
        "color": ["cool"],
        "lighting": ["moody"],
        "contrast": "high",
        "saturation": "muted"
      },
      "evidence": [
        "Tông màu lạnh và tương phản cao"
      ],
      "status": "classified"
    }

score chỉ là điểm xếp hạng của model, không được hiểu là xác suất chính xác. status có thể là classified hoặc needs_review.

## Quy tắc xử lý

1. Nhận thể loại chính từ bước 1.
2. Chọn taxonomy style phù hợp với thể loại đó.
3. Nhận diện các đặc tính thị giác theo danh sách cố định.
4. Xếp hạng một hoặc nhiều style phù hợp từ tổ hợp đặc tính.
5. Trả lại dấu hiệu hỗ trợ nhận định; không chỉ trả tên style.
6. Không chấm ảnh đẹp/xấu và không chỉnh sửa ảnh ở bước này.
7. Giữ riêng style hiện có và style mục tiêu của người dùng.

## Đánh giá chất lượng bước 2

Tạo tập ảnh mẫu và nhờ nhiều người đánh dấu các đặc tính/style phù hợp. Cho phép nhiều nhãn trên cùng ảnh. So sánh kết quả AI với sự đồng thuận giữa người gán nhãn, thay vì chỉ coi một nhãn duy nhất là đáp án đúng.

Ghi lại:

- Mức phù hợp của các style trong top-k.
- Các đặc tính thị giác AI nêu có được người đánh giá đồng ý không.
- Tỷ lệ trường hợp cần xem lại.
- Những cặp style thường bị nhầm.
- Tính nhất quán khi cùng ảnh được đánh giá lại.

## Tiêu chí hoàn thành

- Có taxonomy style và đặc tính thị giác ban đầu.
- Một ảnh có thể trả về nhiều style ứng viên.
- Mỗi style có dấu hiệu giải thích được.
- Trường hợp mơ hồ được giữ lại thay vì ép thành một nhãn.
- Có tập ảnh thử nghiệm và nhận xét của nhiều người để điều chỉnh taxonomy.

## Tài liệu tham khảo

- [AADB: Photo Aesthetics Ranking Network with Attributes and Content Adaptation](https://arxiv.org/abs/1606.01621)
- [Coarse-to-Fine Image Aesthetics Assessment With Dynamic Attribute Selection](https://ieeexplore.ieee.org/abstract/document/10502339/)