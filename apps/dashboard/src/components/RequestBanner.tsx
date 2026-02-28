import type { ReactNode } from 'react'
import { Alert } from 'antd'

type RequestBannerProps = {
  error?: string | null
  title?: string
  description?: ReactNode
  action?: ReactNode
}

export function RequestBanner({ error, title, description, action }: RequestBannerProps) {
  if (!error) return null

  return (
    <Alert
      type="warning"
      title={title || '接口暂不可用'}
      description={description || error}
      action={action}
      showIcon
    />
  )
}
