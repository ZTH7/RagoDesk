import { Button, Form, Input, Select, Space, Typography } from 'antd'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { AuthLayout } from '../../layouts/AuthLayout'
import { setProfile, setScope, setTenantId, setToken } from '../../auth/storage'
import { authApi } from '../../services/auth'
import {
  looksLikeEmail,
  normalizeAccount,
  normalizePhone,
  validateAccount,
} from './utils'

import { uiMessage } from '../../services/uiMessage'
export function ConsoleRegister() {
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const [submitting, setSubmitting] = useState(false)

  const onFinish = async (values: {
    tenant_name: string
    tenant_type?: string
    admin_name?: string
    account: string
    password: string
    confirm_password: string
  }) => {
    try {
      setSubmitting(true)
      const account = normalizeAccount(values.account)
      const tenantName = values.tenant_name.trim()
      const adminName = values.admin_name?.trim()
      const phone = normalizePhone(account)
      const payload = {
        tenant_name: tenantName,
        tenant_type: values.tenant_type,
        admin_name: adminName || undefined,
        email: looksLikeEmail(account) ? account : undefined,
        phone: looksLikeEmail(account) ? undefined : phone,
        password: values.password,
      }
      const res = await authApi.consoleRegister(payload)
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
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout title="Console 注册" subtitle="创建租户并初始化管理员账号">
      <Form
        form={form}
        layout="vertical"
        requiredMark
        style={{ marginTop: 24 }}
        onFinish={onFinish}
      >
        <Form.Item label="租户名称" name="tenant_name" rules={[{ required: true, message: '请输入租户名称' }]}>
          <Input placeholder="例如：Acme Inc" allowClear />
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
          label="管理员账号（邮箱/手机号）"
          name="account"
          rules={[
            { required: true, message: '请输入邮箱或手机号' },
            {
              validator: (_, value: string) => {
                if (!value) return Promise.resolve()
                if (validateAccount(value)) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error('请输入合法邮箱或手机号'))
              },
            },
          ]}
          extra="支持邮箱或手机号，手机号中的空格与 - 会自动忽略"
        >
          <Input placeholder="alice@acme.com 或 +86 13800000000" allowClear autoComplete="username" />
        </Form.Item>
        <Form.Item
          label="密码"
          name="password"
          rules={[
            { required: true, message: '请输入密码' },
            { min: 6, message: '密码至少 6 位' },
          ]}
        >
          <Input.Password placeholder="至少 6 位" autoComplete="new-password" />
        </Form.Item>
        <Form.Item
          label="确认密码"
          name="confirm_password"
          dependencies={['password']}
          rules={[
            { required: true, message: '请再次输入密码' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve()
                }
                return Promise.reject(new Error('两次输入的密码不一致'))
              },
            }),
          ]}
        >
          <Input.Password placeholder="请再次输入密码" autoComplete="new-password" />
        </Form.Item>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button type="primary" htmlType="submit" block loading={submitting}>
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

