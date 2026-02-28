import { Button, Form, Input, Space, Typography } from 'antd'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { AuthLayout } from '../../layouts/AuthLayout'
import { setProfile, setScope, setTenantId, setToken } from '../../auth/storage'
import { authApi } from '../../services/auth'
import { normalizeAccount, validateAccount } from './utils'

import { uiMessage } from '../../services/uiMessage'
export function ConsoleLogin() {
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const [showTenantField, setShowTenantField] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const onFinish = async (values: {
    account: string
    password: string
    tenant_id?: string
  }) => {
    try {
      setSubmitting(true)
      const account = normalizeAccount(values.account)
      if (!account) {
        uiMessage.error('请输入邮箱或手机号')
        return
      }
      const res = await authApi.consoleLogin({
        account,
        password: values.password,
        tenant_id: values.tenant_id?.trim() || undefined,
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
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout title="Console 登录" subtitle="租户管理员登录控制台">
      <Form
        form={form}
        layout="vertical"
        requiredMark
        style={{ marginTop: 24 }}
        onFinish={onFinish}
      >
        <Form.Item
          label="账号（邮箱/手机号）"
          name="account"
          rules={[
            { required: true, message: '请输入邮箱或手机号' },
            {
              validator: (_, value: string) => {
                if (!value || validateAccount(value)) return Promise.resolve()
                return Promise.reject(new Error('请输入合法邮箱或手机号'))
              },
            },
          ]}
          extra="支持邮箱或手机号登录"
        >
          <Input placeholder="请输入邮箱或手机号" allowClear autoComplete="username" />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
          <Input.Password placeholder="请输入密码" autoComplete="current-password" />
        </Form.Item>
        {showTenantField ? (
          <Form.Item
            label="租户 ID（可选）"
            name="tenant_id"
            extra="仅当同一账号加入多个租户时才需要填写"
          >
            <Input placeholder="例如：xxxx-xxxx-xxxx" allowClear />
          </Form.Item>
        ) : null}
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" htmlType="submit" block loading={submitting}>
            登录并进入控制台
          </Button>
          <Button type="link" onClick={() => setShowTenantField((v) => !v)} style={{ paddingInline: 0 }}>
            {showTenantField ? '收起租户 ID（高级）' : '切换租户登录（高级）'}
          </Button>
          <Typography.Text className="muted">使用账号密码登录控制台。</Typography.Text>
          <Typography.Text className="muted">
            还没有账号？<Link to="/console/register">创建租户</Link>
          </Typography.Text>
        </Space>
      </Form>
    </AuthLayout>
  )
}

