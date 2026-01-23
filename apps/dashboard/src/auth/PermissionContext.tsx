import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { defaultConsolePermissions, defaultPlatformPermissions } from './permissions'
import { consoleApi } from '../services/console'
import { platformApi } from '../services/platform'

export type PermissionScope = 'console' | 'platform'

type PermissionState = {
  scope: PermissionScope
  permissions: Set<string>
  loading: boolean
  error: string | null
}

const PermissionContext = createContext<PermissionState | null>(null)

export function PermissionProvider({
  children,
  scope,
}: {
  children: ReactNode
  scope: PermissionScope
}) {
  const [permissions, setPermissions] = useState<Set<string>>(
    scope === 'platform' ? defaultPlatformPermissions : defaultConsolePermissions,
  )
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let active = true
    setLoading(true)
    setError(null)

    const loader = scope === 'platform' ? platformApi.listPermissions : consoleApi.listPermissions

    loader()
      .then((res) => {
        if (!active) return
        const codes = res.items.map((item) => item.code)
        setPermissions(new Set(codes))
      })
      .catch((err: Error) => {
        if (!active) return
        setError(err.message)
      })
      .finally(() => {
        if (!active) return
        setLoading(false)
      })

    return () => {
      active = false
    }
  }, [scope])

  const value = useMemo(
    () => ({
      scope,
      permissions,
      loading,
      error,
    }),
    [scope, permissions, loading, error],
  )

  return <PermissionContext.Provider value={value}>{children}</PermissionContext.Provider>
}

export function usePermissions() {
  const ctx = useContext(PermissionContext)
  if (!ctx) {
    return {
      scope: 'console' as PermissionScope,
      permissions: defaultConsolePermissions,
      loading: false,
      error: null,
    }
  }
  return ctx
}
