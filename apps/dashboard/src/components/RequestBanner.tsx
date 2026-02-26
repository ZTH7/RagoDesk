import { Alert } from 'antd'

type RequestBannerProps = {
  error?: string | null
}

export function RequestBanner({ error }: RequestBannerProps) {
  if (!error) return null

  return (
    <Alert
      type="warning"
      title="接口暂不可用"
      description={error}
      showIcon
    />
  )
}
