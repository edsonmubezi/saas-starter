// src/ui/DocViewerModal.tsx
import { useState, useEffect } from 'react'

type Props = {
  open: boolean
  title?: string
  url?: string | null
  onClose: () => void
}

export default function DocViewerModal({ open, title, url, onClose }: Props) {
  const [displayUrl, setDisplayUrl] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, onClose])

  // Convert cross-origin URLs to blob for iframe display
  useEffect(() => {
    if (!open || !url) { setDisplayUrl(null); return }

    let isCrossOrigin = false
    try {
      const docOrigin = new URL(url, window.location.href).origin
      isCrossOrigin = docOrigin !== window.location.origin
    } catch {
      isCrossOrigin = false
    }

    if (!isCrossOrigin) {
      setDisplayUrl(url)
      return
    }

    // Cross-origin: fetch as blob so iframe can render it
    let revoke: string | null = null
    setLoading(true)
    fetch(url)
      .then(res => {
        if (!res.ok) throw new Error('Failed to load')
        return res.blob()
      })
      .then(blob => {
        revoke = URL.createObjectURL(blob)
        setDisplayUrl(revoke)
      })
      .catch(() => {
        // Blob fetch failed — fall back to direct URL (Open in new tab still works)
        setDisplayUrl(url)
      })
      .finally(() => setLoading(false))

    return () => { if (revoke) URL.revokeObjectURL(revoke) }
  }, [open, url])

  if (!open) return null

  const safeUrl = url ?? ''
  const safeDisplay = displayUrl ?? ''
  const isBlob = safeDisplay.startsWith('blob:')

  // Extract extension from original URL path
  let ext = ''
  try {
    const pathname = new URL(safeUrl, window.location.href).pathname
    ext = pathname.split('.').pop()?.toLowerCase() ?? ''
  } catch {
    ext = safeUrl.split('?')[0].split('#')[0].split('.').pop()?.toLowerCase() ?? ''
  }
  const isImage = !isBlob && !!ext && ['png','jpg','jpeg','gif','webp','bmp'].includes(ext)
  const isPDF = isBlob || ext === 'pdf'

  return (
    <div className="fixed inset-0 z-[1000]">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/80"
        onClick={onClose}
        aria-hidden="true"
      />
      {/* Shell */}
      <div className="absolute inset-0 flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 bg-neutral-900/90 border-b border-foreground/10">
          <div className="text-sm text-foreground/80 truncate pr-3">
            {title || 'Document'}
          </div>
          <div className="flex items-center gap-2">
            {safeUrl && (
              <>
                <a
                  href={safeUrl}
                  download
                  className="px-3 py-1.5 rounded bg-foreground/10 hover:bg-foreground/20 text-foreground text-sm"
                >
                  Download
                </a>
                <a
                  href={safeUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="px-3 py-1.5 rounded bg-foreground/10 hover:bg-foreground/20 text-foreground text-sm"
                >
                  Open in new tab
                </a>
              </>
            )}
            <button
              onClick={onClose}
              className="px-3 py-1.5 rounded bg-foreground/10 hover:bg-foreground/20 text-foreground text-sm"
            >
              Close
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 min-h-0">
          {loading ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-foreground/60">Loading document...</div>
            </div>
          ) : !safeDisplay ? (
            <div className="p-6 text-foreground/70">No file URL available</div>
          ) : isImage ? (
            <div className="w-full h-full overflow-auto flex items-center justify-center bg-black">
              <img
                src={safeDisplay}
                alt={title || 'Document'}
                className="max-w-full max-h-full object-contain"
              />
            </div>
          ) : isPDF ? (
            <iframe
              src={safeDisplay}
              title={title || 'PDF'}
              className="w-full h-full"
            />
          ) : (
            <div className="p-6 text-foreground/80">
              Preview not supported.{' '}
              <a className="text-blue-400 hover:underline" href={safeUrl} target="_blank" rel="noreferrer">
                Download / Open
              </a>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
