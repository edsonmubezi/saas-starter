// src/ui/NumberInput.tsx
import React from "react"

interface NumberInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "value" | "onChange"> {
  value: string | number
  onValueChange: (value: string) => void
}

export default function NumberInput({ value, onValueChange, ...rest }: NumberInputProps) {
  // format value for display
  const formatValue = (val: string | number) => {
    if (val === null || val === undefined || val === "") return ""
    const num = Number(val)
    return isNaN(num) ? String(val) : num.toLocaleString()
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value.replace(/,/g, "") // strip commas
    onValueChange(raw)
  }

  return (
    <input
      {...rest}
      inputMode="decimal"
      value={formatValue(value)}
      onChange={handleChange}
    />
  )
}
