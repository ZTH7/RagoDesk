import { Button, Form, Input, Space, Typography, message } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { AuthLayout } from '../../layouts/AuthLayout'
import { clearTenantId, setProfile, setScope, setTenantId, setToken } from '../../auth/storage'
import { authApi } from '../../services/auth'

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
        setProfile({ scope: 'console' })
        if (values.tenant_id) {
          setTenantId(values.tenant_id)
        } else {
          clearTenantId()
        }
        message.success('Token 已保存')
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
        name: res.profile?.name,
        account: res.profile?.account,
        roles: res.profile?.roles,
        scope: 'console',
      })
      if (res.profile?.tenant_id) {
        setTenantId(res.profile.tenant_id)
      }
      message.success('登录成功')
      navigate('/console/analytics/overview', { replace: true })
    } catch (err) {
      if (err instanceof Error) {
        message.error(err.message)
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
        <Form.Item label="Tenant ID（可选）" name="tenant_id">
          <Input placeholder="当账号跨租户时需指定" />
        </Form.Item>
        <Form.Item label="Access Token（可选）" name="token">
          <Input placeholder="粘贴 JWT 用于调试联调" />
        </Form.Item>
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
