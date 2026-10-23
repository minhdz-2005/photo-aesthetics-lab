import { Button } from '@/components/ui/Button.tsx'
import { Notice } from '@/components/ui/Notice.tsx'
import type { SuggestionItem } from '@/types/preview.ts'
import styles from './SuggestionPanel.module.css'

interface SuggestionPanelProps {
  items: readonly SuggestionItem[]
  activeCriterionId: string | null
  skippedCriterionIds: readonly string[]
  onPreview: (criterionId: string) => void
  onSkip: (criterionId: string) => void
  onRestore: (criterionId: string) => void
}

export function SuggestionPanel({
  items,
  activeCriterionId,
  skippedCriterionIds,
  onPreview,
  onSkip,
  onRestore,
}: SuggestionPanelProps) {
  if (items.length === 0) {
    return (
      <section className={styles.panel} aria-labelledby="suggestions-heading">
        <h2 className="section-title" id="suggestions-heading">
          Gợi ý chỉnh sửa
        </h2>
        <p className={styles.empty}>Không có gợi ý nào cho style này. Ảnh vẫn có thể giữ nguyên.</p>
      </section>
    )
  }

  return (
    <section className={styles.panel} aria-labelledby="suggestions-heading">
      <h2 className="section-title" id="suggestions-heading">
        Gợi ý chỉnh sửa
      </h2>

      <ul className={styles.list}>
        {items.map((item) => {
          const isSkipped = skippedCriterionIds.includes(item.criterionId)
          const isActive = activeCriterionId === item.criterionId
          const canPreview = item.previewFilter !== null

          return (
            <li key={item.criterionId} className={styles.row} data-skipped={isSkipped}>
              <p className={styles.criterion}>{item.criterionName}</p>
              <p className={styles.text}>{item.text}</p>
              {!canPreview && !isSkipped && (
                <p className={styles.unavailable}>Chưa có bản xem trước mô phỏng cho gợi ý này.</p>
              )}
              <div className={styles.actions}>
                {isSkipped ? (
                  <Button variant="quiet" onClick={() => onRestore(item.criterionId)}>
                    Hiện lại
                  </Button>
                ) : (
                  <Button
                    variant="primary"
                    aria-disabled={!canPreview}
                    aria-pressed={isActive}
                    title={canPreview ? undefined : 'Gợi ý này là thay đổi khung hình, chưa mô phỏng được'}
                    onClick={canPreview ? () => onPreview(item.criterionId) : undefined}
                  >
                    {isActive ? 'Đang xem trước' : 'Xem trước'}
                  </Button>
                )}
                {!isSkipped && <Button variant="quiet" onClick={() => onSkip(item.criterionId)}>Bỏ qua</Button>}
              </div>
            </li>
          )
        })}
      </ul>

      <Notice tone="info">
        Ảnh gốc không bị thay đổi. Bản xem trước chỉ là mô phỏng trong prototype và bạn có thể bỏ qua mọi gợi ý.
      </Notice>
    </section>
  )
}
