/*
 * Hợp đồng dữ liệu, mô phỏng nguyên văn schema trong:
 * docs/step-01-image-genre-classification.md
 * docs/step-02-style-identification.md
 * docs/step-03-style-criteria-profiles.md
 * Tên field giữ snake_case theo JSON đã đặc tả để khi nối backend Go không phải đổi shape.
 * Các field ghi chú "prototype only" chưa thuộc hợp đồng nào.
 */

export type Genre =
  | 'portrait'
  | 'landscape'
  | 'architecture'
  | 'street'
  | 'food'
  | 'product'
  | 'animal'
  | 'other'

export type ClassifierStatus = 'classified' | 'needs_review'

/** score là điểm xếp hạng của model, không phải xác suất đã hiệu chỉnh. */
export interface GenreAlternative {
  genre: Genre
  score: number
}

export interface GenreClassification {
  primary_genre: Genre
  secondary_tags: string[]
  alternatives: GenreAlternative[]
  status: ClassifierStatus
}

export type StyleId =
  | 'natural'
  | 'cinematic'
  | 'film_vintage'
  | 'minimalist'
  | 'editorial'
  | 'monochrome'

export type ColorTrait = 'natural' | 'warm' | 'cool' | 'monochrome'
export type LightingTrait = 'bright_airy' | 'balanced' | 'moody' | 'dramatic'
export type ContrastLevel = 'low' | 'medium' | 'high'
export type SaturationLevel = 'muted' | 'natural' | 'vivid'

/** Đặc tính là dấu hiệu quan sát được; style là profile suy ra từ tổ hợp dấu hiệu. */
export interface VisualTraits {
  color: ColorTrait[]
  lighting: LightingTrait[]
  contrast: ContrastLevel
  saturation: SaturationLevel
}

export interface StyleCandidate {
  name: StyleId
  score: number
}

export interface StyleAnalysis {
  styles: StyleCandidate[]
  traits: VisualTraits
  evidence: string[]
  status: ClassifierStatus
}

export type CriterionGroup = 'technical' | 'composition' | 'style_trait'

/** not_applicable và insufficient_evidence không quy đổi thành điểm 0. */
export type CriterionStatus = 'scored' | 'not_applicable' | 'insufficient_evidence' | 'error'

export type ConfidenceLevel = 'low' | 'medium' | 'high'

/** 'unverified' nghĩa là model tự báo, chưa qua kiểm chứng hay hiệu chỉnh. */
export type ConfidenceBasis = 'unverified' | 'calibrated'

export interface CriterionScale {
  min: number
  max: number
  low_label: string
  mid_label: string
  high_label: string
}

export interface CriterionProfile {
  id: string
  name: string
  description: string
  group: CriterionGroup
  requires: string
  not_applicable_when: string
  scale: CriterionScale
  suggestion_template: string
  profile_version: string
}

export interface RubricCriterion {
  criterion_id: string
  weight: number
}

export interface StyleProfile {
  style_id: StyleId
  genre: Genre
  criteria: RubricCriterion[]
  profile_version: string
  /** prototype only: văn hiển thị cho người dùng */
  name: string
  summary: string
}

export interface RubricSelection {
  profiles: StyleProfile[]
}

export interface CriterionEvaluation {
  criterion_id: string
  style_id: StyleId
  status: CriterionStatus
  score: number | null
  confidence: ConfidenceLevel | null
  confidence_basis: ConfidenceBasis
  evidence: string[]
  suggestion: string | null
  /** prototype only: CSS filter giả lập bản xem trước chỉnh sửa */
  preview_filter?: string
}

export interface StyleEvaluation {
  style_id: StyleId
  evaluations: CriterionEvaluation[]
}
