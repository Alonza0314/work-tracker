import type { CategoryColor } from '../../api'
import { categoryColorStyle } from '../../work/format'
import styles from './category-chip.module.css'

interface CategoryChipProps {
  name: string
  color?: CategoryColor
}

export default function CategoryChip({ name, color }: CategoryChipProps) {
  return (
    <span className={styles.chip} style={categoryColorStyle(color)}>
      <span className={styles.dot} aria-hidden="true" />
      {name}
    </span>
  )
}
