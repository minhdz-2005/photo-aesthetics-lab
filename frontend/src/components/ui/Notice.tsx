import type { ReactNode } from 'react'
import styles from './Notice.module.css'

interface NoticeProps {
  tone: 'info' | 'caution'
  children: ReactNode
}

export function Notice({ tone, children }: NoticeProps) {
  const className = tone === 'caution' ? `${styles.notice} ${styles.caution}` : `${styles.notice} ${styles.info}`
  return (
    <p className={className} role="note">
      {children}
    </p>
  )
}
