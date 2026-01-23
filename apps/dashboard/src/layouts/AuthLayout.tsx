import { Layout, Card, Typography } from 'antd'
import type { ReactNode } from 'react'

const { Content } = Layout

type AuthLayoutProps = {
  title: string
  subtitle: string
  children: ReactNode
}

export function AuthLayout({ title, subtitle, children }: AuthLayoutProps) {
  return (
    <Layout className="auth-shell">
      <Content className="auth-content">
        <Card className="auth-card" bordered={false}>
          <div className="auth-brand">RagoDesk</div>
          <Typography.Title level={3} style={{ marginBottom: 4 }}>
            {title}
          </Typography.Title>
          <Typography.Text className="muted">{subtitle}</Typography.Text>
          <div className="auth-body">{children}</div>
        </Card>
      </Content>
    </Layout>
  )
}
