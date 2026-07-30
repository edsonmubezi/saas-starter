import React, { useCallback, useMemo, useState } from 'react'
import { useQuery, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Eye, Pencil, Plus, Settings } from 'lucide-react'
import toast from 'react-hot-toast'

import {
  getOrganizations,
  getOrganization,
  createOrganization,
  updateOrganization,
  updateOrganizationWithLogo,
  type OrganizationsResult,
  type OrganizationListItem,
  type CreateOrganizationInput,
  type UpdateOrganizationInput,
} from '../../utils/orgs'

import useDebounce from '../../state/useDebounce'
import DataTable, { type Column } from '../../ui/DataTable'
import Modal from '../../ui/Modal'
import OrgsForm, { type OrgFormValues } from './OrgsForm'

export default function OrgsPage() {
  // ── Local state ──────────────────────────────────────────────────────────────
  const [search, setSearch] = useState('')
  const debounced = useDebounce(search, 300)

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  const [openCreate, setOpenCreate] = useState(false)
  const [editing, setEditing] = useState<OrganizationListItem | null>(null)
  const [viewing, setViewing] = useState<OrganizationListItem | null>(null)

  const qc = useQueryClient()
  const navigate = useNavigate()

  // ── Query params / key ───────────────────────────────────────────────────────
  const params = useMemo(
    () => ({
      q: debounced || undefined,
      page,
      page_size: pageSize,
    }),
    [debounced, page, pageSize],
  )
  const queryKey = useMemo(() => ['orgs', params] as const, [params])

  // ── Query ────────────────────────────────────────────────────────────────────
  const { data, isFetching: loading } = useQuery<OrganizationsResult>({
    queryKey,
    queryFn: () => getOrganizations(params),
    placeholderData: keepPreviousData,
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const pageFromServer = data?.page ?? page
  const pageSizeFromServer = data?.page_size ?? pageSize

  // ── Helpers ──────────────────────────────────────────────────────────────────
  const refresh = useCallback(
    async (msg?: string) => {
      await qc.invalidateQueries({ queryKey })
      if (msg) toast.success(msg)
    },
    [qc, queryKey],
  )

  const handleCreate = async (values: OrgFormValues) => {
    const payload: CreateOrganizationInput = {
      name: values.name.trim(),
      phone_number: values.phone_number,
      address: values.address,
      contact_person: values.contact_person,
    }
    await createOrganization(payload)
    setOpenCreate(false)
    setPage(1)
    await refresh('Organization created')
  }

  const handleEditSubmit = async (values: OrgFormValues) => {
    if (!editing) return

    // Use multipart endpoint if logo file is provided
    if (values.logo) {
      await updateOrganizationWithLogo(editing.id, {
        name: values.name.trim(),
        phone_number: values.phone_number,
        address: values.address,
        contact_person: values.contact_person,
        email: values.email,
        tin: values.tin,
        registration_number: values.registration_number,
        logo: values.logo,
      })
    } else {
      // Use regular JSON endpoint
      const payload: UpdateOrganizationInput = {
        name: values.name.trim(),
        phone_number: values.phone_number,
        address: values.address,
        contact_person: values.contact_person,
        logo_url: values.logo_url,
        email: values.email,
        tin: values.tin,
        registration_number: values.registration_number,
      }
      await updateOrganization(editing.id, payload)
    }

    setEditing(null)
    await refresh('Organization updated')
  }

  // ── Columns ──────────────────────────────────────────────────────────────────
  const columns: Column<OrganizationListItem & { actions?: React.ReactNode }>[] = useMemo(
    () => [
      {
        key: 'logo_url',
        title: 'Logo',
        width: 60,
        render: r => r.logo_url ? (
          <div className="w-10 h-10 border border-foreground/20 rounded overflow-hidden bg-foreground/5 flex items-center justify-center">
            <img
              src={r.logo_url}
              alt="Logo"
              className="max-w-full max-h-full object-contain"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          </div>
        ) : '—'
      },
      { key: 'name', title: 'Name', sortable: true, render: r => r.name || '—' },
      { key: 'email', title: 'Email', sortable: true, render: r => r.email || '—' },
      { key: 'phone_number', title: 'Phone', sortable: true, render: r => r.phone_number || '—' },
      { key: 'tin', title: 'TIN', sortable: true, render: r => r.tin || '—' },
      { key: 'contact_person', title: 'Contact', sortable: true, render: r => r.contact_person || '—' },
      {
        key: 'created_at',
        title: 'Created',
        sortable: true,
        render: r => (r.created_at ? new Date(r.created_at).toLocaleDateString() : '—'),
      },
      {
        key: 'actions',
        title: 'Actions',
        className: 'text-right w-[160px]',
        render: o => (
          <div className="flex h-10 items-center justify-end gap-1">
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
              title="View"
              onClick={async () => {
                try {
                  const fresh = await getOrganization(o.id)
                  setViewing(fresh)
                } catch {
                  setViewing(o) // fallback to row
                }
              }}
            >
              <Eye size={16} />
            </button>
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-foreground/10"
              title="Edit"
              onClick={() => setEditing(o)}
            >
              <Pencil size={16} />
            </button>
            <button
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg hover:bg-blue-500/20 text-blue-400"
              title="Settings"
              onClick={() => navigate(`/organisations/${o.id}/settings`)}
            >
              <Settings size={16} />
            </button>
          </div>
        ),
      },
    ],
    [],
  )

  // ── Render ───────────────────────────────────────────────────────────────────
  return (
    <div className="grid gap-4">
      {/* Toolbar */}
      <div className="card flex flex-wrap items-center justify-between gap-3 bg-foreground/5 border-foreground/10">
        <input
          className="input max-w-xs"
          placeholder="Search…"
          value={search}
          onChange={e => {
            setSearch(e.target.value)
            setPage(1)
          }}
        />
        <button className="btn flex items-center gap-2" onClick={() => setOpenCreate(true)}>
          <Plus size={16} /> Add organization
        </button>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        rows={Array.isArray(items) ? items : []}
        loading={loading}
        showIndex
        pagination={{
          page: pageFromServer,
          pageSize: pageSizeFromServer,
          total,
          onPageChange: setPage,
          onPageSizeChange: s => {
            setPageSize(s)
            setPage(1)
          },
        }}
      />

      {/* Create modal */}
      <Modal open={openCreate} onClose={() => setOpenCreate(false)} title="Create organization">
        <OrgsForm
          mode="create"
          onSubmit={async v => {
            try {
              await handleCreate(v)
            } catch (e: any) {
              toast.error(e?.message || 'Failed to create')
              throw e
            }
          }}
        />
      </Modal>

      {/* Edit modal */}
      <Modal open={!!editing} onClose={() => setEditing(null)} title="Edit organization">
        {editing && (
          <OrgsForm
            mode="edit"
            initial={{
              name: editing.name,
              phone_number: editing.phone_number,
              address: editing.address,
              contact_person: editing.contact_person,
              email: editing.email,
              tin: editing.tin,
              registration_number: editing.registration_number,
              logo_url: editing.logo_url,
            }}
            onSubmit={async v => {
              try {
                await handleEditSubmit(v)
              } catch (e: any) {
                toast.error(e?.message || 'Failed to update')
                throw e
              }
            }}
          />
        )}
      </Modal>

      {/* View modal */}
      <Modal open={!!viewing} onClose={() => setViewing(null)} title="Organization details">
        {viewing && (
          <div className="space-y-4">
            {viewing.logo_url && (
              <div className="flex justify-center">
                <div className="w-32 h-32 border border-foreground/20 rounded-lg overflow-hidden bg-foreground/5 flex items-center justify-center">
                  <img
                    src={viewing.logo_url}
                    alt="Organization logo"
                    className="max-w-full max-h-full object-contain"
                    onError={(e) => {
                      (e.target as HTMLImageElement).style.display = 'none';
                    }}
                  />
                </div>
              </div>
            )}
            <div className="text-sm">
              <span className="text-foreground/60">Name:</span>{' '}
              <span className="font-medium">{viewing.name || '—'}</span>
            </div>
            <div className="text-sm">
              <span className="text-foreground/60">Email:</span>{' '}
              <span className="font-medium">{viewing.email || '—'}</span>
            </div>
            <div className="text-sm">
              <span className="text-foreground/60">Phone:</span>{' '}
              <span className="font-medium">{viewing.phone_number || '—'}</span>
            </div>
            <div className="text-sm">
              <span className="text-foreground/60">Address:</span>{' '}
              <span className="font-medium">{viewing.address || '—'}</span>
            </div>
            <div className="text-sm">
              <span className="text-foreground/60">Contact person:</span>{' '}
              <span className="font-medium">{viewing.contact_person || '—'}</span>
            </div>
            <div className="text-sm">
              <span className="text-foreground/60">TIN:</span>{' '}
              <span className="font-medium">{viewing.tin || '—'}</span>
            </div>
            <div className="text-sm">
              <span className="text-foreground/60">Registration Number:</span>{' '}
              <span className="font-medium">{viewing.registration_number || '—'}</span>
            </div>
            <div className="text-sm">
              <span className="text-foreground/60">Created:</span>{' '}
              <span className="font-medium">
                {viewing.created_at ? new Date(viewing.created_at).toLocaleString() : '—'}
              </span>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
