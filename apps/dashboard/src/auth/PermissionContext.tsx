import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { consoleApi } from '../services/console'
import { platformApi } from '../services/platform'

export type PermissionScope = 'console' | 'platform'

type PermissionState = {
  scope: PermissionScope
  permissions: Set<string>
  loading: boolean
  error: string | null
  stale: boolean
  refresh: () => void
}

const PermissionContext = createContext<PermissionState | null>(null)

function loadCachedPermissions(scope: PermissionScope): Set<string> {
  if (typeof window === 'undefined') return new Set<string>()
  try {
    const raw = window.localStorage.getItem(`ragodesk.permissions.${scope}`)
    if (!raw) return new Set<string>()
    const parsed = JSON.parse(raw) as string[]
    if (!Array.isArray(parsed)) return new Set<string>()
    return new Set(parsed)
  } catch {
    return new Set<string>()
  }
}

function saveCachedPermissions(scope: PermissionScope, permissions: Set<string>) {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(
    `ragodesk.permissions.${scope}`,
    JSON.stringify(Array.from(permissions)),
  )
}

export function PermissionProvider({
  children,
  scope,
}: {
  children: ReactNode
  scope: PermissionScope
}) {
  const [permissions, setPermissions] = useState<Set<string>>(() => loadCachedPermissions(scope))
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [stale, setStale] = useState(false)
  const [refreshIndex, setRefreshIndex] = useState(0)
  const refresh = useCallback(() => {
    setRefreshIndex((v) => v + 1)
  }, [])

  useEffect(() => {
    let active = true
    const cached = loadCachedPermissions(scope)
    setLoading(true)
    setError(null)
    setStale(false)
    setPermissions(cached)

    const loader =
      scope === 'platform' ? platformApi.listPermissions : consoleApi.listPermissions

    loader()
      .then((res) => {
        if (!active) return
        const next = new Set(res.items.map((item) => item.code))
        setPermissions(next)
        saveCachedPermissions(scope, next)
        setStale(false)
      })
      .catch((err: Error) => {
        if (!active) return
        setError(err.message)
        setStale(cached.size > 0)
      })
      .finally(() => {
        if (!active) return
        setLoading(false)
      })

    return () => {
      active = false
    }
  }, [scope, refreshIndex])

  const value = useMemo(
    () => ({
      scope,
      permissions,
      loading,
      error,
      stale,
      refresh,
    }),
    [scope, permissions, loading, error, stale, refresh],
  )

  return <PermissionContext.Provider value={value}>{children}</PermissionContext.Provider>
}

export function usePermissions() {
  const ctx = useContext(PermissionContext)
  if (!ctx) {
    return {
      scope: 'console' as PermissionScope,
      permissions: new Set<string>(),
      loading: false,
      error: null,
      stale: false,
      refresh: () => {},
    }
  }
  return ctx
}
