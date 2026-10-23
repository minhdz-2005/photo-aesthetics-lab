import type {
  ColorTrait,
  ConfidenceLevel,
  ContrastLevel,
  CriterionGroup,
  CriterionStatus,
  Genre,
  LightingTrait,
  SaturationLevel,
  StyleId,
} from '@/types/analysis.ts'

export const genreLabels: Record<Genre, string> = {
  portrait: 'Chân dung',
  landscape: 'Phong cảnh',
  architecture: 'Kiến trúc',
  street: 'Đường phố',
  food: 'Đồ ăn',
  product: 'Sản phẩm',
  animal: 'Động vật',
  other: 'Khác',
}

export const tagLabels: Record<string, string> = {
  outdoor: 'Ngoài trời',
  mountain: 'Núi',
  lake: 'Hồ',
}

export const styleLabels: Record<StyleId, string> = {
  natural: 'Natural',
  cinematic: 'Cinematic',
  film_vintage: 'Film / Vintage',
  minimalist: 'Minimalist',
  editorial: 'Editorial',
  monochrome: 'Monochrome',
}

export const statusLabels: Record<CriterionStatus, string> = {
  scored: 'Đã chấm',
  not_applicable: 'Không áp dụng',
  insufficient_evidence: 'Chưa đủ dữ liệu',
  error: 'Lỗi khi đánh giá',
}

export const groupLabels: Record<CriterionGroup, string> = {
  technical: 'Kỹ thuật',
  composition: 'Bố cục',
  style_trait: 'Đặc tính style',
}

export const confidenceLabels: Record<ConfidenceLevel, string> = {
  low: 'thấp',
  medium: 'trung bình',
  high: 'cao',
}

const traitLabels: Record<ColorTrait | LightingTrait | ContrastLevel | SaturationLevel, string> = {
  natural: 'tự nhiên',
  warm: 'ấm',
  cool: 'lạnh',
  monochrome: 'đơn sắc',
  bright_airy: 'sáng dịu',
  balanced: 'cân bằng',
  moody: 'trầm',
  dramatic: 'kịch tính',
  low: 'thấp',
  medium: 'trung bình',
  high: 'cao',
  muted: 'dịu',
  vivid: 'rực',
}

export function traitLabel(trait: string): string {
  return traitLabels[trait as keyof typeof traitLabels] ?? trait
}

export function tagLabel(tag: string): string {
  return tagLabels[tag] ?? tag
}

/** Ng ngữ diễn tả mức phù hợp mà không khẳng định đó là xác suất. */
export const matchLevelLabels: readonly [string, string, string] = [
  'phù hợp nhất',
  'cũng có dấu hiệu phù hợp',
  'ít phù hợp hơn',
]

export function matchLevelLabel(index: number, total: number): string {
  if (total <= 1) return 'ứng viên duy nhất'
  if (index === 0) return matchLevelLabels[0] ?? ''
  if (index === total - 1) return matchLevelLabels[2] ?? ''
  return matchLevelLabels[1] ?? ''
}
