import { EvidenceDisclosure } from '@/components/analysis/EvidenceDisclosure.tsx'
import { Notice } from '@/components/ui/Notice.tsx'
import { styleLabels, traitLabel } from '@/lib/labels.ts'
import type { StyleAnalysis, StyleProfile } from '@/types/analysis.ts'
import styles from './StyleOverview.module.css'

interface StyleOverviewProps {
  profile: StyleProfile
  analysis: StyleAnalysis
}

export function StyleOverview({ profile, analysis }: StyleOverviewProps) {
  const { traits } = analysis
  const traitRows = [
    { term: 'Màu', values: traits.color },
    { term: 'Ánh sáng', values: traits.lighting },
    { term: 'Tương phản', values: [traits.contrast] },
    { term: 'Độ bão hòa', values: [traits.saturation] },
  ]
  const mixableNotice = analysis.status === 'needs_review'

  return (
    <section className={styles.overview} aria-labelledby="style-overview-heading">
      <h2 className={styles.title} id="style-overview-heading">
        {styleLabels[profile.style_id]}
      </h2>
      <p className={styles.summary}>{profile.summary}</p>

      {mixableNotice && (
        <Notice tone="caution">
          Có thể pha trộn với style khác: hai ứng viên đứng đầu có điểm xếp hạng sát nhau nên chưa chốt một style duy nhất.
        </Notice>
      )}

      <dl className={styles.traits}>
        {traitRows.map((row) => (
          <div key={row.term} className={styles.traitRow}>
            <dt className={styles.traitTerm}>{row.term}</dt>
            <dd className={styles.traitValue}>{row.values.map(traitLabel).join(', ')}</dd>
          </div>
        ))}
      </dl>

      <EvidenceDisclosure items={analysis.evidence} footnote="Dấu hiệu thị giác dùng để suy ra style, không phải điểm thẩm mỹ." />

      <p className={styles.meta}>
        Rubric đang xem: {profile.criteria.length} tiêu chí · hồ sơ {profile.genre} v{profile.profile_version}
      </p>
    </section>
  )
}
