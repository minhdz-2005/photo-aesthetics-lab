import styles from './StyleCandidateChip.module.css'

interface StyleCandidateChipProps {
  name: string
  matchLabel: string
  score: number
}

export function StyleCandidateChip({ name, matchLabel, score }: StyleCandidateChipProps) {
  const fillWidth = `${Math.round(score * 100)}%`

  return (
    <span className={styles.chip}>
      <span className={styles.name}>{name}</span>
      <span className={styles.match}>{matchLabel}</span>
      <span className={styles.meterRow}>
        <span className={styles.track} aria-hidden="true">
          <span className={styles.fill} style={{ width: fillWidth }} />
        </span>
        <span className={styles.score}>{score.toFixed(2)}</span>
      </span>
      <span className={styles.caption}>điểm xếp hạng</span>
    </span>
  )
}
