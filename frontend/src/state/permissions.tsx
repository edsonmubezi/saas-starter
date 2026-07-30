import React from 'react';

export function usePerms(user: any) {
  const permSet = new Set<string>(Array.isArray(user?.permissions) ? user.permissions : []);

  const allow = (requiredAny?: string[], requiredAll?: string[]) => {
    if (!permSet.size) return false;
    if (requiredAll?.length && !requiredAll.every(p => permSet.has(p))) return false;
    if (requiredAny?.length && !requiredAny.some(p => permSet.has(p))) return false;
    return true;
  };

  // Expose .has() method for direct permission checking
  const has = (permission: string) => permSet.has(permission);

  const ShowIf = ({ any, all, children }: { any?: string[]; all?: string[]; children: React.ReactNode }) =>
    allow(any, all) ? <>{children}</> : null;

  // Admin (SuperAdmin/HQ) permission helpers
  const canManageUsers = allow(['admin.user.view', 'admin.user.create', 'admin.user.edit', 'admin.user.delete']);
  const canManageRoles = allow(['admin.role.view', 'admin.role.create', 'admin.role.edit', 'admin.role.delete']);
  const canManagePermissions = allow(['admin.permission.view', 'admin.permission.edit']);
  const canManageOrganizations = allow(['admin.organization.view', 'admin.organization.create', 'admin.organization.edit', 'admin.organization.delete']);

  // Tenant (Organization-level) permission helpers
  const canManageTenantUsers = allow(['tenant.user.view', 'tenant.user.create', 'tenant.user.edit', 'tenant.user.delete']);
  const canManageTenantRoles = allow(['tenant.role.view', 'tenant.role.create', 'tenant.role.edit', 'tenant.role.delete']);

  return {
    allow,
    has,
    ShowIf,
    // Admin (HQ) permissions
    canManageUsers,
    canManageRoles,
    canManagePermissions,
    canManageOrganizations,
    // Tenant (Organization-level) permissions
    canManageTenantUsers,
    canManageTenantRoles,
  };
}
