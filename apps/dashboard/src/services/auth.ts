import { request } from './client'

export type AuthProfile = {
  subject_id: string
  tenant_id?: string
  account: string
  name?: string
  roles: string[]
}

export type AuthResponse = {
  token: string
  expires_at?: string
  profile: AuthProfile
}

export const authApi = {
  consoleLogin(payload: { account: string; password: string; tenant_id?: string }) {
    return request<AuthResponse>('/console/v1/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  consoleRegister(payload: {
    tenant_name: string
    tenant_type?: string
    admin_name?: string
    email?: string
    phone?: string
    password: string
  }) {
    return request<AuthResponse>('/console/v1/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  platformLogin(payload: { account: string; password: string }) {
    return request<AuthResponse>('/platform/v1/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
}
