import styles from './logo.module.css'

interface LogoProps {
  size?: number
}

export default function Logo({ size = 32 }: LogoProps) {
  return (
    <span className={styles.mark} style={{ width: size, height: size }} aria-hidden="true">
      <svg viewBox="0 0 32 32" width={size * 0.62} height={size * 0.62} fill="none">
        <circle cx="16" cy="16" r="11" stroke="currentColor" strokeWidth="2.5" opacity="0.35" />
        <path d="M16 5a11 11 0 0 1 11 11" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
        <path d="M16 10v6l4 2.5" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </span>
  )
}
