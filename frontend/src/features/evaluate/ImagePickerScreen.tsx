import { useId, useState } from 'react'
import type { ChangeEvent, DragEvent } from 'react'
import { Button } from '@/components/ui/Button.tsx'
import { Notice } from '@/components/ui/Notice.tsx'
import type { SelectedPhoto } from '@/types/photo.ts'
import styles from './ImagePickerScreen.module.css'

const IMAGE_MIME_PREFIX = 'image/'

interface ImagePickerScreenProps {
  photo: SelectedPhoto | null
  onPick: (file: File) => void
  onClear: () => void
  onUseSample: () => void
  onAnalyze: () => void
}

export function ImagePickerScreen({ photo, onPick, onClear, onUseSample, onAnalyze }: ImagePickerScreenProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [invalidName, setInvalidName] = useState<string | null>(null)
  const inputId = useId()
  const hintId = useId()

  const acceptFiles = (files: FileList | null) => {
    const file = files?.[0]
    if (!file) return

    if (!file.type.startsWith(IMAGE_MIME_PREFIX)) {
      setInvalidName(file.name)
      return
    }

    setInvalidName(null)
    onPick(file)
  }

  const handleInputChange = (event: ChangeEvent<HTMLInputElement>) => {
    acceptFiles(event.target.files)
    event.currentTarget.value = ''
  }

  const handleDragOver = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setIsDragging(true)
  }

  const handleDragLeave = () => setIsDragging(false)

  const handleDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setIsDragging(false)
    acceptFiles(event.dataTransfer.files)
  }

  const handleClear = () => {
    setInvalidName(null)
    onClear()
  }

  return (
    <section className={styles.screen} aria-labelledby="picker-heading">
      <header className={styles.intro}>
        <p className={styles.kicker}>Đánh giá ảnh · prototype</p>
        <h1 id="picker-heading">Khám phá cách bức ảnh của bạn được thể hiện</h1>
        <p className={styles.lede}>Chọn một ảnh để xem phân tích theo thể loại, style và từng tiêu chí cụ thể.</p>
      </header>

      <div
        className={styles.dropzone}
        data-dragging={isDragging}
        data-filled={photo !== null}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <div className={styles.pickerRow}>
          <input
            id={inputId}
            className={styles.fileInput}
            type="file"
            accept="image/*"
            aria-describedby={hintId}
            onChange={handleInputChange}
          />
          <label htmlFor={inputId} className={styles.choose}>
            {photo ? 'Chọn ảnh khác' : 'Chọn ảnh'}
          </label>
          <p id={hintId} className={styles.hint}>
            {photo ? `Đang giữ: ${photo.name}` : 'Kéo thả ảnh vào khung này, hoặc chọn từ máy.'}
          </p>
        </div>

        {photo ? (
          <figure className={styles.preview}>
            <img className={styles.previewImage} src={photo.url} alt={photo.alt} />
            <figcaption className={styles.previewCaption}>
              {photo.isDemo ? 'Ảnh mẫu của prototype' : `Ảnh bạn vừa chọn: ${photo.name}`}
            </figcaption>
          </figure>
        ) : (
          <div className={styles.placeholder} aria-hidden="true">
            <svg viewBox="0 0 48 48" width="44" height="44" fill="none" stroke="currentColor" strokeWidth="1.4">
              <rect x="6" y="10" width="36" height="28" rx="3" />
              <path d="M6 31l11-9 8 7 6-5 11 9" />
              <circle cx="17" cy="19" r="2.6" />
            </svg>
            <span>Chưa có ảnh nào được chọn</span>
          </div>
        )}
      </div>

      {invalidName !== null && (
        <Notice tone="caution">
          “{invalidName}” không phải ảnh hợp lệ. Hãy thử lại với JPG hoặc PNG.
        </Notice>
      )}

      <div className={styles.actions}>
        <Button variant="primary" onClick={onAnalyze} disabled={photo === null}>
          Phân tích ảnh
        </Button>
        {photo !== null && <Button variant="quiet" onClick={handleClear}>Xóa ảnh đã chọn</Button>}
      </div>

      <p className={styles.sampleRow}>
        Chưa có ảnh sẵn?{' '}
        <Button variant="quiet" onClick={onUseSample}>
          Dùng ảnh phong cảnh mẫu
        </Button>
      </p>

      <p className={styles.privacy}>Prototype chỉ giữ ảnh trong lượt phân tích này trên máy bạn; không gửi đi đâu.</p>
    </section>
  )
}
