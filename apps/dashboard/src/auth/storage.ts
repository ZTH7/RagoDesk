const TOKEN_KEY = 'ragodesk_token'
const TENANT_KEY = 'ragodesk_tenant_id'

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
