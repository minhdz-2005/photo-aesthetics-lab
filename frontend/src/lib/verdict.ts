import type { CriterionStatus } from '@/types/analysis.ts'

export type VerdictTone = 'good' | 'fair' | 'watch' | 'unknown' | 'not-applicable' | 'failed'

export interface Verdict {
  tone: VerdictTone
  label: string
}

const GOOD_RATIO = 0.75
const FAIR_RATIO = 0.5

/**
 * Dẫn xuất nhãn diễn giải từ status + điểm của worker.
 * not_applicable và insufficient_evidence không rơi về nhãn điểm thấp.
 */
export function verdictFor(status: CriterionStatus, score: number | null, scaleMax: number): Verdict {
  if (status === 'not_applicable') return { tone: 'not-applicable', label: 'Không áp dụng' }
  if (status === 'insufficient_evidence') return { tone: 'unknown', label: 'Chưa đủ dữ liệu' }
  if (status === 'error') return { tone: 'failed', label: 'Không đánh giá được' }
  if (score === null) return { tone: 'unknown', label: 'Chưa có điểm' }

  const ratio = scaleMax > 0 ? score / scaleMax : 0
  if (ratio >= GOOD_RATIO) return { tone: 'good', label: 'Ổn định' }
  if (ratio >= FAIR_RATIO) return { tone: 'fair', label: 'Tạm ổn' }
  return { tone: 'watch', label: 'Cần chú ý' }
}
