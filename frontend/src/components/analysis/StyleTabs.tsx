import { createContext, useContext, useId, useRef } from 'react'
import type { KeyboardEvent, ReactNode } from 'react'
import styles from './StyleTabs.module.css'

interface StyleTabsContextValue {
  selected: string
  select: (value: string) => void
  tabId: (value: string) => string
  panelId: (value: string) => string
}

const StyleTabsContext = createContext<StyleTabsContextValue | null>(null)

const ARROW_STEP: Record<string, number> = {
  ArrowRight: 1,
  ArrowLeft: -1,
}

function useStyleTabs(): StyleTabsContextValue {
  const context = useContext(StyleTabsContext)
  if (!context) {
    throw new Error('StyleTab phải nằm trong <StyleTabs>')
  }
  return context
}

interface StyleTabsProps {
  selected: string
  onSelect: (value: string) => void
  children: ReactNode
}

export function StyleTabs({ selected, onSelect, children }: StyleTabsProps) {
  const baseId = useId()

  const context: StyleTabsContextValue = {
    selected,
    select: onSelect,
    tabId: (value) => `${baseId}-tab-${value}`,
    panelId: (value) => `${baseId}-panel-${value}`,
  }

  return <StyleTabsContext.Provider value={context}>{children}</StyleTabsContext.Provider>
}

interface StyleTabListProps {
  label: string
  children: ReactNode
}

export function StyleTabList({ label, children }: StyleTabListProps) {
  const { select } = useStyleTabs()
  const listRef = useRef<HTMLDivElement>(null)

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    const tabs = listRef.current ? Array.from(listRef.current.querySelectorAll('[role="tab"]')) : []
    if (tabs.length === 0) return

    const focusedIndex = tabs.findIndex((tab) => tab === document.activeElement)
    const arrowStep = ARROW_STEP[event.key]

    let nextIndex: number
    if (event.key === 'Home') {
      nextIndex = 0
    } else if (event.key === 'End') {
      nextIndex = tabs.length - 1
    } else if (arrowStep !== undefined && focusedIndex !== -1) {
      nextIndex = (focusedIndex + arrowStep + tabs.length) % tabs.length
    } else {
      return
    }

    event.preventDefault()
    const next = tabs[nextIndex]
    if (!(next instanceof HTMLElement)) return

    next.focus()
    const value = next.dataset.value
    if (value) select(value)
  }

  return (
    <div ref={listRef} className={styles.list} role="tablist" aria-label={label} aria-orientation="horizontal" onKeyDown={handleKeyDown}>
      {children}
    </div>
  )
}

interface StyleTabProps {
  value: string
  children: ReactNode
}

export function StyleTab({ value, children }: StyleTabProps) {
  const { selected, select, tabId, panelId } = useStyleTabs()
  const isSelected = selected === value

  return (
    <button
      type="button"
      role="tab"
      id={tabId(value)}
      data-value={value}
      className={styles.tab}
      aria-selected={isSelected}
      aria-controls={panelId(value)}
      tabIndex={isSelected ? 0 : -1}
      onClick={() => select(value)}
    >
      {children}
    </button>
  )
}

interface StyleTabPanelProps {
  value: string
  children: ReactNode
}

export function StyleTabPanel({ value, children }: StyleTabPanelProps) {
  const { selected, tabId, panelId } = useStyleTabs()

  return (
    <div
      role="tabpanel"
      id={panelId(value)}
      aria-labelledby={tabId(value)}
      tabIndex={0}
      className={styles.panel}
      hidden={selected !== value}
    >
      {children}
    </div>
  )
}
