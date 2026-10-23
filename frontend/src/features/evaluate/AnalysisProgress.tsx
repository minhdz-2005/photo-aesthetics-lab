import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/Button.tsx'
import type { SelectedPhoto } from '@/types/photo.ts'
import styles from './AnalysisProgress.module.css'

/** prototype only: nhịp tiến trình mô phỏng theo docs/ui-specification.md, chưa gọi backend. */
const ANALYSIS_STEPS = [
  { id: 'genre', label: 'Đang nhận diện thể loại ảnh', durationMs: 900 },
  { id: 'style', label: 'Đang tìm các style phù hợp', durationMs: 1100 },
  { id: 'rubric', label: 'Đang áp dụng bộ tiêu chí', durationMs: 900 },
  { id: 'summary', label: 'Đang tổng hợp nhận xét', durationMs: 800 },
]

const HANDOFF_MS = 500

const STATE_LABELS = {
  done: 'Hoàn tất',
  active: 'Đang thực hiện',
  waiting: 'Chờ',
} as const

type StepState = keyof typeof STATE_LABELS

interface AnalysisProgressProps {
  photo: SelectedPhoto
  onCancel: () => void
  onComplete: () => void
}

export function AnalysisProgress({ photo, onCancel, onComplete }: AnalysisProgressProps) {
  const [stepIndex, setStepIndex] = useState(0)

  useEffect(() => {
    const step = ANALYSIS_STEPS[stepIndex]

    if (!step) {
      const handoff = window.setTimeout(onComplete, HANDOFF_MS)
      return () => window.clearTimeout(handoff)
    }

    const timer = window.setTimeout(() => setStepIndex((index) => index + 1), step.durationMs)
    return () => window.clearTimeout(timer)
  }, [stepIndex, onComplete])

  const current = ANALYSIS_STEPS[stepIndex]

  return (
    <section className={styles.screen} aria-labelledby="progress-heading">
      <header className={styles.intro}>
        <h1 id="progress-heading">Đang phân tích ảnh</h1>
        <p className={styles.photoNote}>{photo.name}</p>
      </header>

      <div className={styles.body}>
        <img className={styles.thumb} src={photo.url} alt="" aria-hidden="true" />

        <ol className={styles.steps}>
          {ANALYSIS_STEPS.map((step, index) => {
            const state: StepState = index < stepIndex ? 'done' : index === stepIndex ? 'active' : 'waiting'
            return (
              <li
                key={step.id}
                className={styles.step}
                data-state={state}
                aria-current={state === 'active' ? 'step' : undefined}
              >
                <span className={styles.marker} aria-hidden="true" />
                <span className={styles.stepLabel}>{step.label}</span>
                <span className={styles.stepState}>{STATE_LABELS[state]}</span>
              </li>
            )
          })}
        </ol>

        <p className={styles.live} role="status" aria-live="polite">
          {current ? current.label : 'Đang hoàn tất…'}
        </p>
      </div>

      <Button variant="ghost" onClick={onCancel}>
        Hủy và chọn ảnh khác
      </Button>
    </section>
  )
}
