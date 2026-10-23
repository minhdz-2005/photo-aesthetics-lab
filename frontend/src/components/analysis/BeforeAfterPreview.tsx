import { useId, useState } from 'react'
import { Notice } from '@/components/ui/Notice.tsx'
import styles from './BeforeAfterPreview.module.css'

interface BeforeAfterPreviewProps {
  photoUrl: string
  alt: string
  previewFilter: string | null
  previewLabel: string | null
}

const FULL_RANGE = 100
const START_REVEAL = 50

export function BeforeAfterPreview({ photoUrl, alt, previewFilter, previewLabel }: BeforeAfterPreviewProps) {
  const [reveal, setReveal] = useState(START_REVEAL)
  const sliderId = useId()

  if (previewFilter === null) {
    return (
      <section className={styles.section} aria-labelledby="compare-heading">
        <h2 className="section-title" id="compare-heading">
          Ảnh gốc / bản xem trước
        </h2>
        <figure className={styles.plain}>
          <img className={styles.image} src={photoUrl} alt={alt} />
          <figcaption className={styles.plainCaption}>Ảnh gốc, chưa áp dụng gợi ý nào.</figcaption>
        </figure>
      </section>
    )
  }

  return (
    <section className={styles.section} aria-labelledby="compare-heading">
      <h2 className="section-title" id="compare-heading">
        Ảnh gốc / bản xem trước
      </h2>

      <div className={styles.frame}>
        <img className={styles.image} src={photoUrl} alt={`${alt} — ảnh gốc`} />
        <img
          className={styles.image}
          src={photoUrl}
          alt=""
          aria-hidden="true"
          style={{ filter: previewFilter, clipPath: `inset(0 0 0 ${reveal}%)` }}
        />
        <span className={styles.handle} style={{ left: `${reveal}%` }} aria-hidden="true" />
        <span className={`${styles.tag} ${styles.tagLeft}`}>Ảnh gốc</span>
        <span className={`${styles.tag} ${styles.tagRight}`}>Bản xem trước</span>
      </div>

      <div className={styles.sliderRow}>
        <label className={styles.sliderLabel} htmlFor={sliderId}>
          Vị trí so sánh
        </label>
        <input
          id={sliderId}
          className={styles.slider}
          type="range"
          min={0}
          max={FULL_RANGE}
          step={1}
          value={reveal}
          aria-valuetext={`${reveal}% ảnh gốc, ${FULL_RANGE - reveal}% bản xem trước`}
          onChange={(event) => setReveal(Number(event.target.value))}
        />
      </div>

      {previewLabel && <p className={styles.caption}>Đang xem trước: {previewLabel}</p>}

      <Notice tone="info">
        Bản xem trước mô phỏng bằng bộ lọc để minh hoạ giao diện, chưa phải ảnh chỉnh sửa thật. Ảnh gốc luôn được giữ nguyên.
      </Notice>
    </section>
  )
}
