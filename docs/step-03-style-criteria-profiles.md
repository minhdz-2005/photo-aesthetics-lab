# Bước 3 — Hồ sơ tiêu chí và worker đánh giá

## Mục tiêu

Từ thể loại ảnh và danh sách style ứng viên, chọn bộ tiêu chí phù hợp để chuẩn bị cho bước chấm điểm.

Bước này tạo rubric, chưa đánh giá ảnh. Mỗi tiêu chí có một hồ sơ riêng; mỗi lần đánh giá tiêu chí được giao cho một evaluator worker tương ứng.

## Nguyên tắc thiết kế

- Không để AI tự nghĩ ra tiêu chí mới cho mỗi ảnh. Dùng một thư viện hồ sơ tiêu chí đã định nghĩa và kiểm chứng.
- Style profile tham chiếu các tiêu chí trong thư viện, kèm trọng số và mức mục tiêu riêng.
- Một tiêu chí có thể được dùng lại trong nhiều style, nhưng cách chấm phải nhận style profile làm ngữ cảnh.
- Một tiêu chí không áp dụng cho ảnh thì trả trạng thái not_applicable hoặc insufficient_evidence; không quy đổi thành điểm 0.
- Tính các đặc trưng thị giác dùng chung một lần nếu có thể; sau đó giao các đánh giá độc lập cho worker.

## Cấu trúc hồ sơ

### Hồ sơ tiêu chí

Mỗi criterion profile mô tả một tiêu chí và cách đánh giá nó:

- id ổn định, duy nhất.
- tên và mô tả ngắn.
- nhóm: technical, composition hoặc style_trait.
- dấu hiệu/dữ liệu cần để đánh giá.
- điều kiện áp dụng.
- thang điểm và ý nghĩa từng mức.
- trường hợp không đủ bằng chứng.
- mẫu phản hồi/gợi ý chỉnh sửa.
- phiên bản hồ sơ.

Ví dụ tiêu chí: horizon_level.

- Nhóm: technical/composition.
- Dữ liệu cần: đường chân trời nhìn thấy được và góc nghiêng ước tính.
- Không áp dụng: đường chân trời bị che hoặc không có trong ảnh.
- Kết quả: điểm, bằng chứng, độ chắc chắn và trạng thái.

### Style profile

Mỗi style profile tham chiếu các criterion profile và bổ sung:

- thể loại áp dụng.
- tiêu chí được bật/tắt.
- trọng số của từng tiêu chí.
- mức mục tiêu hoặc khoảng chấp nhận được.
- ưu tiên phong cách dùng để diễn giải kết quả.
- phiên bản profile.

Tiêu chí dùng chung có thể xuất hiện ở nhiều profile. Ví dụ, horizon_level có thể được dùng cho natural, cinematic và minimalist, nhưng trọng số hoặc cách diễn giải có thể khác nhau.

## Worker đánh giá tiêu chí

Mỗi evaluator worker nhận:

1. ảnh hoặc dữ liệu thị giác đã trích xuất dùng chung;
2. một criterion profile;
3. một style profile đang được đánh giá;
4. context hủy/thời hạn của lượt đánh giá.

Worker chỉ đánh giá tiêu chí được giao, không tự thay đổi tiêu chí, trọng số hoặc điểm tổng. Kết quả worker trả về gồm:

- criterion_id.
- profile/style_id.
- status: scored, not_applicable, insufficient_evidence hoặc error.
- score theo thang đã định nghĩa.
- confidence hoặc mức chắc chắn đã được hiệu chỉnh/kiểm chứng.
- evidence: dấu hiệu quan sát/đo được.
- suggestion: gợi ý liên quan nếu có.

Các worker có thể chạy song song khi không phụ thuộc dữ liệu của nhau. Phần trích xuất chung nên chạy trước; kết quả được thu và tổng hợp sau khi các worker hoàn tất. Không để các worker cùng sửa một biến điểm tổng.

Nếu một tiêu chí được dùng trong nhiều style profile, có thể dùng chung phép đo thô, nhưng phải chấm riêng theo từng profile vì mục tiêu và trọng số có thể khác nhau.

## Ví dụ rubric cho ảnh phong cảnh

Các mức dưới đây là ví dụ khởi đầu, cần được kiểm chứng với người đánh giá.

### Style natural

- Horizon level: trọng số cao; ưu tiên đường chân trời không bị nghiêng ngoài ý muốn.
- Exposure detail: trọng số cao; giữ thông tin ở vùng trời và vùng tối.
- Color naturalness: trọng số cao; tránh đẩy màu quá mức nếu mục tiêu là tự nhiên.
- Subject/environment balance: trọng số trung bình; giữ cân bằng giữa chủ thể và cảnh.

### Style cinematic

- Horizon level: dùng chung, trọng số trung bình.
- Tonal separation: trọng số cao; đánh giá sự phân tách vùng sáng/tối.
- Color palette coherence: trọng số cao; màu có thống nhất theo profile không.
- Subject emphasis: trọng số cao; chủ thể có nổi bật trong khung không.

### Style minimalist

- Horizon level: dùng chung nếu đường chân trời xuất hiện.
- Negative space: trọng số cao; có khoảng trống hỗ trợ chủ thể không.
- Visual clutter: trọng số cao; yếu tố phụ có làm phân tán không.
- Subject clarity: trọng số cao; chủ thể chính có dễ nhận ra không.

Không coi các nguyên tắc như rule of thirds hoặc tương phản mạnh là điều kiện bắt buộc. Tiêu chí phải mô tả mục tiêu của profile, không áp một chuẩn thẩm mỹ phổ quát.

## Hợp đồng đầu vào/đầu ra của bước 3

### Input

- Thể loại chính từ bước 1.
- Danh sách style ứng viên từ bước 2.
- Thư viện criterion profiles.
- Thư viện style profiles.

### Output

    {
      "profiles": [
        {
          "style_id": "natural",
          "criteria": [
            {"criterion_id": "horizon_level", "weight": 0.2},
            {"criterion_id": "exposure_detail", "weight": 0.3},
            {"criterion_id": "color_naturalness", "weight": 0.3}
          ],
          "profile_version": "1.0"
        }
      ]
    }

Đầu ra là danh sách rubric đã chọn cho từng style ứng viên; chưa chứa điểm đánh giá ảnh.

## Đánh giá chất lượng bước 3

- Kiểm tra mỗi tiêu chí có mô tả, bằng chứng cần thiết và điều kiện áp dụng rõ ràng.
- Nhờ nhiều người có hiểu biết về nhiếp ảnh xem rubric có hợp với từng style không.
- Ghi nhận tiêu chí mơ hồ, trùng lặp hoặc không thể quan sát ổn định.
- Kiểm tra tổng trọng số và tính nhất quán trong từng style profile.
- Xác nhận cùng criterion profile có thể tái sử dụng mà không làm mất khác biệt giữa các style.

## Tiêu chí hoàn thành

- Có thư viện criterion profiles có phiên bản.
- Có style profiles tham chiếu tiêu chí và định nghĩa trọng số/mục tiêu.
- Có hợp đồng đầu vào/đầu ra thống nhất cho evaluator worker.
- Có trạng thái phân biệt giữa điểm thấp, không áp dụng, thiếu bằng chứng và lỗi.
- Bước 3 chỉ chọn rubric; bước 4 mới thực thi các worker để chấm ảnh.

## Tài liệu tham khảo

- [AADB: Photo Aesthetics Ranking Network with Attributes and Content Adaptation](https://arxiv.org/abs/1606.01621)
- [Coarse-to-Fine Image Aesthetics Assessment With Dynamic Attribute Selection](https://ieeexplore.ieee.org/abstract/document/10502339/)
- [Go: Pipelines and cancellation](https://go.dev/blog/pipelines)