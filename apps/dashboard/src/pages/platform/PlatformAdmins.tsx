import { Button, Descriptions, Form, Input, Modal, Radio, Select, Space, Switch, Tag, Typography } from 'antd'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'
import { formatDateTime } from '../../utils/datetime'

import { uiMessage } from '../../services/uiMessage'
const statusColors: Record<string, string> = {
  active: 'green',
  disabled: 'red',
}

const statusLabels: Record<string, string> = {
  active: '启用',
  disabled: '停用',
}

export function PlatformAdmins() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [assignOpen, setAssignOpen] = useState(false)
  const [activeAdminId, setActiveAdminId] = useState('')
  const [createMode, setCreateMode] = useState<'password' | 'invite'>('password')
  const [inviteLink, setInviteLink] = useState('')
  const [inviteOpen, setInviteOpen] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [createForm] = Form.useForm()
  const [assignForm] = Form.useForm()

  const { data, loading, source, error, reload } = useRequest(() => platformApi.listAdmins(), { items: [] })
  const { data: roleData } = useRequest(() => platformApi.listRoles(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && item.status !== status) return false
      if (
        keyword &&
        !item.name.toLowerCase().includes(keyword.toLowerCase()) &&
        !item.email.toLowerCase().includes(keyword.toLowerCase())
      ) {
        return false
      }
      return true
    })
  }, [data.items, keyword, status])

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields()
      const payload: any = {
        name: values.name,
        email: values.email,
        phone: values.phone,
        status: values.status,
      }
      if (createMode === 'password') {
        payload.password = values.password
      } else {
        payload.send_invite = true
        payload.invite_base_url = window.location.origin
      }
      const res = await platformApi.createAdmin(payload)
      uiMessage.success('已创建管理员')
      setCreateOpen(false)
      createForm.resetFields()
      if (res.invite_link) {
        setInviteLink(res.invite_link)
        setInviteOpen(true)
      }
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleAssign = async () => {
    try {
      const values = await assignForm.validateFields()
      await platformApi.assignAdminRole(activeAdminId, values.role_id)
      uiMessage.success('已分配角色')
      setAssignOpen(false)
      assignForm.resetFields()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="平台管理员"
        description="平台管理员列表与角色"
        extra={<DataSourceTag source={source} />}
      />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="按姓名或邮箱搜索" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <>
            <Select
              value={status}
              style={{ width: 160 }}
              onChange={setStatus}
              options={[
                { value: 'all', label: '全部状态' },
                { value: 'active', label: '启用' },
                { value: 'disabled', label: '停用' },
              ]}
            />
            <Button type="primary" onClick={() => setCreateOpen(true)}>
              新建管理员
            </Button>
            <Space size={6}>
              <Typography.Text className="muted">高级列</Typography.Text>
              <Switch checked={showAdvanced} onChange={setShowAdvanced} />
            </Space>
          </>
        }
      />
      <TableCard
        table={{
          rowKey: 'id',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 8 },
          expandable: showAdvanced
            ? {
                expandedRowRender: (record) => (
                  <Descriptions column={2} bordered size="small">
                    <Descriptions.Item label="Admin ID">{record.id}</Descriptions.Item>
                    <Descriptions.Item label="创建时间">{formatDateTime(record.created_at)}</Descriptions.Item>
                  </Descriptions>
                ),
              }
            : undefined,
          columns: [
            {
              title: '姓名',
              dataIndex: 'name',
              render: (_: string, record) => <Link to={`/platform/admins/${record.id}`}>{record.name}</Link>,
            },
            { title: '邮箱', dataIndex: 'email' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={statusColors[value] || 'default'}>{statusLabels[value] || value}</Tag>,
            },
            { title: '创建时间', dataIndex: 'created_at', render: (value: string) => formatDateTime(value) },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Button
                    size="small"
                    onClick={() => {
                      setActiveAdminId(record.id)
                      setAssignOpen(true)
                    }}
                  >
                    分配角色
                  </Button>
                </Space>
              ),
            },
          ],
        }}
      />

      <Modal
        title="新建平台管理员"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        okText="创建"
      >
        <Form
          form={createForm}
          layout="vertical"
          initialValues={{ status: 'active' }}
          onValuesChange={(changed) => {
            if (changed.createMode) {
              setCreateMode(changed.createMode)
            }
          }}
        >
          <Form.Item label="姓名" name="name" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="管理员姓名" />
          </Form.Item>
          <Form.Item label="邮箱" name="email" rules={[{ required: true, message: '请输入邮箱' }]}>
            <Input placeholder="admin@company.com" />
          </Form.Item>
          <Form.Item label="手机号" name="phone">
            <Input placeholder="可选" />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={[{ value: 'active', label: '启用' }, { value: 'disabled', label: '停用' }]} />
          </Form.Item>
          <Form.Item label="创建方式" name="createMode" initialValue="password">
            <Radio.Group>
              <Radio value="password">设置初始密码</Radio>
              <Radio value="invite">发送邀请链接</Radio>
            </Radio.Group>
          </Form.Item>
          {createMode === 'password' ? (
            <Form.Item label="初始密码" name="password" rules={[{ required: true, message: '请输入初始密码' }]}>
              <Input.Password placeholder="设置初始密码" />
            </Form.Item>
          ) : (
            <Typography.Text className="muted">
              系统会生成临时密码并返回邀请链接，请复制后发送给管理员。
            </Typography.Text>
          )}
        </Form>
      </Modal>

      <Modal
        title="邀请链接"
        open={inviteOpen}
        onCancel={() => setInviteOpen(false)}
        footer={<Button onClick={() => setInviteOpen(false)}>关闭</Button>}
      >
        <Typography.Paragraph className="muted">
          复制以下链接发送给管理员完成登录。
        </Typography.Paragraph>
        <Input.TextArea value={inviteLink} readOnly rows={3} />
      </Modal>

      <Modal
        title="分配角色"
        open={assignOpen}
        onCancel={() => setAssignOpen(false)}
        onOk={handleAssign}
        okText="保存"
      >
        <Form form={assignForm} layout="vertical">
          <Form.Item label="角色" name="role_id" rules={[{ required: true, message: '请选择角色' }]}>
            <Select
              placeholder="选择角色"
              options={roleData.items.map((role) => ({
                value: role.id,
                label: role.name,
              }))}
            />
          </Form.Item>
          {roleData.items.length === 0 ? (
            <Typography.Text className="muted">暂无角色，请先前往「平台角色」创建角色。</Typography.Text>
          ) : null}
        </Form>
      </Modal>
    </div>
  )
}

