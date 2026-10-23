import type { StyleProfile } from '@/types/analysis.ts'

/**
 * Style profiles cho thể loại landscape.
 * Cùng một criterion dùng lại được, nhưng trọng số và mục tiêu khác nhau theo style.
 */
export const landscapeStyleProfiles: readonly StyleProfile[] = [
  {
    style_id: 'natural',
    genre: 'landscape',
    name: 'Natural',
    summary: 'Tông màu gần tự nhiên, ánh sáng dịu, ưu tiên giữ chi tiết ở cả trời và vùng tối.',
    profile_version: '1.0',
    criteria: [
      { criterion_id: 'horizon_level', weight: 0.2 },
      { criterion_id: 'exposure_detail', weight: 0.3 },
      { criterion_id: 'color_naturalness', weight: 0.3 },
      { criterion_id: 'subject_environment_balance', weight: 0.2 },
    ],
  },
  {
    style_id: 'cinematic',
    genre: 'landscape',
    name: 'Cinematic',
    summary: 'Tông lạnh, tương phản cao, các dải sáng tách lớp và chủ thể được nhấn.',
    profile_version: '1.0',
    criteria: [
      { criterion_id: 'horizon_level', weight: 0.15 },
      { criterion_id: 'tonal_separation', weight: 0.3 },
      { criterion_id: 'color_palette_coherence', weight: 0.3 },
      { criterion_id: 'subject_emphasis', weight: 0.25 },
    ],
  },
  {
    style_id: 'minimalist',
    genre: 'landscape',
    name: 'Minimalist',
    summary: 'Ít yếu tố, khoảng trống rộng và chủ thể dễ nhận ra ngay từ lần nhìn đầu.',
    profile_version: '1.0',
    criteria: [
      { criterion_id: 'horizon_level', weight: 0.15 },
      { criterion_id: 'negative_space', weight: 0.3 },
      { criterion_id: 'visual_clutter', weight: 0.25 },
      { criterion_id: 'subject_clarity', weight: 0.3 },
    ],
  },
]
