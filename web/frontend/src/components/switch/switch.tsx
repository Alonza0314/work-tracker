import styles from './switch.module.css'

interface SwitchProps {
  checked: boolean
  onChange: (checked: boolean) => void
  label?: string
  // accessible name when the visible label lives outside the switch
  ariaLabel?: string
  disabled?: boolean
}

export default function Switch({ checked, onChange, label, ariaLabel, disabled = false }: SwitchProps) {
  return (
    <label className={styles.wrapper}>
      <span className={styles.switch}>
        <input
          type="checkbox"
          role="switch"
          aria-label={ariaLabel}
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          disabled={disabled}
        />
        <span className={styles.slider}></span>
      </span>
      {label && <span className={styles.label}>{label}</span>}
    </label>
  )
}
