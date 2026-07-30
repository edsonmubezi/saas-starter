import React from 'react'
import { NavLink } from 'react-router-dom'
import {
  Building2, Users, Shield,
  FileSearch, ShieldAlert, FileText
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

interface HQAdminMenuProps {
  allow: (permissions?: string[], requiredAll?: string[]) => boolean
  ShowIf: ({ any, all, children }: { any?: string[]; all?: string[]; children: React.ReactNode }) => React.ReactNode
  openSection: string | null
  toggleSection: (sectionId: string) => void
}

export default function HQAdminMenu({ allow, ShowIf, openSection, toggleSection }: HQAdminMenuProps) {
  return (
    <>
      {/* People - Super Admin */}
      {allow(['admin.user.view']) && (
        <Collapsible
          title="Users"
          icon={<Users size={16}/>}
          isOpen={openSection === 'people'}
          onToggle={() => toggleSection('people')}
        >
          <ShowIf any={['admin.user.view']}><NavItem to="/superadmin-users">View All Users</NavItem></ShowIf>
        </Collapsible>
      )}

      {/* Access Control */}
      {(allow(['admin.role.view']) || allow(['admin.permission.view'])) && (
        <Collapsible
          title="Access & Roles"
          icon={<Shield size={16}/>}
          isOpen={openSection === 'access-control'}
          onToggle={() => toggleSection('access-control')}
        >
          <ShowIf any={['admin.role.view']}><NavItem to="/manage-roles">Manage Roles</NavItem></ShowIf>
          <ShowIf any={['admin.permission.view']}><NavItem to="/view-permission">Manage Permissions</NavItem></ShowIf>
          <ShowIf any={['admin.role.assign']}><NavItem to="/assign-user-roles">Assign Roles</NavItem></ShowIf>
        </Collapsible>
      )}

      {/* Manage Organization */}
      {allow(['admin.organization.create']) && (
        <Collapsible
          title="Organizations"
          icon={<Building2 size={16}/>}
          isOpen={openSection === 'organization'}
          onToggle={() => toggleSection('organization')}
        >
          <ShowIf any={['admin.organization.view']}>
            <NavItem to="/view-organisations">View All</NavItem>
          </ShowIf>
        </Collapsible>
      )}

      {/* Audit Logs - Admin level */}
      {(allow(['admin.audit.view'])) && (
        <Collapsible
          title="Audit Trail"
          icon={<FileSearch size={16}/>}
          isOpen={openSection === 'audit'}
          onToggle={() => toggleSection('audit')}
        >
          <ShowIf any={['admin.audit.view']}><NavItem to="/audit/events">View Events</NavItem></ShowIf>
        </Collapsible>
      )}

      {/* Security - Admin level */}
      {(allow(['admin.security.view'])) && (
        <Collapsible
          title="Security"
          icon={<ShieldAlert size={16}/>}
          isOpen={openSection === 'security'}
          onToggle={() => toggleSection('security')}
        >
          <ShowIf any={['admin.security.view']}><NavItem to="/security/dashboard">Overview</NavItem></ShowIf>
          <ShowIf any={['admin.security.view']}><NavItem to="/security/events">View Events</NavItem></ShowIf>
        </Collapsible>
      )}

      {/* System Logs - Admin level */}
      {(allow(['admin.logs.view'])) && (
        <Collapsible
          title="System Logs"
          icon={<FileText size={16}/>}
          isOpen={openSection === 'logs'}
          onToggle={() => toggleSection('logs')}
        >
          <ShowIf any={['admin.logs.view']}><NavItem to="/logs/application">App Logs</NavItem></ShowIf>
          <ShowIf any={['admin.logs.view']}><NavItem to="/logs/access">Access Logs</NavItem></ShowIf>
        </Collapsible>
      )}
    </>
  )
}
