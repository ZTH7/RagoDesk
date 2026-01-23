import { Button, Card, Descriptions, Form, Input, Space, Typography, message } from 'antd'
import { useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { clearTenantId, clearToken, getTenantId, getToken, setTenantId, setToken } from '../../auth/storage'

export function Profile() {
  const [form] = Form.useForm()
  const [token, setTokenState] = useState(getToken() ?? '')
  const [tenantId, setTenantIdState] = useState(getTenantId() ?? '')

  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      const nextToken = (values.token as string).trim()
      const nextTenantId = (values.tenantId as string).trim()

      if (nextToken) {
        setToken(nextToken)
        setTokenState(nextToken)
      } else {
        clearToken()
        setTokenState('')
      }

      if (nextTenantId) {
        setTenantId(nextTenantId)
        setTenantIdState(nextTenantId)
      } else {
        clearTenantId()
        setTenantIdState('')
      }

      message.success('已更新本地会话信息')
    } catch (err) {
      if (err instanceof Error) {
        message.error(err.message)
      }
    }
  }

  return (
    <div className="page">
      <PageHeader title="个人中心" description="会话信息与本地设置" />
      <Card>
        <Descriptions column={1} bordered size="middle">
          <Descriptions.Item label="Token 状态">{token ? '已设置' : '未设置'}</Descriptions.Item>
          <Descriptions.Item label="Tenant ID">{tenantId || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>
      <Card>
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            token,
            tenantId,
          }}
        >
          <Form.Item label="Tenant ID" name="tenantId">
            <Input placeholder="用于 Console 成员管理等路径参数" />
          </Form.Item>
          <Form.Item label="Bearer Token" name="token">
            <Input.Password placeholder="可在登录页或此处手动设置" />
          </Form.Item>
          <Typography.Paragraph type="secondary" style={{ marginBottom: 16 }}>
            没有登录接口时，可在此设置本地 Token 与 Tenant ID 以便调用 API。
          </Typography.Paragraph>
          <Space>
            <Button type="primary" onClick={handleSave}>
              保存设置
            </Button>
            <Button
              onClick={() => {
                clearToken()
                clearTenantId()
                setTokenState('')
                setTenantIdState('')
                form.setFieldsValue({ token: '', tenantId: '' })
                message.success('已清空本地会话')
              }}
            >
              清空会话
            </Button>
          </Space>
        </Form>
      </Card>
    </div>
  )
}
