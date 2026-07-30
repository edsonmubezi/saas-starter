// Organization type - determines import template and other org-specific features
export type OrganizationType = 'single_company' | 'multiple_company' | 'multi_branch' | 'outsourcing';

// Canonical shape for the authenticated user returned to the app
export interface Me {
  // In your payload, id is an encrypted string (not a number)
  id: string;

  email: string | null;
  fullname: string | null;

  // Optional status/meta
  active_status?: boolean;
  created_at?: string;
  updated_at?: string;
  photo?: string;

  // You already have this numeric field in the payload
  organization_id?: number;

  // Account setup flags
  must_change_password?: boolean;
  email_verified?: boolean;
  phone_number?: string | null;

  // Organization type from org settings
  organization_type?: OrganizationType;

  // Nested organization object from /users/me
  organization?: {
    id: string;
    name: string;
    address?: string;
    contact_person?: string;
    phone_number?: string;
    created_at?: string;
    updated_at?: string;
  };

  // Role object (name only needed in UI; id useful for debugging)
  role?: {
    id: number;
    name: string;
  };

  // 🔑 Effective permissions (flattened from role.permissions[].name)
  permissions: string[];
}
