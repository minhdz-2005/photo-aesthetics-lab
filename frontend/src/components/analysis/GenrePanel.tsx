import { Notice } from '@/components/ui/Notice.tsx'
import { genreLabels, tagLabel } from '@/lib/labels.ts'
import type { GenreClassification } from '@/types/analysis.ts'
import styles from './GenrePanel.module.css'

interface GenrePanelProps {
  classification: GenreClassification
}

export function GenrePanel({ classification }: GenrePanelProps) {
  const { primary_genre: primary, secondary_tags: tags, alternatives, status } = classification
  const needsReview = status === 'needs_review'

  return (
    <section className={styles.panel} aria-labelledby="genre-heading">
      <h2 className="section-title" id="genre-heading">
        Thể loại nhận diện
      </h2>

      <p className={styles.primary}>{genreLabels[primary]}</p>

      {tags.length > 0 && (
        <ul className={styles.tags} aria-label="Tag phụ">
          {tags.map((tag) => (
            <li key={tag} className={styles.tag}>
              {tagLabel(tag)}
            </li>
          ))}
        </ul>
      )}

      {alternatives.length > 0 && (
        <div className={styles.alternatives}>
          <p className="label">Các thể loại gần nhất</p>
          <ul className={styles.altList}>
            {alternatives.map((alternative) => {
              const barWidth = `${Math.round(alternative.score * 100)}%`
              return (
                <li key={alternative.genre} className={styles.alt}>
                  <span className={styles.altName}>{genreLabels[alternative.genre]}</span>
                  <span className={styles.altTrack} aria-hidden="true">
                    <span className={styles.altFill} style={{ width: barWidth }} />
                  </span>
                  <span className={styles.altScore}>{alternative.score.toFixed(2)}</span>
                </li>
              )
            })}
          </ul>
          <p className={styles.rankingNote}>Điểm do model trả về, chỉ dùng để xếp hạng, không phải xác suất.</p>
        </div>
      )}

      {needsReview && (
        <Notice tone="caution">
          Thể loại chưa rõ ràng. Các ứng viên đứng đầu sát nhau nên hệ thống giữ lại thay vì ép chọn một nhãn.
        </Notice>
      )}
    </section>
  )
}
