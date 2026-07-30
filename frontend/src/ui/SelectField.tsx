import React from 'react'
import { ChevronDown } from 'lucide-react'

export type SelectOption = {
  value: string | number
  label: string
  disabled?: boolean
}

type Props = Omit<React.SelectHTMLAttributes<HTMLSelectElement>, 'children'> & {
  /** Label shown above the select */
  label?: string
  /** Options to render */
  options: SelectOption[]
  /** When truthy, shows red ring + error text */
  error?: string | boolean
  /** Placeholder option text (defaults to "— Select —") */
  placeholder?: string
  /** Optional helper/hint under the field (mutually exclusive with error text) */
  hint?: string
  /** Extra className to merge into the select */
  className?: string
}

/**
 * Pretty native select with consistent styles + custom chevron.
 * Keeps your original logic but makes it reusable.
 */
const SelectField = React.forwardRef<HTMLSelectElement, Props>(
  (
    {
      label,
      options,
      error,
      placeholder = '— Select —',
      hint,
      className = '',
      disabled,
      required,
      ...props
    },
    ref
  ) => {
    const selectEl = (
      <div className="relative">
        <select
          ref={ref}
          disabled={disabled}
          required={required}
          {...props}
          className={[
            // base: match your input look, hide native arrow, add space for chevron
            'input appearance-none pr-10',
            // dark-friendly surface + states (tweak if your .input already handles these)
            'bg-foreground/5 border-foreground/10 text-foreground placeholder-foreground/40',
            'hover:bg-foreground/10 focus:ring-2 focus:ring-foreground/20',
            disabled ? 'opacity-60 cursor-not-allowed' : '',
            error ? 'ring-1 ring-rose-400' : '',
            className,
          ].join(' ')}
        >
          <option value="">{placeholder}</option>
          {options.map(opt => (
            <option key={String(opt.value)} value={String(opt.value)} disabled={opt.disabled}>
              {opt.label}
            </option>
          ))}
        </select>

        {/* Chevron icon */}
        <ChevronDown
          className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-foreground/60"
          aria-hidden="true"
        />
      </div>
    )

    // With or without label wrapper
    return label ? (
      <label className="grid gap-1">
        <span className="text-sm text-foreground/70">{label}</span>
        {selectEl}
        {typeof error === 'string' && error && (
          <span className="text-xs text-rose-400">{error}</span>
        )}
        {!error && hint && <span className="text-xs text-foreground/50">{hint}</span>}
      </label>
    ) : (
      selectEl
    )
  }
)
SelectField.displayName = 'SelectField'
export default SelectField

/** Helper to turn an array of objects into SelectOptions */
export const toOptions = <T,>(
  items: T[],
  valueKey: keyof T,
  labelKey: keyof T
): SelectOption[] =>
  items.map((it: any) => ({
    value: it[valueKey],
    label: String(it[labelKey] ?? ''),
  }))
