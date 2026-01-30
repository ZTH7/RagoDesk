import { Button, Form, Input, Space, Typography, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { AuthLayout } from '../../layouts/AuthLayout'
import { setProfile, setScope, setToken } from '../../auth/storage'
import { authApi } from '../../services/auth'

export function PlatformLogin() {
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const onFinish = async (values: { account: string; password: string; token?: string }) => {
    try {
      if (values.token) {
        setToken(values.token)
        setScope('platform')
        setProfile({ scope: 'platform' })
        message.success('Token 已保存')
        navigate('/platform/tenants', { replace: true })
        return
      }
      const res = await authApi.platformLogin({
        account: values.account,
        password: values.password,
      })
      setToken(res.token)
      setScope('platform')
      setProfile({
        name: res.profile?.name,
        account: res.profile?.account,
        roles: res.profile?.roles,
        scope: 'platform',
      })
      message.success('登录成功')
      navigate('/platform/tenants', { replace: true })
    } catch (err) {
      if (err instanceof Error) {
        message.error(err.message)
      }
    }
  }

  return (
    <AuthLayout title="Platform 登录" subtitle="平台管理员登录平台管理区">
      <Form form={form} layout="vertical" style={{ marginTop: 24 }} onFinish={onFinish}>
        <Form.Item label="账号（邮箱/手机号）" name="account" rules={[{ required: true }]}>
          <Input placeholder="admin@ragodesk.ai" />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true }]}>
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
        <Form.Item label="Access Token（可选）" name="token">
          <Input placeholder="粘贴 JWT 用于调试联调" />
        </Form.Item>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" htmlType="submit" block>
            登录
          </Button>
          <Typography.Text className="muted">使用账号密码登录，或粘贴 JWT 直接进入平台。</Typography.Text>
        </Space>
      </Form>
    </AuthLayout>
  )
}
