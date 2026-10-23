import type { CriterionStatus } from '@/types/analysis.ts'
import { verdictFor, type VerdictTone } from '@/lib/verdict.ts'
import styles from './StatusBadge.module.css'

interface StatusBadgeProps {
  status: CriterionStatus
  score: number | null
  scaleMax: number
}

const GLYPH_PATHS: Record<VerdictTone, string> = {
  good: 'M5 13l4 4 10-10',
  fair: 'M5 12h14',
  watch: 'M12 6v7M12 18h.01',
  unknown: 'M9.2 9a3 3 0 1 1 4.2 2.8c-.9.4-1.4 1-1.4 2M12 18h.01',
  'not-applicable': 'M7 7l10 10M17 7L7 17',
  failed: 'M12 5l7.5 14H4.5z',
}

function Glyph({ tone }: { tone: VerdictTone }) {
  return (
    <svg className={styles.glyph} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d={GLYPH_PATHS[tone]} />
    </svg>
  )
}

/** Trạng thái diễn giải gồm cả hình lẫn chữ, không chỉ dựa vào màu. */
export function StatusBadge({ status, score, scaleMax }: StatusBadgeProps) {
  const verdict = verdictFor(status, score, scaleMax)

  return (
    <span className={styles.badge} data-tone={verdict.tone}>
      <Glyph tone={verdict.tone} />
      {verdict.label}
      <span className={styles.raw}>{status}</span>
    </span>
  )
}
