import { Button, Form, Input, Space, Typography, message } from 'antd'
import { AuthLayout } from '../../layouts/AuthLayout'
import { setToken } from '../../auth/storage'

export function PlatformLogin() {
  const [form] = Form.useForm()

  const onFinish = (values: { email: string; password: string; token?: string }) => {
    if (values.token) {
      setToken(values.token)
      message.success('Token 已保存')
      return
    }
    message.info('登录接口尚未接入，请配置 Token 进行联调')
  }

  return (
    <AuthLayout title="Platform 登录" subtitle="平台管理员登录平台管理区">
      <Form form={form} layout="vertical" style={{ marginTop: 24 }} onFinish={onFinish}>
        <Form.Item label="邮箱" name="email" rules={[{ required: true }]}>
          <Input placeholder="admin@ragodesk.ai" />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true }]}>
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
        <Form.Item label="Access Token（可选）" name="token">
          <Input placeholder="粘贴 JWT 用于调试" />
        </Form.Item>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" htmlType="submit" block>
            登录
          </Button>
          <Typography.Text className="muted">
            登录接口尚未接入，后续将对接 JWT。
          </Typography.Text>
        </Space>
      </Form>
    </AuthLayout>
  )
}
