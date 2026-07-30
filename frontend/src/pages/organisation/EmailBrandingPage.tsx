import React, { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Mail, Save, Loader2, Eye, ChevronDown, Search } from 'lucide-react'
import toast from 'react-hot-toast'
import {
  getEmailBranding,
  saveEmailBranding,
  getAvailableFonts,
  DEFAULT_EMAIL_BRANDING,
  type EmailBrandingSettings,
} from '../../utils/orgs'
import Modal from '../../ui/Modal'

/* ─── Toggle switch ───────────────────────────────────── */
function Toggle({ checked, onChange, label }: { checked: boolean; onChange: (v: boolean) => void; label: string }) {
  return (
    <label className="flex items-center justify-between py-2.5 cursor-pointer group">
      <span className="text-sm text-foreground/80 group-hover:text-foreground transition-colors">{label}</span>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ${checked ? 'bg-blue-600' : 'bg-foreground/20'}`}
      >
        <span className={`pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform transition-transform duration-200 ${checked ? 'translate-x-5' : 'translate-x-0'}`} />
      </button>
    </label>
  )
}

/* ─── Color picker ────────────────────────────────────── */
function ColorInput({ value, onChange, label }: { value: string; onChange: (v: string) => void; label: string }) {
  return (
    <div className="flex items-center justify-between py-2.5">
      <span className="text-sm text-foreground/80">{label}</span>
      <div className="flex items-center gap-2">
        <input
          type="color"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="w-8 h-8 rounded border border-foreground/20 cursor-pointer bg-transparent"
        />
        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="w-20 px-2 py-1 text-xs bg-foreground/5 border border-foreground/10 rounded outline-none focus:ring-1 focus:ring-blue-500 font-mono"
          maxLength={7}
        />
      </div>
    </div>
  )
}


/* ─── Email Preview ───────────────────────────────────── */
type PreviewType = 'leave-request' | 'leave-approved' | 'leave-rejected' | 'salary-slip'

/* ─── Detail row for info box ─────────────────────────── */
const rowStyle: React.CSSProperties = { borderBottom: '1px solid #eee' }
const labelStyle: React.CSSProperties = { fontWeight: 600, color: '#555', padding: '8px 12px 8px 0', whiteSpace: 'nowrap', verticalAlign: 'top' }
const valStyle: React.CSSProperties = { padding: '8px 0', color: '#333' }

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <tr style={rowStyle}>
      <td style={labelStyle}>{label}</td>
      <td style={valStyle}>{value}</td>
    </tr>
  )
}

function EmailPreview({ form, previewType }: { form: EmailBrandingSettings; previewType: PreviewType }) {
  const font = form.font_family || 'Arial, sans-serif'

  // Determine header color based on template type
  const getHeaderColor = () => {
    switch (previewType) {
      case 'leave-approved': return '#10B981'
      case 'leave-rejected': return '#EF4444'
      default: return form.primary_color
    }
  }

  const getHeaderTextColor = () => {
    switch (previewType) {
      case 'leave-approved':
      case 'leave-rejected':
        return '#FFFFFF'
      default: return form.header_text_color
    }
  }

  const getInfoBoxBorder = () => {
    switch (previewType) {
      case 'leave-approved': return '#10B981'
      case 'leave-rejected': return '#EF4444'
      default: return form.accent_color
    }
  }

  const getInfoBoxBg = () => {
    switch (previewType) {
      case 'leave-approved': return '#ECFDF5'
      case 'leave-rejected': return '#FEF2F2'
      default: return '#EEF2FF'
    }
  }

  const titles: Record<PreviewType, string> = {
    'leave-request': 'New Leave Request',
    'leave-approved': 'Leave Approved',
    'leave-rejected': 'Leave Request Rejected',
    'salary-slip': 'Salary Slip - January 2026',
  }

  const infoBoxStyle: React.CSSProperties = {
    backgroundColor: getInfoBoxBg(),
    borderLeft: `4px solid ${getInfoBoxBorder()}`,
    padding: '4px 16px',
    margin: '16px 0',
    borderRadius: 4,
  }

  return (
    <div
      style={{ maxWidth: 480, fontFamily: font, color: '#333', fontSize: 14, lineHeight: 1.6, margin: '0 auto', backgroundColor: '#fff', borderRadius: 8, overflow: 'hidden', boxShadow: '0 4px 20px rgba(0,0,0,0.15)' }}
    >
      {/* Header */}
      <div
        style={{ backgroundColor: getHeaderColor(), color: getHeaderTextColor(), padding: '20px', textAlign: 'center' }}
      >
        <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, fontFamily: font }}>{titles[previewType]}</h2>
      </div>

      {/* Content */}
      <div style={{ padding: '24px 28px', backgroundColor: '#f9f9f9', borderLeft: '1px solid #ddd', borderRight: '1px solid #ddd', fontFamily: font }}>
        {previewType === 'leave-request' && (
          <>
            <p style={{ margin: '0 0 12px' }}>A new leave request has been submitted and requires your attention.</p>
            <div style={infoBoxStyle}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <tbody>
                  <InfoRow label="Employee:" value="John Doe" />
                  <InfoRow label="Leave Type:" value="Annual Leave" />
                  <InfoRow label="From:" value="2026-03-10" />
                  <InfoRow label="To:" value="2026-03-14" />
                  <InfoRow label="Days:" value="5" />
                  <InfoRow label="Reason:" value="Family vacation" />
                </tbody>
              </table>
            </div>
            <p style={{ margin: '0 0 12px' }}>Please log in to the system to review and take action on this request.</p>
          </>
        )}

        {previewType === 'leave-approved' && (
          <>
            <div style={{ textAlign: 'center', fontSize: 48, margin: '12px 0' }}>&#10003;</div>
            <p style={{ margin: '0 0 12px' }}>Dear John Doe,</p>
            <p style={{ margin: '0 0 12px' }}>Your leave request has been approved.</p>
            <div style={infoBoxStyle}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <tbody>
                  <InfoRow label="Leave Type:" value="Annual Leave" />
                  <InfoRow label="From:" value="2026-03-10" />
                  <InfoRow label="To:" value="2026-03-14" />
                  <InfoRow label="Days:" value="5" />
                  <InfoRow label="Approved By:" value="Jane Smith" />
                </tbody>
              </table>
            </div>
          </>
        )}

        {previewType === 'leave-rejected' && (
          <>
            <p style={{ margin: '0 0 12px' }}>Dear John Doe,</p>
            <p style={{ margin: '0 0 12px' }}>Unfortunately, your leave request has been rejected.</p>
            <div style={infoBoxStyle}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <tbody>
                  <InfoRow label="Leave Type:" value="Annual Leave" />
                  <InfoRow label="From:" value="2026-03-10" />
                  <InfoRow label="To:" value="2026-03-14" />
                  <InfoRow label="Days:" value="5" />
                  <InfoRow label="Reason:" value="Insufficient leave balance" />
                </tbody>
              </table>
            </div>
          </>
        )}

        {previewType === 'salary-slip' && (
          <>
            <p style={{ margin: '0 0 12px' }}>Dear John Doe,</p>
            <p style={{ margin: '0 0 12px' }}>Please find your salary slip for the period <strong>January 2026</strong> attached to this email.</p>
            <div style={infoBoxStyle}>
              <p style={{ margin: 0, color: '#666' }}>Your salary slip has been generated and is attached to this email as a PDF.</p>
            </div>
            <p style={{ margin: '0 0 12px' }}>If you have any questions regarding your salary slip, please contact the HR department.</p>
          </>
        )}

        {form.sign_off ? (
          <p style={{ margin: 0, whiteSpace: 'pre-line' }}>{form.sign_off}</p>
        ) : (
          <p style={{ margin: 0 }}>Best regards,<br/>Acme Corporation Team</p>
        )}
      </div>

      {/* Footer */}
      {form.footer_text && (
        <div style={{ padding: '12px 28px', textAlign: 'center', fontSize: 12, color: '#999', borderLeft: '1px solid #ddd', borderRight: '1px solid #ddd', borderBottom: '1px solid #ddd', borderRadius: '0 0 5px 5px' }}>
          <p style={{ margin: 0 }}>{form.footer_text}</p>
        </div>
      )}
    </div>
  )
}

/* ═══════════════════════════════════════════════════════════
   Main Page
   ═══════════════════════════════════════════════════════════ */
export default function EmailBrandingPage() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<EmailBrandingSettings>(DEFAULT_EMAIL_BRANDING)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [previewType, setPreviewType] = useState<PreviewType>('leave-request')
  const [fontSearch, setFontSearch] = useState('')
  const [fontOpen, setFontOpen] = useState(false)
  const fontDropRef = useRef<HTMLDivElement>(null)

  const { data: settings, isLoading } = useQuery({
    queryKey: ['email-branding'],
    queryFn: getEmailBranding,
    staleTime: 60_000,
  })

  const { data: availableFonts = ['Arial', 'Helvetica', 'Times', 'Courier'] } = useQuery({
    queryKey: ['available-fonts'],
    queryFn: getAvailableFonts,
    staleTime: 300_000,
  })

  useEffect(() => {
    if (settings) setForm(settings)
  }, [settings])

  // Close font dropdown on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (fontDropRef.current && !fontDropRef.current.contains(e.target as Node)) setFontOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const saveMutation = useMutation({
    mutationFn: saveEmailBranding,
    onSuccess: () => {
      toast.success('Email branding saved')
      queryClient.invalidateQueries({ queryKey: ['email-branding'] })
    },
    onError: (err: any) => toast.error(err?.message || 'Failed to save'),
  })

  const handleSave = () => saveMutation.mutate(form)

  const set = <K extends keyof EmailBrandingSettings>(key: K, value: EmailBrandingSettings[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }))

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-6 h-6 animate-spin text-foreground/40" />
      </div>
    )
  }

  const previewTabs: { key: PreviewType; label: string }[] = [
    { key: 'leave-request', label: 'Leave Request' },
    { key: 'leave-approved', label: 'Approved' },
    { key: 'leave-rejected', label: 'Rejected' },
    { key: 'salary-slip', label: 'Salary Slip' },
  ]

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl md:text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent flex items-center gap-2">
            <Mail className="w-6 h-6 text-blue-400" />
            Email Branding
          </h1>
          <p className="text-foreground/60 text-sm mt-1">Customize the look of emails sent by the system</p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setPreviewOpen(true)}
            className="px-4 py-2.5 rounded-xl border border-foreground/10 hover:bg-foreground/5 text-foreground/80 font-medium transition-all flex items-center gap-2"
          >
            <Eye className="w-4 h-4" />
            Preview
          </button>
          <button
            onClick={handleSave}
            disabled={saveMutation.isPending}
            className="px-5 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-medium shadow-lg shadow-blue-500/20 transition-all flex items-center gap-2 disabled:opacity-50"
          >
            {saveMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            {saveMutation.isPending ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>

      {/* Main Grid: Settings + Inline Preview */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Column 1: Settings */}
        <div className="space-y-4">
          {/* Colors */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-3">Colors</h3>
            <div className="divide-y divide-foreground/5">
              <ColorInput value={form.primary_color} onChange={(v) => set('primary_color', v)} label="Primary Color" />
              <ColorInput value={form.header_text_color} onChange={(v) => set('header_text_color', v)} label="Header Text Color" />
              <ColorInput value={form.accent_color} onChange={(v) => set('accent_color', v)} label="Accent Color" />
            </div>
            <p className="text-xs text-foreground/40 mt-3">
              Primary color is used for email headers and buttons. Accent color is used for info box borders. Semantic colors (green for approved, red for rejected) are not affected.
            </p>
          </div>

          {/* Typography */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-3">Typography</h3>
            <div className="py-2">
              <span className="text-sm text-foreground/80 block mb-1.5">Font Family</span>
              <div className="relative" ref={fontDropRef}>
                <button
                  type="button"
                  onClick={() => { setFontOpen(!fontOpen); setFontSearch('') }}
                  className="w-full flex items-center justify-between px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg text-sm hover:bg-foreground/10 transition-all"
                >
                  <span style={{ fontFamily: form.font_family }}>{form.font_family || 'Arial'}</span>
                  <ChevronDown className={`w-4 h-4 text-foreground/40 transition-transform ${fontOpen ? 'rotate-180' : ''}`} />
                </button>
                {fontOpen && (
                  <div className="absolute z-50 mt-1 w-full bg-surface-elevated border border-foreground/10 rounded-lg shadow-xl max-h-72 flex flex-col overflow-hidden">
                    <div className="flex items-center gap-2 px-3 py-2 border-b border-foreground/10">
                      <Search className="w-4 h-4 text-foreground/40 shrink-0" />
                      <input
                        type="text"
                        value={fontSearch}
                        onChange={(e) => setFontSearch(e.target.value)}
                        className="w-full bg-transparent outline-none text-sm placeholder:text-foreground/30"
                        placeholder="Search fonts..."
                        autoFocus
                      />
                    </div>
                    <div className="overflow-y-auto flex-1">
                      {availableFonts
                        .filter((f) => f.toLowerCase().includes(fontSearch.toLowerCase()))
                        .map((f) => (
                          <button
                            key={f}
                            type="button"
                            onClick={() => { set('font_family', f); setFontOpen(false) }}
                            className={`w-full text-left px-3 py-1.5 text-sm hover:bg-foreground/10 transition-colors flex items-center justify-between ${form.font_family === f ? 'bg-blue-600/20 text-blue-400' : ''}`}
                          >
                            <span style={{ fontFamily: f }}>{f}</span>
                            {form.font_family === f && <span className="text-xs text-blue-400">Selected</span>}
                          </button>
                        ))}
                      {availableFonts.filter((f) => f.toLowerCase().includes(fontSearch.toLowerCase())).length === 0 && (
                        <div className="px-3 py-4 text-sm text-foreground/40 text-center">No fonts match "{fontSearch}"</div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Header */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-3">Header</h3>
            <Toggle checked={form.show_logo} onChange={(v) => set('show_logo', v)} label="Show Organization Logo" />
          </div>

          {/* Sign-Off */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-3">Sign-Off</h3>
            <div className="py-2">
              <label className="block text-sm text-foreground/80 mb-1">Closing Text</label>
              <textarea
                value={form.sign_off}
                onChange={(e) => set('sign_off', e.target.value)}
                className="w-full px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg outline-none focus:ring-2 focus:ring-blue-500/40 text-sm transition-all resize-none"
                placeholder="e.g. Best regards,&#10;Acme Corporation Team"
                rows={3}
              />
              <p className="text-xs text-foreground/40 mt-1">Appears at the bottom of the email body. Leave empty to use default (Best regards, Company Team)</p>
            </div>
          </div>

          {/* Footer */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-3">Footer</h3>
            <div className="py-2">
              <label className="block text-sm text-foreground/80 mb-1">Custom Footer Text</label>
              <input
                type="text"
                value={form.footer_text}
                onChange={(e) => set('footer_text', e.target.value)}
                className="w-full px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg outline-none focus:ring-2 focus:ring-blue-500/40 text-sm transition-all"
                placeholder="e.g. Powered by Acme Corp"
              />
              <p className="text-xs text-foreground/40 mt-1">Displayed in the footer section below the email body</p>
            </div>
          </div>
        </div>

        {/* Column 2: Inline Live Preview */}
        <div className="space-y-4">
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider">Live Preview</h3>
              <div className="flex gap-1">
                {previewTabs.map((tab) => (
                  <button
                    key={tab.key}
                    onClick={() => setPreviewType(tab.key)}
                    className={`px-2.5 py-1 text-xs rounded-md transition-all ${previewType === tab.key ? 'bg-blue-600 text-white' : 'text-foreground/50 hover:bg-foreground/10'}`}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>
            </div>
            <div className="overflow-auto max-h-[70vh]">
              <EmailPreview form={form} previewType={previewType} />
            </div>
          </div>
        </div>
      </div>

      {/* Full Preview Modal */}
      <Modal open={previewOpen} onClose={() => setPreviewOpen(false)} title="Email Preview" size="lg">
        <div className="mb-4 flex justify-center">
          <div className="flex gap-1 border-b border-foreground/10">
            {previewTabs.map((tab) => (
              <button
                key={tab.key}
                onClick={() => setPreviewType(tab.key)}
                className={`px-4 py-1.5 text-xs font-medium border-b-2 transition-all -mb-px ${previewType === tab.key ? 'border-blue-500 text-blue-400' : 'border-transparent text-foreground/50 hover:text-foreground/70'}`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>
        <div className="flex justify-center">
          <div className="w-full max-w-xl">
            <EmailPreview form={form} previewType={previewType} />
          </div>
        </div>
      </Modal>
    </div>
  )
}
