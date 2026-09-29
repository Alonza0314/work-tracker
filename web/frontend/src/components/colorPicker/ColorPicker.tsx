import { useEffect, useRef, useState } from 'react'
import { Check } from 'lucide-react'
import { CategoryColor } from '../../api'
import { categoryColorStyle } from '../../work/format'
import styles from './color-picker.module.css'

interface ColorPickerProps {
  value?: CategoryColor
  disabled?: boolean
  label: string
  onChange: (color: CategoryColor) => void
}

const COLORS = Object.values(CategoryColor)

// a swatch button that opens the category palette
export default function ColorPicker({ value, disabled = false, label, onChange }: ColorPickerProps) {
  // fixed position of the open palette, so panels with overflow: hidden do
  // not clip it; null while closed
  const [position, setPosition] = useState<{ top: number, left: number } | null>(null)
  const open = position !== null
  const root = useRef<HTMLDivElement>(null)

  function toggle() {
    const rect = root.current?.getBoundingClientRect()
    setPosition(open || !rect ? null : { top: rect.bottom + 6, left: rect.left + rect.width / 2 })
  }

  useEffect(() => {
    if (!open) {
      return
    }
    const closeOnScroll = () => setPosition(null)
    function close(event: MouseEvent | KeyboardEvent) {
      if (event instanceof KeyboardEvent ? event.key === 'Escape' : !root.current?.contains(event.target as Node)) {
        setPosition(null)
      }
    }
    document.addEventListener('mousedown', close)
    document.addEventListener('keydown', close)
    window.addEventListener('scroll', closeOnScroll, true)
    return () => {
      document.removeEventListener('mousedown', close)
      document.removeEventListener('keydown', close)
      window.removeEventListener('scroll', closeOnScroll, true)
    }
  }, [open])

  return (
    <div className={styles.root} ref={root}>
      <button
        type="button"
        className={styles.swatch}
        style={categoryColorStyle(value)}
        disabled={disabled}
        onClick={toggle}
        aria-label={label}
        aria-expanded={open}
        title={label}
      />
      {open && (
        <div className={styles.palette} style={position} role="listbox" aria-label={label}>
          {COLORS.map((color) => (
            <button
              key={color}
              type="button"
              role="option"
              aria-selected={color === value}
              aria-label={color}
              title={color}
              className={styles.option}
              style={categoryColorStyle(color)}
              onClick={() => {
                setPosition(null)
                if (color !== value) {
                  onChange(color)
                }
              }}
            >
              {color === value && <Check size={12} strokeWidth={3} aria-hidden="true" />}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
