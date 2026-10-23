import type { ButtonHTMLAttributes, ReactNode } from 'react'
import styles from './Button.module.css'

type ButtonVariant = 'primary' | 'ghost' | 'quiet'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  children: ReactNode
}

export function Button({ variant = 'ghost', type = 'button', className, children, ...rest }: ButtonProps) {
  const variantClass = styles[variant]
  const composed = [styles.button, variantClass, className].filter(Boolean).join(' ')

  return (
    <button type={type} className={composed} {...rest}>
      {children}
    </button>
  )
}
