const TOKEN_KEY = 'ragodesk_token'
const TENANT_KEY = 'ragodesk_tenant_id'
const PROFILE_KEY = 'ragodesk_profile'
const SCOPE_KEY = 'ragodesk_scope'

export type StoredProfile = {
  subject_id?: string
  tenant_id?: string
  name?: string
  account?: string
  roles?: string[]
  scope?: 'console' | 'platform'
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function getTenantId() {
  return localStorage.getItem(TENANT_KEY)
}

export function setTenantId(tenantId: string) {
  localStorage.setItem(TENANT_KEY, tenantId)
}

export function clearTenantId() {
  localStorage.removeItem(TENANT_KEY)
}

export function setProfile(profile: StoredProfile) {
  localStorage.setItem(PROFILE_KEY, JSON.stringify(profile))
  if (profile.tenant_id) {
    setTenantId(profile.tenant_id)
  }
}

export function getProfile(): StoredProfile | null {
  const raw = localStorage.getItem(PROFILE_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as StoredProfile
  } catch {
    return null
  }
}

export function clearProfile() {
  localStorage.removeItem(PROFILE_KEY)
}

export function setScope(scope: 'console' | 'platform') {
  localStorage.setItem(SCOPE_KEY, scope)
}

export function getScope() {
  const raw = localStorage.getItem(SCOPE_KEY)
  if (raw === 'console' || raw === 'platform') {
    return raw
  }
  return null
}

export function clearScope() {
  localStorage.removeItem(SCOPE_KEY)
}

export function getCurrentTenantId() {
  const profile = getProfile()
  if (profile?.tenant_id) return profile.tenant_id
  return getTenantId()
}
