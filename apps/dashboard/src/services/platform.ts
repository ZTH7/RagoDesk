import { request } from './client'
import type { ListParams } from './types'

export type TenantItem = {
  id: string
  name: string
  type: string
  plan: string
  status: string
  created_at: string
}

export type PlatformAdminItem = {
  id: string
  name: string
  email: string
  status: string
  created_at: string
}

export type PlatformRoleItem = {
  id: string
  name: string
  created_at?: string
}

export type PlatformPermissionItem = {
  code: string
  scope: string
  description?: string
  created_at?: string
}

export type CreateTenantInput = {
  name: string
  type: string
  plan: string
  status: string
}

export type CreatePlatformAdminInput = {
  name: string
  email: string
  phone?: string
  status: string
  password?: string
  send_invite?: boolean
  invite_base_url?: string
}

export type CreatePlatformRoleInput = {
  name: string
}

export type CreatePermissionInput = {
  code: string
  description: string
  scope: string
}

export const platformApi = {
  listTenants(params?: ListParams) {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: TenantItem[] }>(`/platform/v1/tenants${suffix}`)
  },
  createTenant(payload: CreateTenantInput) {
    return request<{ tenant: TenantItem }>('/platform/v1/tenants', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  getTenant(id: string) {
    return request<{ tenant: TenantItem }>(`/platform/v1/tenants/${id}`)
  },
  listAdmins(params?: ListParams) {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: PlatformAdminItem[] }>(`/platform/v1/admins${suffix}`)
  },
  getAdmin(id: string) {
    return request<{ admin: PlatformAdminItem }>(`/platform/v1/admins/${id}`)
  },
  createAdmin(payload: CreatePlatformAdminInput) {
    return request<{ admin: PlatformAdminItem; invite_link?: string }>('/platform/v1/admins', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  listRoles(params?: ListParams) {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: PlatformRoleItem[] }>(`/platform/v1/roles${suffix}`)
  },
  getRole(id: string) {
    return request<{ role: PlatformRoleItem }>(`/platform/v1/roles/${id}`)
  },
  createRole(payload: CreatePlatformRoleInput) {
    return request<{ role: PlatformRoleItem }>('/platform/v1/roles', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  assignAdminRole(adminId: string, roleId: string) {
    return request<void>(`/platform/v1/admins/${adminId}/roles`, {
      method: 'POST',
      body: JSON.stringify({ admin_id: adminId, role_id: roleId }),
    })
  },
  listAdminRoles(adminId: string) {
    return request<{ items: PlatformRoleItem[] }>(`/platform/v1/admins/${adminId}/roles`)
  },
  removeAdminRole(adminId: string, roleId: string) {
    return request<void>(`/platform/v1/admins/${adminId}/roles/${roleId}`, {
      method: 'DELETE',
    })
  },
  listRolePermissions(roleId: string) {
    return request<{ items: PlatformPermissionItem[] }>(`/platform/v1/roles/${roleId}/permissions`)
  },
  assignRolePermissions(roleId: string, permissionCodes: string[]) {
    return request<void>(`/platform/v1/roles/${roleId}/permissions`, {
      method: 'POST',
      body: JSON.stringify({ role_id: roleId, permission_codes: permissionCodes }),
    })
  },
  listPermissions(params?: ListParams & { scope?: string }) {
    const query = new URLSearchParams()
    if (params?.limit) query.set('limit', String(params.limit))
    if (params?.offset) query.set('offset', String(params.offset))
    if (params?.scope) query.set('scope', params.scope)
    const suffix = query.toString() ? `?${query.toString()}` : ''
    return request<{ items: PlatformPermissionItem[] }>(`/platform/v1/permissions${suffix}`)
  },
  createPermission(payload: CreatePermissionInput) {
    return request<{ permission: PlatformPermissionItem }>('/platform/v1/permissions', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
}
