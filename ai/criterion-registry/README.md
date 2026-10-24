# criterion-registry — Thư viện hồ sơ tiêu chí

Chứa danh mục tiêu chí đã định nghĩa và kiểm chứng. Người chấm điểm chỉ dùng tiêu chí có trong thư viện này.

Chi tiết gốc: [docs/step-03-style-criteria-profiles.md](../../docs/step-03-style-criteria-profiles.md), mục "Hồ sơ tiêu chí"

Trạng thái: **chưa có code**.

## Nguyên tắc nền

Không để AI tự nghĩ ra tiêu chí mới cho mỗi ảnh. Mọi tiêu chí phải được định nghĩa trước, có phiên bản, và quan sát ổn định.

## Một hồ sơ tiêu chí gồm

- `id` ổn định, duy nhất.
- Tên và mô tả ngắn.
- Nhóm: `technical` · `composition` · `style_trait`.
- Dấu hiệu/dữ liệu cần để đánh giá.
- Điều kiện áp dụng.
- Thang điểm và ý nghĩa từng mức.
- Trường hợp không đủ bằng chứng.
- Mẫu phản hồi/gợi ý chỉnh sửa.
- Phiên bản hồ sơ.

## Ví dụ: `horizon_level`

- Nhóm: technical/composition.
- Dữ liệu cần: đường chân trời nhìn thấy được và góc nghiêng ước tính.
- Không áp dụng: đường chân trời bị che hoặc không có trong ảnh.
- Kết quả: điểm, bằng chứng, độ chắc chắn và trạng thái.

## Quy tắc

- Một tiêu chí có thể dùng lại trong nhiều style, nhưng **cách chấm phải nhận style profile làm ngữ cảnh** — cùng tiêu chí, mục tiêu và trọng số khác nhau.
- Không áp dụng thì trả `not_applicable` hoặc `insufficient_evidence`; **không quy đổi thành điểm 0**.
- Tiêu chí mô tả mục tiêu của profile, không áp một chuẩn thẩm mỹ phổ quát. Rule of thirds hay tương phản mạnh không phải điều kiện bắt buộc.
- Sửa nội dung diễn giải hoặc thang điểm là tăng phiên bản hồ sơ, để kết quả cũ vẫn giải thích được.

## Tổ chức file dự kiến

Ba nhóm `technical` / `composition` / `style_trait` có thể tách thành thư mục con khi số tiêu chí tăng. Chưa chốt cho tới khi có hồ sơ đầu tiên.

## Kiểm chứng

- Mỗi tiêu chí phải có mô tả, bằng chứng cần thiết và điều kiện áp dụng rõ ràng.
- Nhờ nhiều người hiểu về nhiếp ảnh xem tiêu chí có quan sát được và có nghĩa không.
- Ghi nhận tiêu chí mơ hồ, trùng lặp, hoặc không thể quan sát ổn định.

## Câu hỏi mở

- Hồ sơ để dạng file khai báo (JSON/YAML) hay định nghĩa trong code?
- Phiên bản hồ sơ và phiên bản style profile chốt với nhau thế nào?
- Bắt đầu với bao nhiêu tiêu chí cho phong cảnh?
