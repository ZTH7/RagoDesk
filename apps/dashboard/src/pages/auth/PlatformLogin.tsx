import { Button, Form, Input, Space, Typography } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AuthLayout } from '../../layouts/AuthLayout'
import { setProfile, setScope, setToken } from '../../auth/storage'
import { authApi } from '../../services/auth'
import { normalizeAccount, validateAccount } from './utils'

import { uiMessage } from '../../services/uiMessage'
export function PlatformLogin() {
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const [submitting, setSubmitting] = useState(false)

  const onFinish = async (values: { account: string; password: string }) => {
    try {
      setSubmitting(true)
      const account = normalizeAccount(values.account)
      if (!account) {
        uiMessage.error('请输入邮箱或手机号')
        return
      }
      const res = await authApi.platformLogin({
        account,
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
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout title="Platform 登录" subtitle="平台管理员登录平台管理区">
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
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" htmlType="submit" block loading={submitting}>
            登录并进入平台管理区
          </Button>
          <Typography.Text className="muted">使用平台管理员账号登录平台管理区。</Typography.Text>
        </Space>
      </Form>
    </AuthLayout>
  )
}

