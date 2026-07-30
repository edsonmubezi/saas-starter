import React, { useState } from 'react'
import { NavLink } from 'react-router-dom'
import {
  Users, Mail, Search, Sparkles
} from 'lucide-react'
import clsx from 'clsx'

function NavItem({ to, children }: { to: string; children: React.ReactNode }) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        clsx(
          'block rounded-lg px-3 py-2 text-sm transition-all',
          isActive
            ? 'bg-blue-500/10 text-blue-300 font-medium border-l-2 border-blue-500'
            : 'text-foreground/70 hover:bg-foreground/10 hover:text-foreground/90'
        )
      }
      end
    >
      {children}
    </NavLink>
  )
}

function Collapsible({
  title, icon, children, isOpen, onToggle
}: {
  title: string;
  icon?: React.ReactNode;
  children: React.ReactNode;
  isOpen: boolean;
  onToggle: () => void;
}) {
  return (
    <div>
      <button
        onClick={onToggle}
        className={clsx(
          'w-full flex items-center justify-between rounded-lg px-3 py-2 text-sm transition-all',
          isOpen
            ? 'text-foreground font-medium bg-foreground/5'
            : 'text-foreground/70 hover:bg-foreground/10 hover:text-foreground/90'
        )}
        aria-expanded={isOpen}
      >
        <span className="flex items-center gap-2">{icon}{title}</span>
        <span className={clsx('transition-transform text-xs', isOpen ? 'rotate-180' : '')}>▼</span>
      </button>
      <div className={clsx('grid gap-1 pl-6', isOpen ? 'mt-1' : 'hidden')}>
        {children}
      </div>
    </div>
  )
}

function SectionHeader({ label, visible }: { label: string; visible?: boolean }) {
  if (visible === false) return null
  return (
    <p className="text-[10px] uppercase tracking-wider text-foreground/30 font-medium px-3 mt-5 mb-2 first:mt-0">
      {label}
    </p>
  )
}

interface OrgAdminMenuProps {
  allow: (permissions?: string[], requiredAll?: string[]) => boolean
  ShowIf: ({ any, all, children }: { any?: string[]; all?: string[]; children: React.ReactNode }) => React.ReactNode
  openSection: string | null
  toggleSection: (sectionId: string) => void
}

export default function OrgAdminMenu({ allow, ShowIf, openSection, toggleSection }: OrgAdminMenuProps) {
  const [searchTerm, setSearchTerm] = useState('')

  const q = searchTerm.toLowerCase().trim()
  const match = (label: string) => !q || label.toLowerCase().includes(q)
  const sectionMatch = (...labels: string[]) => !q || labels.some(l => l.toLowerCase().includes(q))

  // When searching, force sections open
  const isOpen = (sectionId: string, labels: string[]) =>
    q ? labels.some(l => l.toLowerCase().includes(q)) : openSection === sectionId
  const onToggle = (sectionId: string) => {
    if (!q) toggleSection(sectionId)
  }

  // --- Group visibility ---
  const peopleLabels = ['View Users', 'Assign Roles', 'Create Role']
  const settingsLabels = ['Email Settings', 'Document Branding', 'Email Branding', 'Knowledge Base']

  const showPeople = sectionMatch(...peopleLabels)
  const showSettings = sectionMatch(...settingsLabels)

  return (
    <>
      {/* Search Filter */}
      <div className="relative px-1 mb-3">
        <Search size={14} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-foreground/30" />
        <input
          type="text"
          placeholder="Search menu..."
          value={searchTerm}
          onChange={e => setSearchTerm(e.target.value)}
          className="w-full pl-8 pr-3 py-2 rounded-lg bg-foreground/5 border border-foreground/10 text-sm text-foreground placeholder:text-foreground/30 focus:outline-none focus:border-blue-500/40 focus:bg-foreground/[0.07] transition-all"
        />
      </div>

      {/* --- PEOPLE --- */}
      <SectionHeader label="People" visible={showPeople} />

      {/* Manage Users */}
      {sectionMatch('View Users', 'Assign Roles', 'Create Role') &&
        (allow(['tenant.user.view']) || allow(['tenant.role.assign']) || allow(['tenant.role.create'])) && (
        <Collapsible
          title="Manage Users"
          icon={<Users size={16}/>}
          isOpen={isOpen('user-management', ['View Users', 'Assign Roles', 'Create Role'])}
          onToggle={() => onToggle('user-management')}
        >
          {match('View Users') && <ShowIf any={['tenant.user.view']}><NavItem to="/org-users">View Users</NavItem></ShowIf>}
          {match('Assign Roles') && <ShowIf any={['tenant.role.assign']}><NavItem to="/org-assign-roles">Assign Roles</NavItem></ShowIf>}
          {match('Create Role') && <ShowIf any={['tenant.role.create']}><NavItem to="/org-create-role">Create Role</NavItem></ShowIf>}
        </Collapsible>
      )}

      {/* --- SETTINGS --- */}
      <SectionHeader label="Settings" visible={showSettings} />

      {/* Communication & Branding */}
      {sectionMatch('Email Settings', 'Document Branding', 'Email Branding') &&
        (allow(['tenant.broadcast.view']) || allow(['tenant.document_branding.manage']) || allow(['tenant.email_branding.manage'])) && (
        <Collapsible
          title="Communication"
          icon={<Mail size={16}/>}
          isOpen={isOpen('communication', ['Email Settings', 'Document Branding', 'Email Branding'])}
          onToggle={() => onToggle('communication')}
        >
          {match('Email Settings') && <ShowIf any={['tenant.broadcast.view']}><NavItem to="/email-config">Email Settings</NavItem></ShowIf>}
          {/* Branding */}
          {sectionMatch('Document Branding', 'Email Branding') && (
            <p className="text-[10px] uppercase tracking-wider text-foreground/30 font-medium px-1 mt-3 mb-1">Branding</p>
          )}
          {match('Document Branding') && <ShowIf any={['tenant.document_branding.manage']}><NavItem to="/document-branding">Document Branding</NavItem></ShowIf>}
          {match('Email Branding') && <ShowIf any={['tenant.email_branding.manage']}><NavItem to="/email-branding">Email Branding</NavItem></ShowIf>}
        </Collapsible>
      )}

      {/* Knowledge Base */}
      {match('Knowledge Base') && (
        <ShowIf any={['tenant.knowledge_base.view', 'tenant.knowledge_base.manage']}>
          <NavItem to="/knowledge-base">
            <span className="flex items-center gap-2"><Sparkles size={16}/> Knowledge Base</span>
          </NavItem>
        </ShowIf>
      )}
    </>
  )
}
