import React, { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Palette, Save, Loader2, FileText, Table, Eye, Upload, Image, Type, X, Search, ChevronDown } from 'lucide-react'
import toast from 'react-hot-toast'
import {
  getDocumentBranding,
  saveDocumentBranding,
  uploadWatermarkImage,
  getAvailableFonts,
  DEFAULT_BRANDING,
  type DocumentBrandingSettings,
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

/* ─── Text input row ──────────────────────────────────── */
function TextRow({ value, onChange, label, placeholder }: { value: string; onChange: (v: string) => void; label: string; placeholder?: string }) {
  return (
    <div className="py-2">
      <label className="block text-sm text-foreground/80 mb-1">{label}</label>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg outline-none focus:ring-2 focus:ring-blue-500/40 text-sm transition-all"
        placeholder={placeholder}
      />
    </div>
  )
}

/* ─── Shared PDF page shell (header + watermark + footer wrapper) ─── */
function PDFPageShell({ form, title, children }: { form: DocumentBrandingSettings; title: string; children: React.ReactNode }) {
  const displayName = form.header_org_name || 'Your Organization'
  const displayAddr = form.header_address || '123 Business Street, City'
  const displayPhone = form.header_phone || '+123 456 789'
  const displayEmail = form.header_email || 'info@company.com'
  const displayTIN = form.header_tin || '123-456-789'
  const font = form.font_family

  return (
    <div className="bg-white rounded-lg shadow-xl overflow-hidden mx-auto" style={{ aspectRatio: '210/297', maxHeight: '70vh' }}>
      <div className="h-full flex flex-col text-black p-6 text-[11px] relative">
        {/* Watermark */}
        {form.show_watermark && (
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none overflow-hidden">
            {form.watermark_type === 'image' && form.watermark_image_path ? (
              <img src={form.watermark_image_path} alt="watermark" className="opacity-[0.06] max-w-[60%] max-h-[60%] rotate-[-15deg]" />
            ) : (
              <span className="text-5xl font-bold opacity-[0.07] rotate-[-35deg] whitespace-nowrap select-none" style={{ fontFamily: font }}>
                {form.watermark_text || displayName}
              </span>
            )}
          </div>
        )}

        {/* Header */}
        <div className="flex items-start gap-3 mb-1 relative z-10">
          {form.show_logo && (
            <div className="w-12 h-12 rounded bg-gray-100 flex items-center justify-center flex-shrink-0 border border-gray-200">
              <FileText className="w-6 h-6 text-foreground/50" />
            </div>
          )}
          <div className="flex-1 min-w-0">
            {form.show_org_name && <div className="font-bold text-sm leading-tight" style={{ fontFamily: font }}>{displayName}</div>}
            {form.show_address && <div className="text-[9px] text-foreground/50 mt-0.5" style={{ fontFamily: font }}>{displayAddr}</div>}
            {form.show_contact && (
              <div className="text-[9px] text-foreground/40" style={{ fontFamily: font }}>Tel: {displayPhone} | Email: {displayEmail}</div>
            )}
            {form.show_tin && <div className="text-[9px] text-foreground/40" style={{ fontFamily: font }}>TIN: {displayTIN}</div>}
          </div>
        </div>

        <div className="h-[2px] my-2" style={{ backgroundColor: form.primary_color }} />

        {/* Document title */}
        <div className="text-center font-bold text-sm my-2" style={{ fontFamily: font }}>{title}</div>

        {/* Content */}
        <div className="flex-1 relative z-10 min-h-0 overflow-hidden" style={{ fontFamily: font }}>
          {children}
        </div>

        {/* Footer */}
        {form.show_footer && (
          <div className="border-t border-gray-200 pt-2 mt-2 flex justify-between text-[8px] text-foreground/50 relative z-10" style={{ fontFamily: font }}>
            <span>
              {form.footer_text && <>{form.footer_text} {form.show_generated_date && '| '}</>}
              {form.show_generated_date && `Generated: ${new Date().toLocaleDateString()}`}
            </span>
            {form.show_page_numbers && <span>Page 1 of 1</span>}
          </div>
        )}
      </div>
    </div>
  )
}

/* ─── Sample PDF Preview ──────────────────────────────── */
function SamplePDFPreview({ form }: { form: DocumentBrandingSettings }) {
  return (
    <PDFPageShell form={form} title="Sample Document">
      <div className="space-y-2 px-2 pt-2">
        <div className="h-2.5 bg-gray-200 rounded w-full" />
        <div className="h-2.5 bg-gray-200 rounded w-5/6" />
        <div className="h-2.5 bg-gray-200 rounded w-4/5" />
        <div className="h-2.5 bg-gray-100 rounded w-full mt-4" />
        <div className="h-2.5 bg-gray-100 rounded w-3/4" />
        <div className="h-2.5 bg-gray-100 rounded w-5/6" />
        <div className="h-2.5 bg-gray-200 rounded w-full mt-4" />
        <div className="h-2.5 bg-gray-200 rounded w-2/3" />
      </div>
    </PDFPageShell>
  )
}

/* ─── Salary Slip PDF Preview ─────────────────────────── */
const fmtMoney = (n: number) => n.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })

function SalarySlipPreview({ form }: { form: DocumentBrandingSettings }) {
  const font = form.font_family
  const thCls = 'px-2 py-1 text-left text-[9px] font-semibold'
  const tdCls = 'px-2 py-0.5 text-[9px]'
  const tdR = `${tdCls} text-right`
  const rowAlt = 'bg-gray-50'
  const sectionHead = 'bg-gray-100'
  const totalRow = 'bg-gray-200 font-semibold'

  return (
    <PDFPageShell form={form} title="Salary Slip">
      <div className="space-y-2 text-[9px]">
        {/* Employee info row */}
        <div className="flex justify-between px-1 py-1 bg-gray-50 rounded text-[9px]" style={{ fontFamily: font }}>
          <span><strong>Employee:</strong> John Doe</span>
          <span><strong>Period:</strong> January 2026</span>
        </div>

        {/* Earnings table */}
        <table className="w-full border-collapse" style={{ fontFamily: font }}>
          <thead>
            <tr className={sectionHead}>
              <th className={thCls}>Earnings</th>
              <th className={`${thCls} text-right`}>Amount (TZS)</th>
            </tr>
          </thead>
          <tbody>
            <tr><td className={tdCls}>Basic Salary</td><td className={tdR}>{fmtMoney(1500000)}</td></tr>
            <tr className={rowAlt}><td className={tdCls}>Housing Allowance</td><td className={tdR}>{fmtMoney(300000)}</td></tr>
            <tr><td className={tdCls}>Transport Allowance</td><td className={tdR}>{fmtMoney(150000)}</td></tr>
            <tr className={totalRow}><td className={tdCls}>Gross Salary</td><td className={tdR}>{fmtMoney(1950000)}</td></tr>
          </tbody>
        </table>

        {/* Deductions table */}
        <table className="w-full border-collapse" style={{ fontFamily: font }}>
          <thead>
            <tr className={sectionHead}>
              <th className={thCls}>Deductions</th>
              <th className={`${thCls} text-right`}>Amount (TZS)</th>
            </tr>
          </thead>
          <tbody>
            <tr><td className={tdCls}>NSSF (Pension)</td><td className={tdR}>{fmtMoney(195000)}</td></tr>
            <tr className={rowAlt}><td className={tdCls}>NHIF (Health Insurance)</td><td className={tdR}>{fmtMoney(58500)}</td></tr>
            <tr><td className={tdCls}>PAYE (Tax)</td><td className={tdR}>{fmtMoney(247350)}</td></tr>
            <tr className={rowAlt}><td className={tdCls}>Loan Repayment <span className="text-foreground/50">(Remaining: 1,200,000)</span></td><td className={tdR}>{fmtMoney(100000)}</td></tr>
            <tr className={totalRow}><td className={tdCls}>Total Deductions</td><td className={tdR}>{fmtMoney(600850)}</td></tr>
          </tbody>
        </table>

        {/* Non-taxable benefits */}
        <table className="w-full border-collapse" style={{ fontFamily: font }}>
          <thead>
            <tr className={sectionHead}>
              <th className={thCls}>Non-Taxable Benefits</th>
              <th className={`${thCls} text-right`}>Amount (TZS)</th>
            </tr>
          </thead>
          <tbody>
            <tr><td className={tdCls}>Lunch Allowance</td><td className={tdR}>{fmtMoney(100000)}</td></tr>
            <tr className={totalRow}><td className={tdCls}>Total Non-Taxable Benefits</td><td className={tdR}>{fmtMoney(100000)}</td></tr>
          </tbody>
        </table>

        {/* Net salary */}
        <div className="flex justify-between items-center px-3 py-2 rounded text-white text-[11px] font-bold" style={{ backgroundColor: '#4287ca' }}>
          <span>NET SALARY</span>
          <span>{fmtMoney(1449150)}</span>
        </div>

        {/* Footer note */}
        <p className="text-[7px] text-foreground/50 italic px-1">
          This is a computer-generated document and does not require a signature. For any queries, please contact the HR department.
        </p>
      </div>
    </PDFPageShell>
  )
}

/* ─── Excel Preview ───────────────────────────────────── */
function ExcelPreview({ form }: { form: DocumentBrandingSettings }) {
  const displayName = form.header_org_name || 'Your Organization'
  return (
    <div className="bg-white rounded-lg shadow-xl overflow-hidden mx-auto" style={{ maxHeight: '70vh' }}>
      <div className="p-4">
        <table className="w-full border-collapse text-[11px] text-black">
          <thead>
            <tr>
              <th
                colSpan={5}
                className="p-3 text-center font-bold border border-gray-300 text-sm"
                style={{ backgroundColor: form.primary_color, color: form.header_text_color, fontFamily: form.font_family }}
              >
                {displayName} — Employee Report
              </th>
            </tr>
            <tr>
              {['#', 'Name', 'Department', 'Position', 'Status'].map((h) => (
                <th
                  key={h}
                  className="p-2 text-center font-bold border border-gray-300"
                  style={{ backgroundColor: form.primary_color, color: form.header_text_color, fontFamily: form.font_family }}
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody style={{ fontFamily: form.font_family }}>
            {[
              ['1', 'John Doe', 'Engineering', 'Developer', 'Active'],
              ['2', 'Jane Smith', 'Finance', 'Accountant', 'Active'],
              ['3', 'Bob Wilson', 'HR', 'Manager', 'On Leave'],
              ['4', 'Alice Brown', 'Marketing', 'Designer', 'Active'],
              ['5', 'Tom Clark', 'Sales', 'Executive', 'Active'],
            ].map((row, i) => (
              <tr key={i} className={i % 2 === 1 ? 'bg-gray-50' : ''}>
                {row.map((cell, j) => (
                  <td key={j} className="p-1.5 border border-gray-200 text-center">{cell}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

/* ═══════════════════════════════════════════════════════════
   Main Page
   ═══════════════════════════════════════════════════════════ */
export default function DocumentBrandingPage() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<DocumentBrandingSettings>(DEFAULT_BRANDING)
  const [previewOpen, setPreviewOpen] = useState(false)
  const [previewTab, setPreviewTab] = useState<'pdf' | 'excel'>('pdf')
  const [pdfDoc, setPdfDoc] = useState<'sample' | 'salary-slip'>('salary-slip')
  const [uploading, setUploading] = useState(false)
  const [fontSearch, setFontSearch] = useState('')
  const [fontOpen, setFontOpen] = useState(false)
  const fontDropRef = useRef<HTMLDivElement>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  const { data: settings, isLoading } = useQuery({
    queryKey: ['document-branding'],
    queryFn: getDocumentBranding,
    staleTime: 60_000,
  })

  const { data: availableFonts = ['Arial', 'Helvetica', 'Times', 'Courier'] } = useQuery({
    queryKey: ['available-fonts'],
    queryFn: getAvailableFonts,
    staleTime: 300_000,
  })

  // Close font dropdown on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (fontDropRef.current && !fontDropRef.current.contains(e.target as Node)) setFontOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  useEffect(() => {
    if (settings) setForm(settings)
  }, [settings])

  const saveMutation = useMutation({
    mutationFn: saveDocumentBranding,
    onSuccess: () => {
      toast.success('Document branding saved')
      queryClient.invalidateQueries({ queryKey: ['document-branding'] })
    },
    onError: (err: any) => toast.error(err?.message || 'Failed to save'),
  })

  const handleSave = () => saveMutation.mutate(form)

  const set = <K extends keyof DocumentBrandingSettings>(key: K, value: DocumentBrandingSettings[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }))

  const handleWatermarkUpload = async (file: File) => {
    setUploading(true)
    try {
      const path = await uploadWatermarkImage(file)
      set('watermark_image_path', path)
      toast.success('Watermark image uploaded')
    } catch (err: any) {
      toast.error(err?.message || 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-6 h-6 animate-spin text-foreground/40" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl md:text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent flex items-center gap-2">
            <Palette className="w-6 h-6 text-blue-400" />
            Document Branding
          </h1>
          <p className="text-foreground/60 text-sm mt-1">Configure how your PDFs and Excel exports look</p>
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

      {/* Settings Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Column 1 */}
        <div className="space-y-4">
          {/* Header — unified toggles + editable fields */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-1">Header</h3>
            <p className="text-xs text-foreground/50 mb-4">Toggle visibility and override values. Leave fields blank to use your organization profile defaults.</p>
            <div className="space-y-4">
              {/* Logo */}
              <Toggle checked={form.show_logo} onChange={(v) => set('show_logo', v)} label="Show Logo" />

              {/* Organization Name */}
              <div className="border-t border-foreground/5 pt-3">
                <Toggle checked={form.show_org_name} onChange={(v) => set('show_org_name', v)} label="Organization Name" />
                {form.show_org_name && (
                  <input
                    type="text"
                    value={form.header_org_name}
                    onChange={(e) => set('header_org_name', e.target.value)}
                    className="w-full mt-1 px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg outline-none focus:ring-2 focus:ring-blue-500/40 text-sm transition-all"
                    placeholder="Leave blank to use profile name"
                  />
                )}
              </div>

              {/* Address */}
              <div className="border-t border-foreground/5 pt-3">
                <Toggle checked={form.show_address} onChange={(v) => set('show_address', v)} label="Address" />
                {form.show_address && (
                  <textarea
                    value={form.header_address}
                    onChange={(e) => set('header_address', e.target.value)}
                    rows={2}
                    className="w-full mt-1 px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg outline-none focus:ring-2 focus:ring-blue-500/40 text-sm transition-all resize-none"
                    placeholder="Leave blank to use profile address"
                  />
                )}
              </div>

              {/* Contact (Phone + Email) */}
              <div className="border-t border-foreground/5 pt-3">
                <Toggle checked={form.show_contact} onChange={(v) => set('show_contact', v)} label="Contact Info" />
                {form.show_contact && (
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-1">
                    <input
                      type="text"
                      value={form.header_phone}
                      onChange={(e) => set('header_phone', e.target.value)}
                      className="w-full px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg outline-none focus:ring-2 focus:ring-blue-500/40 text-sm transition-all"
                      placeholder="Phone"
                    />
                    <input
                      type="text"
                      value={form.header_email}
                      onChange={(e) => set('header_email', e.target.value)}
                      className="w-full px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg outline-none focus:ring-2 focus:ring-blue-500/40 text-sm transition-all"
                      placeholder="Email"
                    />
                  </div>
                )}
              </div>

              {/* TIN */}
              <div className="border-t border-foreground/5 pt-3">
                <Toggle checked={form.show_tin} onChange={(v) => set('show_tin', v)} label="TIN" />
                {form.show_tin && (
                  <input
                    type="text"
                    value={form.header_tin}
                    onChange={(e) => set('header_tin', e.target.value)}
                    className="w-full mt-1 px-3 py-2 bg-foreground/5 border border-foreground/10 rounded-lg outline-none focus:ring-2 focus:ring-blue-500/40 text-sm transition-all"
                    placeholder="Leave blank to use profile TIN"
                  />
                )}
              </div>
            </div>
          </div>

          {/* Footer Settings */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-3">Footer</h3>
            <div className="divide-y divide-foreground/5">
              <Toggle checked={form.show_footer} onChange={(v) => set('show_footer', v)} label="Show Footer" />
              {form.show_footer && (
                <>
                  <TextRow value={form.footer_text} onChange={(v) => set('footer_text', v)} label="Footer Text" placeholder="Custom footer text" />
                  <TextRow value={form.footer_org_name} onChange={(v) => set('footer_org_name', v)} label="Footer Organization Name" placeholder="Leave blank to use default" />
                  <Toggle checked={form.show_page_numbers} onChange={(v) => set('show_page_numbers', v)} label="Show Page Numbers" />
                  <Toggle checked={form.show_generated_date} onChange={(v) => set('show_generated_date', v)} label="Show Generated Date" />
                </>
              )}
            </div>
          </div>
        </div>

        {/* Column 2 */}
        <div className="space-y-4">
          {/* Watermark Settings */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-3">Watermark</h3>
            <div className="divide-y divide-foreground/5">
              <Toggle checked={form.show_watermark} onChange={(v) => set('show_watermark', v)} label="Show Watermark" />
              {form.show_watermark && (
                <div className="py-3 space-y-3">
                  {/* Type selector */}
                  <div>
                    <label className="block text-sm text-foreground/80 mb-2">Watermark Type</label>
                    <div className="flex rounded-lg bg-foreground/10 p-0.5 w-fit">
                      <button
                        type="button"
                        onClick={() => set('watermark_type', 'text')}
                        className={`px-4 py-1.5 text-xs font-medium rounded-md transition-all flex items-center gap-1.5 ${form.watermark_type === 'text' ? 'bg-blue-600 text-white shadow' : 'text-foreground/60 hover:text-foreground'}`}
                      >
                        <Type className="w-3.5 h-3.5" />
                        Text
                      </button>
                      <button
                        type="button"
                        onClick={() => set('watermark_type', 'image')}
                        className={`px-4 py-1.5 text-xs font-medium rounded-md transition-all flex items-center gap-1.5 ${form.watermark_type === 'image' ? 'bg-blue-600 text-white shadow' : 'text-foreground/60 hover:text-foreground'}`}
                      >
                        <Image className="w-3.5 h-3.5" />
                        Image
                      </button>
                    </div>
                  </div>

                  {form.watermark_type === 'text' ? (
                    <TextRow value={form.watermark_text} onChange={(v) => set('watermark_text', v)} label="Watermark Text" placeholder="Leave empty to use organization name" />
                  ) : (
                    <div>
                      <label className="block text-sm text-foreground/80 mb-2">Watermark Image</label>
                      {form.watermark_image_path ? (
                        <div className="flex items-center gap-3 p-3 bg-foreground/5 rounded-lg border border-foreground/10">
                          <img src={form.watermark_image_path} alt="watermark" className="w-16 h-16 object-contain rounded bg-white p-1" />
                          <div className="flex-1 min-w-0">
                            <p className="text-xs text-foreground/60 truncate">{form.watermark_image_path}</p>
                          </div>
                          <button
                            type="button"
                            onClick={() => set('watermark_image_path', '')}
                            className="p-1.5 rounded-lg hover:bg-foreground/10 text-foreground/40 hover:text-red-400 transition-colors"
                          >
                            <X className="w-4 h-4" />
                          </button>
                        </div>
                      ) : (
                        <button
                          type="button"
                          onClick={() => fileRef.current?.click()}
                          disabled={uploading}
                          className="w-full flex items-center justify-center gap-2 px-4 py-6 border-2 border-dashed border-foreground/15 rounded-lg hover:border-blue-500/40 hover:bg-blue-500/5 transition-all text-foreground/50"
                        >
                          {uploading ? (
                            <Loader2 className="w-5 h-5 animate-spin" />
                          ) : (
                            <Upload className="w-5 h-5" />
                          )}
                          <span className="text-sm">{uploading ? 'Uploading...' : 'Upload watermark image'}</span>
                        </button>
                      )}
                      <input
                        ref={fileRef}
                        type="file"
                        accept="image/png,image/jpeg,image/gif,image/webp"
                        className="hidden"
                        onChange={(e) => {
                          const file = e.target.files?.[0]
                          if (file) handleWatermarkUpload(file)
                          e.target.value = ''
                        }}
                      />
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Style Settings */}
          <div className="rounded-2xl border border-foreground/10 p-5 bg-foreground/[0.03]">
            <h3 className="text-sm font-semibold text-foreground/70 uppercase tracking-wider mb-3">Style</h3>
            <div className="divide-y divide-foreground/5">
              <ColorInput value={form.primary_color} onChange={(v) => set('primary_color', v)} label="Primary Color" />
              <ColorInput value={form.header_text_color} onChange={(v) => set('header_text_color', v)} label="Header Text Color" />
              {/* Searchable font picker */}
              <div className="py-2.5">
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
          </div>
        </div>
      </div>

      {/* Preview Modal */}
      <Modal open={previewOpen} onClose={() => setPreviewOpen(false)} title="Document Preview" size="full">
        {/* Top-level tab switcher: PDF | Excel */}
        <div className="flex justify-center mb-4">
          <div className="flex rounded-lg bg-foreground/10 p-0.5">
            <button
              onClick={() => setPreviewTab('pdf')}
              className={`px-5 py-2 text-sm font-medium rounded-md transition-all flex items-center gap-2 ${previewTab === 'pdf' ? 'bg-blue-600 text-white shadow' : 'text-foreground/60 hover:text-foreground'}`}
            >
              <FileText className="w-4 h-4" />
              PDF
            </button>
            <button
              onClick={() => setPreviewTab('excel')}
              className={`px-5 py-2 text-sm font-medium rounded-md transition-all flex items-center gap-2 ${previewTab === 'excel' ? 'bg-green-600 text-white shadow' : 'text-foreground/60 hover:text-foreground'}`}
            >
              <Table className="w-4 h-4" />
              Excel
            </button>
          </div>
        </div>

        {/* PDF document sub-tabs */}
        {previewTab === 'pdf' && (
          <div className="flex justify-center mb-4">
            <div className="flex gap-1 border-b border-foreground/10">
              {([
                { key: 'salary-slip', label: 'Salary Slip' },
                { key: 'sample', label: 'Sample Document' },
              ] as const).map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setPdfDoc(tab.key)}
                  className={`px-4 py-1.5 text-xs font-medium border-b-2 transition-all -mb-px ${pdfDoc === tab.key ? 'border-blue-500 text-blue-400' : 'border-transparent text-foreground/50 hover:text-foreground/70'}`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>
        )}

        <div className="flex justify-center">
          <div className="w-full max-w-2xl">
            {previewTab === 'pdf' ? (
              pdfDoc === 'salary-slip' ? <SalarySlipPreview form={form} /> : <SamplePDFPreview form={form} />
            ) : (
              <ExcelPreview form={form} />
            )}
          </div>
        </div>
      </Modal>
    </div>
  )
}
