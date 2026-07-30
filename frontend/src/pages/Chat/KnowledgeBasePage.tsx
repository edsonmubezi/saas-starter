import { useCallback, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Pencil, Plus, Trash2, Sparkles, Loader2 } from 'lucide-react'
import toast from 'react-hot-toast'
import Modal from '../../ui/Modal'
import DataTable, { type Column } from '../../ui/DataTable'
import useDebounce from '../../state/useDebounce'
import { useConfirm } from '../../ui/ConfirmProvider'
import {
  getKnowledgeArticles,
  createKnowledgeArticle,
  updateKnowledgeArticle,
  deleteKnowledgeArticle,
  seedKnowledge,
  KNOWLEDGE_CATEGORIES,
  type KnowledgeArticle,
} from '../../utils/knowledge'

const categoryColors: Record<string, string> = {
  module_guide: 'bg-blue-500/15 text-blue-400 border-blue-500/30',
  how_to: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30',
  troubleshooting: 'bg-amber-500/15 text-amber-400 border-amber-500/30',
  faq: 'bg-purple-500/15 text-purple-400 border-purple-500/30',
  policy: 'bg-rose-500/15 text-rose-400 border-rose-500/30',
}

function CategoryBadge({ category }: { category: string }) {
  const label = KNOWLEDGE_CATEGORIES.find(c => c.value === category)?.label ?? category
  const color = categoryColors[category] ?? 'bg-foreground/10 text-foreground/60'
  return (
    <span className={`inline-block rounded-full border px-2.5 py-0.5 text-[11px] font-medium ${color}`}>
      {label}
    </span>
  )
}

interface FormValues {
  title: string
  category: string
  content: string
}

const emptyForm: FormValues = { title: '', category: 'how_to', content: '' }

export default function KnowledgeBasePage() {
  const confirm = useConfirm()
  const qc = useQueryClient()

  const [search, setSearch] = useState('')
  const debounced = useDebounce(search, 300)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [openCreate, setOpenCreate] = useState(false)
  const [editing, setEditing] = useState<KnowledgeArticle | null>(null)
  const [form, setForm] = useState<FormValues>(emptyForm)
  const [submitting, setSubmitting] = useState(false)
  const [seeding, setSeeding] = useState(false)

  const { data, isFetching: loading } = useQuery({
    queryKey: ['knowledge-articles', page, pageSize],
    queryFn: () => getKnowledgeArticles({ page, page_size: pageSize }),
    staleTime: 30_000,
  })

  const articles = data?.data ?? []
  const total = data?.total ?? 0

  // Client-side search filter
  const filtered = useMemo(() => {
    if (!debounced) return articles
    const q = debounced.toLowerCase()
    return articles.filter(
      a => a.title.toLowerCase().includes(q) || a.category.toLowerCase().includes(q),
    )
  }, [articles, debounced])

  const refresh = useCallback(
    async (msg?: string) => {
      await qc.invalidateQueries({ queryKey: ['knowledge-articles'] })
      if (msg) toast.success(msg)
    },
    [qc],
  )

  const handleCreate = async () => {
    if (!form.title.trim() || !form.content.trim()) {
      toast.error('Title and content are required')
      return
    }
    setSubmitting(true)
    try {
      await createKnowledgeArticle({
        title: form.title.trim(),
        content: form.content.trim(),
        category: form.category,
      })
      setOpenCreate(false)
      setForm(emptyForm)
      await refresh('Article created')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to create article')
    } finally {
      setSubmitting(false)
    }
  }

  const handleEdit = async () => {
    if (!editing) return
    if (!form.title.trim() || !form.content.trim()) {
      toast.error('Title and content are required')
      return
    }
    setSubmitting(true)
    try {
      await updateKnowledgeArticle(editing.id, {
        title: form.title.trim(),
        content: form.content.trim(),
        category: form.category,
      })
      setEditing(null)
      setForm(emptyForm)
      await refresh('Article updated')
    } catch (e: any) {
      toast.error(e?.message || 'Failed to update article')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (row: KnowledgeArticle) => {
    const ok = await confirm({
      title: 'Delete Article',
      description: `Are you sure you want to delete "${row.title}"? This action cannot be undone.`,
      confirmText: 'Delete',
      cancelText: 'Cancel',
      variant: 'destructive',
    })
    if (!ok) return
    try {
      await deleteKnowledgeArticle(row.id)
      await refresh('Article deleted')
    } catch (e: any) {
      toast.error(e?.message || 'Delete failed')
    }
  }

  const handleSeed = async () => {
    const ok = await confirm({
      title: 'Seed Default Knowledge',
      description:
        'This will populate the knowledge base with default articles about this platform. This uses the OpenAI API to generate embeddings (small cost). Continue?',
      confirmText: 'Seed Knowledge',
      cancelText: 'Cancel',
    })
    if (!ok) return
    setSeeding(true)
    try {
      await seedKnowledge()
      await refresh('Knowledge base seeded successfully')
    } catch (e: any) {
      toast.error(e?.message || 'Seeding failed')
    } finally {
      setSeeding(false)
    }
  }

  const openEditModal = (article: KnowledgeArticle) => {
    setForm({ title: article.title, category: article.category, content: article.content })
    setEditing(article)
  }

  const columns: Column<KnowledgeArticle>[] = useMemo(
    () => [
      {
        key: 'title',
        title: 'Title',
        sortable: true,
        render: r => <span className="font-medium">{r.title}</span>,
      },
      {
        key: 'category',
        title: 'Category',
        sortable: true,
        render: r => <CategoryBadge category={r.category} />,
      },
      {
        key: 'created_at',
        title: 'Created',
        sortable: true,
        render: r => (r.created_at ? new Date(r.created_at).toLocaleDateString() : '—'),
      },
      {
        key: 'actions' as any,
        title: 'Actions',
        className: 'text-right w-[120px]',
        render: o => (
          <div className="flex h-10 items-center justify-end gap-1">
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
              title="Edit"
              onClick={() => openEditModal(o)}
            >
              <Pencil size={16} />
            </button>
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10 text-rose-300"
              title="Delete"
              onClick={() => handleDelete(o)}
            >
              <Trash2 size={16} />
            </button>
          </div>
        ),
      },
    ],
    [],
  )

  const formFields = (
    <div className="space-y-4">
      <div>
        <label className="mb-1 block text-sm font-medium text-foreground/70">Title</label>
        <input
          className="input w-full"
          placeholder="Article title"
          value={form.title}
          onChange={e => setForm(f => ({ ...f, title: e.target.value }))}
        />
      </div>
      <div>
        <label className="mb-1 block text-sm font-medium text-foreground/70">Category</label>
        <select
          className="input w-full"
          value={form.category}
          onChange={e => setForm(f => ({ ...f, category: e.target.value }))}
        >
          {KNOWLEDGE_CATEGORIES.map(c => (
            <option key={c.value} value={c.value}>
              {c.label}
            </option>
          ))}
        </select>
      </div>
      <div>
        <label className="mb-1 block text-sm font-medium text-foreground/70">Content</label>
        <textarea
          className="input w-full min-h-[200px] resize-y"
          placeholder="Article content (markdown supported)..."
          value={form.content}
          onChange={e => setForm(f => ({ ...f, content: e.target.value }))}
        />
      </div>
    </div>
  )

  return (
    <div className="grid gap-4">
      {/* Toolbar */}
      <div className="card flex flex-wrap items-center justify-between gap-3 bg-foreground/5 border-foreground/10">
        <input
          className="input max-w-xs"
          placeholder="Search articles…"
          value={search}
          onChange={e => {
            setSearch(e.target.value)
            setPage(1)
          }}
        />
        <div className="flex items-center gap-2">
          <button
            className="btn flex items-center gap-2 bg-purple-500/15 text-purple-300 hover:bg-purple-500/25 border-purple-500/30"
            onClick={handleSeed}
            disabled={seeding}
          >
            {seeding ? <Loader2 size={16} className="animate-spin" /> : <Sparkles size={16} />}
            {seeding ? 'Seeding…' : 'Seed Default Knowledge'}
          </button>
          <button
            className="btn flex items-center gap-2"
            onClick={() => {
              setForm(emptyForm)
              setOpenCreate(true)
            }}
          >
            <Plus size={16} /> Add Article
          </button>
        </div>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        rows={filtered}
        loading={loading}
        showIndex
        pagination={{
          page,
          pageSize,
          total: debounced ? filtered.length : total,
          onPageChange: setPage,
          onPageSizeChange: s => {
            setPageSize(s)
            setPage(1)
          },
        }}
      />

      {/* Create modal */}
      <Modal open={openCreate} onClose={() => setOpenCreate(false)} title="Create Knowledge Article" size="lg">
        {formFields}
        <div className="mt-4 flex justify-end gap-2">
          <button className="btn-secondary" onClick={() => setOpenCreate(false)}>
            Cancel
          </button>
          <button className="btn" onClick={handleCreate} disabled={submitting}>
            {submitting ? 'Creating…' : 'Create'}
          </button>
        </div>
      </Modal>

      {/* Edit modal */}
      <Modal open={!!editing} onClose={() => setEditing(null)} title="Edit Knowledge Article" size="lg">
        {formFields}
        <div className="mt-4 flex justify-end gap-2">
          <button className="btn-secondary" onClick={() => setEditing(null)}>
            Cancel
          </button>
          <button className="btn" onClick={handleEdit} disabled={submitting}>
            {submitting ? 'Saving…' : 'Save Changes'}
          </button>
        </div>
      </Modal>
    </div>
  )
}
