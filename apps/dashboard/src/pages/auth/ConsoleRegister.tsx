import { Button, Form, Input, Select, Space, Typography } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { AuthLayout } from '../../layouts/AuthLayout'
import { setProfile, setScope, setTenantId, setToken } from '../../auth/storage'
import { authApi } from '../../services/auth'

import { uiMessage } from '../../services/uiMessage'
export function ConsoleRegister() {
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const onFinish = async (values: {
    tenant_name: string
    tenant_type?: string
    admin_name?: string
    email?: string
    phone?: string
    password: string
  }) => {
    try {
      const res = await authApi.consoleRegister(values)
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
      uiMessage.success('注册成功')
      navigate('/console/analytics/overview', { replace: true })
    } catch (err) {
      if (err instanceof Error) {
        uiMessage.error(err.message)
      }
    }
  }

  return (
    <AuthLayout title="Console 注册" subtitle="创建租户并初始化管理员账号">
      <Form form={form} layout="vertical" style={{ marginTop: 24 }} onFinish={onFinish}>
        <Form.Item label="租户名称" name="tenant_name" rules={[{ required: true }]}>
          <Input placeholder="Acme Inc" />
        </Form.Item>
        <Form.Item label="租户类型" name="tenant_type" initialValue="enterprise">
          <Select
            options={[
              { label: '企业', value: 'enterprise' },
              { label: '个人', value: 'personal' },
            ]}
          />
        </Form.Item>
        <Form.Item label="管理员姓名" name="admin_name">
          <Input placeholder="Alice" />
        </Form.Item>
        <Form.Item
          label="管理员邮箱"
          name="email"
          rules={[
            { type: 'email', message: '邮箱格式不正确' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (value || getFieldValue('phone')) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error('邮箱或手机号至少填写一项'))
              },
            }),
          ]}
        >
          <Input placeholder="alice@acme.com" />
        </Form.Item>
        <Form.Item label="管理员手机号" name="phone">
          <Input placeholder="+86 13800000000" />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true, min: 6 }]}> 
          <Input.Password placeholder="至少 6 位" />
        </Form.Item>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" htmlType="submit" block>
            注册并进入控制台
          </Button>
          <Typography.Text className="muted">注册会自动创建 tenant_admin 角色并授予租户权限。</Typography.Text>
          <Typography.Text className="muted">
            已有账号？<Link to="/console/login">返回登录</Link>
          </Typography.Text>
        </Space>
      </Form>
    </AuthLayout>
  )
}

