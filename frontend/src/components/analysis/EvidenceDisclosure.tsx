import { useId, useState } from 'react'
import styles from './EvidenceDisclosure.module.css'

interface EvidenceDisclosureProps {
  items: readonly string[]
  footnote?: string
}

export function EvidenceDisclosure({ items, footnote }: EvidenceDisclosureProps) {
  const [isOpen, setIsOpen] = useState(false)
  const panelId = useId()

  if (items.length === 0) return null

  return (
    <div className={styles.wrap}>
      <button
        type="button"
        className={styles.trigger}
        aria-expanded={isOpen}
        aria-controls={panelId}
        onClick={() => setIsOpen((prev) => !prev)}
      >
        <svg className={styles.chevron} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
          <path d="M5 3l5 5-5 5" />
        </svg>
        {isOpen ? 'Ẩn bằng chứng' : `Xem bằng chứng (${items.length})`}
      </button>
      <div id={panelId} className={styles.panel} hidden={!isOpen}>
        <ul className={styles.list}>
          {items.map((item) => (
            <li key={item} className={styles.item}>
              {item}
            </li>
          ))}
        </ul>
        {footnote && <p className={styles.footnote}>{footnote}</p>}
      </div>
    </div>
  )
}
