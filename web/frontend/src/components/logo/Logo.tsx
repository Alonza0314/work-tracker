import styles from './logo.module.css'

interface LogoProps {
  size?: number
}

export default function Logo({ size = 32 }: LogoProps) {
  return (
    <img
      className={styles.mark}
      src="/wt-favicon.jpg"
      width={size}
      height={size}
      alt=""
      aria-hidden="true"
    />
  )
}
