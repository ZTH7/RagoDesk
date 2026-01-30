import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { getScope, getToken } from '../auth/storage'

type RequireAuthProps = {
  scope: 'console' | 'platform'
  children: ReactNode
}

export function RequireAuth({ scope, children }: RequireAuthProps) {
  const location = useLocation()
  const token = getToken()
  const storedScope = getScope()

  if (!token) {
    return <Navigate to={`/${scope}/login`} replace state={{ from: location.pathname }} />
  }

  if (storedScope && storedScope !== scope) {
    return <Navigate to={`/${storedScope}/login`} replace state={{ from: location.pathname }} />
  }

  return <>{children}</>
}
