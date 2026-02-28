import { Button, Form, Input, Space, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'
import { AuthLayout } from '../../layouts/AuthLayout'
import { setProfile, setScope, setToken } from '../../auth/storage'
import { authApi } from '../../services/auth'

import { uiMessage } from '../../services/uiMessage'
export function PlatformLogin() {
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const onFinish = async (values: { account: string; password: string }) => {
    try {
      const res = await authApi.platformLogin({
        account: values.account,
        password: values.password,
      })
      setToken(res.token)
      setScope('platform')
      setProfile({
        subject_id: res.profile?.subject_id,
        name: res.profile?.name,
        account: res.profile?.account,
        roles: res.profile?.roles,
        scope: 'platform',
      })
      uiMessage.success('登录成功')
      navigate('/platform/tenants', { replace: true })
    } catch (err) {
      if (err instanceof Error) {
        uiMessage.error(err.message)
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
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" htmlType="submit" block>
            登录
          </Button>
          <Typography.Text className="muted">使用平台管理员账号登录平台管理区。</Typography.Text>
        </Space>
      </Form>
    </AuthLayout>
  )
}

