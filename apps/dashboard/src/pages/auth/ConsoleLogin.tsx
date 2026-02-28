import { Button, Collapse, Form, Input, Space, Typography } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { AuthLayout } from '../../layouts/AuthLayout'
import { clearTenantId, setProfile, setScope, setTenantId, setToken } from '../../auth/storage'
import { authApi } from '../../services/auth'

import { uiMessage } from '../../services/uiMessage'
export function ConsoleLogin() {
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const onFinish = async (values: {
    account: string
    password: string
    tenant_id?: string
    token?: string
  }) => {
    try {
      if (values.token) {
        setToken(values.token)
        setScope('console')
        setProfile({
          scope: 'console',
          tenant_id: values.tenant_id || undefined,
        })
        if (values.tenant_id) {
          setTenantId(values.tenant_id)
        } else {
          clearTenantId()
        }
        uiMessage.success('Token 已保存')
        navigate('/console/analytics/overview', { replace: true })
        return
      }
      const res = await authApi.consoleLogin({
        account: values.account,
        password: values.password,
        tenant_id: values.tenant_id || undefined,
      })
      setToken(res.token)
      setScope('console')
      setProfile({
        subject_id: res.profile?.subject_id,
        tenant_id: res.profile?.tenant_id,
        name: res.profile?.name,
        account: res.profile?.account,
        roles: res.profile?.roles,
        scope: 'console',
      })
      if (res.profile?.tenant_id) {
        setTenantId(res.profile.tenant_id)
      }
      uiMessage.success('登录成功')
      navigate('/console/analytics/overview', { replace: true })
    } catch (err) {
      if (err instanceof Error) {
        uiMessage.error(err.message)
      }
    }
  }

  return (
    <AuthLayout title="Console 登录" subtitle="租户管理员登录控制台">
      <Form form={form} layout="vertical" style={{ marginTop: 24 }} onFinish={onFinish}>
        <Form.Item label="账号（邮箱/手机号）" name="account" rules={[{ required: true }]}>
          <Input placeholder="you@company.com" />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true }]}>
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
        <Collapse
          size="small"
          items={[
            {
              key: 'advanced',
              label: '高级登录选项（一般无需填写）',
              children: (
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Form.Item label="Tenant ID（可选）" name="tenant_id" style={{ marginBottom: 0 }}>
                    <Input placeholder="仅账号跨租户时填写" />
                  </Form.Item>
                  <Form.Item label="Access Token（可选）" name="token" style={{ marginBottom: 0 }}>
                    <Input placeholder="仅联调时粘贴 JWT" />
                  </Form.Item>
                </Space>
              ),
            },
          ]}
        />
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" htmlType="submit" block>
            登录
          </Button>
          <Typography.Text className="muted">使用账号密码登录，或粘贴 JWT 直接进入控制台。</Typography.Text>
          <Typography.Text className="muted">
            还没有账号？<Link to="/console/register">创建租户</Link>
          </Typography.Text>
        </Space>
      </Form>
    </AuthLayout>
  )
}

