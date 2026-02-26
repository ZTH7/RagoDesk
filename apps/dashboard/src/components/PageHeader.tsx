import type { ReactNode } from 'react'
import { Typography } from 'antd'

type PageHeaderProps = {
  title: string
  description?: string
  extra?: ReactNode
}

export function PageHeader({ title, description, extra }: PageHeaderProps) {
  return (
    <div className="page-header motion-enter">
      <div>
        <Typography.Title level={3} style={{ marginBottom: 4 }}>
          {title}
        </Typography.Title>
        {description ? (
          <Typography.Text type="secondary">{description}</Typography.Text>
        ) : null}
      </div>
      {extra ? <div className="page-actions">{extra}</div> : null}
    </div>
  )
}
