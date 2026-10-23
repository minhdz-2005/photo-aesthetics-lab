import { EvidenceDisclosure } from '@/components/analysis/EvidenceDisclosure.tsx'
import { StatusBadge } from '@/components/ui/StatusBadge.tsx'
import { confidenceLabels, groupLabels } from '@/lib/labels.ts'
import { verdictFor } from '@/lib/verdict.ts'
import type { CriterionEvaluation, CriterionProfile } from '@/types/analysis.ts'
import styles from './CriterionCard.module.css'

interface CriterionCardProps {
  profile: CriterionProfile
  evaluation: CriterionEvaluation
  weight: number
}

export function CriterionCard({ profile, evaluation, weight }: CriterionCardProps) {
  const { scale } = profile
  const { score, status, confidence, confidence_basis: basis } = evaluation

  const verdict = verdictFor(status, score, scale.max)
  const segments = Array.from({ length: scale.max }, (_, index) => index + 1)
  const weightPercent = `${Math.round(weight * 100)}%`
  const anchorText = `${scale.low_label} — ${scale.mid_label} — ${scale.high_label}`
  const confidenceText = confidence === null ? 'không báo cáo' : confidenceLabels[confidence]
  const basisText = basis === 'unverified' ? 'chưa hiệu chỉnh' : 'đã hiệu chỉnh'

  const conditionNote =
    status === 'not_applicable'
      ? `Không áp dụng khi: ${profile.not_applicable_when}`
      : `Cần có để chấm: ${profile.requires}`

  return (
    <article className={styles.card} data-tone={verdict.tone}>
      <div className={styles.head}>
        <h3 className={styles.name}>{profile.name}</h3>
        <StatusBadge status={status} score={score} scaleMax={scale.max} />
      </div>

      <p className={styles.meta}>
        <span className="label">{groupLabels[profile.group]}</span>
        <span className={styles.separator} aria-hidden="true">
          ·
        </span>
        <span>trọng số {weightPercent}</span>
        <span className={styles.separator} aria-hidden="true">
          ·
        </span>
        <span className={styles.version}>v{profile.profile_version}</span>
      </p>

      <p className={styles.description}>{profile.description}</p>

      {status === 'scored' && score !== null ? (
        <div className={styles.scoreRow}>
          <div
            className={styles.meter}
            role="meter"
            aria-valuemin={scale.min}
            aria-valuemax={scale.max}
            aria-valuenow={score}
            aria-valuetext={`Điểm ${score} trên ${scale.max}`}
          >
            {segments.map((segment) => (
              <span key={segment} className={styles.segment} data-filled={segment <= score} />
            ))}
          </div>
          <p className={styles.anchors}>{anchorText}</p>
        </div>
      ) : (
        <p className={styles.condition}>{conditionNote}</p>
      )}

      {evaluation.suggestion && (
        <p className={styles.suggestion}>
          <span className="label">Gợi ý</span>
          {evaluation.suggestion}
        </p>
      )}

      <EvidenceDisclosure items={evaluation.evidence} footnote="Dấu hiệu quan sát trên ảnh mẫu, không phải kết luận về chất lượng ảnh." />

      <p className={styles.confidence}>
        Độ chắc chắn: {confidenceText} · {basisText}
      </p>
    </article>
  )
}
