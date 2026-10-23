import styles from './AppHeader.module.css'

export function AppHeader() {
  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        <p className={styles.brand}>
          <span className={styles.brandName}>Photo Aesthetics Lab</span>
          <span className={styles.brandNote}>prototype</span>
        </p>
        <nav className={styles.nav} aria-label="Điều hướng chính">
          <span className={styles.navCurrent} aria-current="page">
            Đánh giá ảnh
          </span>
          <button type="button" className={styles.navItem} aria-disabled="true" title="Chưa có trong prototype">
            Lịch sử
          </button>
          <button type="button" className={styles.iconButton} aria-label="Cài đặt, chưa có trong prototype">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.6" aria-hidden="true">
              <path d="M4 7h8M18 7h2M4 17h2M12 17h8" />
              <circle cx="15" cy="7" r="2.4" />
              <circle cx="9" cy="17" r="2.4" />
            </svg>
          </button>
        </nav>
      </div>
    </header>
  )
}
