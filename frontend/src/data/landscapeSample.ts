import type { GenreClassification, StyleAnalysis, StyleEvaluation } from '@/types/analysis.ts'

/** Ảnh phong cảnh mẫu dùng cho prototype (docs/ui-specification.md — mục dữ liệu mẫu). */
export const demoPhoto = {
  name: 'ho-nui-sang-som.png',
  url: '/demo/landscape.png',
  alt: 'Hồ núi sáng sớm, dải sương mỏng trên mặt nước và sườn núi phía sau',
}

export const genreSample: GenreClassification = {
  primary_genre: 'landscape',
  secondary_tags: ['outdoor', 'mountain', 'lake'],
  alternatives: [
    { genre: 'architecture', score: 0.31 },
    { genre: 'portrait', score: 0.12 },
  ],
  status: 'classified',
}

export const styleSample: StyleAnalysis = {
  styles: [
    { name: 'natural', score: 0.61 },
    { name: 'cinematic', score: 0.55 },
    { name: 'minimalist', score: 0.34 },
  ],
  traits: {
    color: ['natural', 'cool'],
    lighting: ['balanced', 'moody'],
    contrast: 'medium',
    saturation: 'muted',
  },
  evidence: [
    'Tông màu dịu, hơi thiên lạnh ở vùng bóng',
    'Ánh sáng phân bổ đều, không có vùng cháy rõ',
    'Sương mỏng tạo lớp chuyển tiếp giữa mặt hồ và sườn núi',
  ],
  // natural và cinematic sát nhau nên kết quả để ở trạng thái cần xem lại
  status: 'needs_review',
}

export const evaluationSample: StyleEvaluation[] = [
  {
    style_id: 'natural',
    evaluations: [
      {
        criterion_id: 'horizon_level',
        style_id: 'natural',
        status: 'scored',
        score: 4,
        confidence: 'medium',
        confidence_basis: 'unverified',
        evidence: ['Đường chân trời nhìn thấy được và gần như cân bằng', 'Độ nghiêng ước tính 0.8°'],
        suggestion: null,
      },
      {
        criterion_id: 'exposure_detail',
        style_id: 'natural',
        status: 'scored',
        score: 3,
        confidence: 'medium',
        confidence_basis: 'unverified',
        evidence: [
          'Vùng trời sát ngưỡng cháy ở góc trên bên phải',
          'Vùng tối phía trái vẫn còn chi tiết đá và cây',
        ],
        suggestion: 'Hạ sáng vùng trời khoảng 1/3 stop trước khi xét đẩy tương phản.',
        preview_filter: 'brightness(1.07) contrast(0.93)',
      },
      {
        criterion_id: 'color_naturalness',
        style_id: 'natural',
        status: 'scored',
        score: 4,
        confidence: 'high',
        confidence_basis: 'unverified',
        evidence: ['Cân bằng trắng hơi lạnh so với ánh sáng buổi sớm', 'Kênh xanh lá chưa bị đẩy quá mức'],
        suggestion: 'Hạ bão hòa riêng kênh xanh lá nếu muốn gần màu tự nhiên hơn.',
        preview_filter: 'saturate(0.9)',
      },
      {
        criterion_id: 'subject_environment_balance',
        style_id: 'natural',
        status: 'insufficient_evidence',
        score: null,
        confidence: 'low',
        confidence_basis: 'unverified',
        evidence: ['Chủ thể chính chưa được xác định rõ trong ảnh này'],
        suggestion: null,
      },
    ],
  },
  {
    style_id: 'cinematic',
    evaluations: [
      {
        criterion_id: 'horizon_level',
        style_id: 'cinematic',
        status: 'scored',
        score: 4,
        confidence: 'medium',
        confidence_basis: 'unverified',
        evidence: ['Đường chân trời nhìn thấy được và gần như cân bằng'],
        suggestion: null,
      },
      {
        criterion_id: 'tonal_separation',
        style_id: 'cinematic',
        status: 'scored',
        score: 2,
        confidence: 'medium',
        confidence_basis: 'unverified',
        evidence: ['Mặt hồ và sườn núi gần dính nhau về dải sáng', 'Lớp sương làm giảm ranh giới giữa hai lớp'],
        suggestion: 'Dồn sáng vào dải giữa khung, để hai lớp còn lại chìm xuống.',
        preview_filter: 'contrast(1.2) brightness(0.95)',
      },
      {
        criterion_id: 'color_palette_coherence',
        style_id: 'cinematic',
        status: 'scored',
        score: 3,
        confidence: 'medium',
        confidence_basis: 'unverified',
        evidence: ['Nhóm màu lạnh chiếm ưu thế', 'Một cụm màu ấm nhỏ ở mép trái chưa ăn nhập với phần còn lại'],
        suggestion: 'Thu bảng màu về hai nhóm chính, giữ cụm ấm làm điểm nhấn duy nhất.',
        preview_filter: 'sepia(0.18) hue-rotate(-12deg) saturate(0.85)',
      },
      {
        criterion_id: 'subject_emphasis',
        style_id: 'cinematic',
        status: 'not_applicable',
        score: null,
        confidence: null,
        confidence_basis: 'unverified',
        evidence: ['Không có chủ thể riêng biệt để làm nổi bật trong khung cảnh này'],
        suggestion: null,
      },
    ],
  },
  {
    style_id: 'minimalist',
    evaluations: [
      {
        criterion_id: 'horizon_level',
        style_id: 'minimalist',
        status: 'scored',
        score: 4,
        confidence: 'medium',
        confidence_basis: 'unverified',
        evidence: ['Đường chân trời xuất hiện và gần như cân bằng'],
        suggestion: null,
      },
      {
        criterion_id: 'negative_space',
        style_id: 'minimalist',
        status: 'scored',
        score: 2,
        confidence: 'low',
        confidence_basis: 'unverified',
        evidence: ['Khoảng trời chiếm khoảng 30% khung nhưng bị nhóm cây bên trái cắt ngang'],
        suggestion: 'Cắt bỏ mép dưới bên trái để khoảng trống dẫn mắt vào sườn núi.',
      },
      {
        criterion_id: 'visual_clutter',
        style_id: 'minimalist',
        status: 'scored',
        score: 3,
        confidence: 'medium',
        confidence_basis: 'unverified',
        evidence: ['Đá và cây ở mép dưới là yếu tố phụ chính', 'Mặt nước phía trước khá sạch'],
        suggestion: 'Gở bớt đá ở mép dưới trước khi xử lý màu.',
        preview_filter: 'contrast(1.05) saturate(0.92)',
      },
      {
        criterion_id: 'subject_clarity',
        style_id: 'minimalist',
        status: 'scored',
        score: 3,
        confidence: 'medium',
        confidence_basis: 'unverified',
        evidence: ['Sườn núi đọc được là chủ thể', 'Cạnh sườn chưa tách hẳn khỏi lớp sương'],
        suggestion: 'Tăng tách cạnh cho sườn núi, giảm chi tiết lớp sương phía sau.',
        preview_filter: 'contrast(1.12) brightness(0.98)',
      },
    ],
  },
]
