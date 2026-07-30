
import React from 'react'
export default function StatusFilter({ value, onChange }:{ value: '' | 'active' | 'disabled'; onChange:(v:''|'active'|'disabled')=>void }) {
  return (
    <div className="flex items-center gap-2">
      <span className="text-sm text-foreground/70">Status</span>
      <select className="select w-full sm:w-[140px]" value={value} onChange={e=>onChange(e.target.value as any)}>
        <option value="">All</option>
        <option value="active">Active</option>
        <option value="disabled">Disabled</option>
      </select>
    </div>
  )
}
