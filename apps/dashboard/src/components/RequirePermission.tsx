import type { ReactNode } from 'react'
import { Skeleton } from 'antd'
import { usePermissions } from '../auth/PermissionContext'
import { Forbidden } from '../pages/Forbidden'

type RequirePermissionProps = {
  permission?: string
  children: ReactNode
}

export function RequirePermission({ permission, children }: RequirePermissionProps) {
  const { permissions, loading } = usePermissions()

  if (loading) {
    return <Skeleton active paragraph={{ rows: 3 }} />
  }

  if (permission && !permissions.has(permission)) {
    return <Forbidden />
  }

  return <>{children}</>
}
