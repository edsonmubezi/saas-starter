
import React, { useEffect } from 'react'
import clsx from 'clsx'

type Props = { open: boolean; title?: string; onClose: () => void; children: React.ReactNode; footer?: React.ReactNode; size?: 'sm'|'md'|'lg'|'full' }
export default function Modal({ open, title, onClose, children, footer, size='md' }: Props) {
  useEffect(() => { function onKey(e: KeyboardEvent) { if (e.key === 'Escape') onClose() } if (open) document.addEventListener('keydown', onKey); return () => document.removeEventListener('keydown', onKey) }, [open, onClose])
  if (!open) return null
  const maxw = size === 'sm'
    ? 'max-w-sm sm:max-w-md'
    : size === 'full'
      ? 'max-w-[96vw] lg:max-w-6xl'
      : size === 'lg'
        ? 'max-w-[95vw] sm:max-w-2xl lg:max-w-3xl'
        : 'max-w-[90vw] sm:max-w-lg lg:max-w-xl'
  return (
    <div className="fixed inset-0 z-50 grid place-items-center p-4">
      <div className="fixed inset-0 bg-black/40" onClick={onClose} aria-hidden="true" />
      <div role="dialog" aria-modal="true" aria-label={title || 'Dialog'} className={clsx('card relative w-full mx-2 sm:mx-4 p-4 sm:p-6 max-h-[90vh] flex flex-col', maxw)}>
        {title && <h2 className="text-lg font-semibold mb-4 flex-shrink-0">{title}</h2>}
        <div className="overflow-y-auto flex-1 min-h-0">{children}</div>
        {footer && <div className="mt-6 flex justify-end gap-2 flex-shrink-0">{footer}</div>}
        <button aria-label="Close" onClick={onClose} className="absolute top-3 right-3 rounded-lg px-2 py-1 text-foreground/60 hover:bg-foreground/10">&times;</button>
      </div>
    </div>
  )
}
