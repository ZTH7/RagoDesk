import { Button, Form, Input, Modal, Select, Space, Tag, message } from 'antd'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'
import { getTenantId } from '../../auth/storage'

export function Users() {
  const tenantId = getTenantId() ?? ''
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [inviteOpen, setInviteOpen] = useState(false)
  const [assignOpen, setAssignOpen] = useState(false)
  const [assigningUserId, setAssigningUserId] = useState('')
  const [inviteForm] = Form.useForm()
  const [assignForm] = Form.useForm()

  const { data, loading, source, error, reload } = useRequest(
    () => consoleApi.listUsers(tenantId),
    { items: [] },
    { enabled: Boolean(tenantId), deps: [tenantId] },
  )
  const { data: roleData } = useRequest(() => consoleApi.listRoles(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && item.status !== status) return false
      if (keyword && !item.name.toLowerCase().includes(keyword.toLowerCase())) return false
      return true
    })
  }, [data.items, keyword, status])

  const handleInvite = async () => {
    if (!tenantId) {
      message.error('请先在个人中心设置 Tenant ID')
      return
    }
    try {
      const values = await inviteForm.validateFields()
      await consoleApi.createUser(tenantId, values)
      message.success('已邀请成员')
      inviteForm.resetFields()
      setInviteOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  const handleAssign = async () => {
    try {
      const values = await assignForm.validateFields()
      await consoleApi.assignRole(assigningUserId, values.role_id)
      message.success('已分配角色')
      setAssignOpen(false)
      assignForm.resetFields()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader title="成员管理" description="邀请与管理租户成员" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索成员" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <>
            <Select
              value={status}
              style={{ width: 160 }}
              onChange={setStatus}
              options={[
                { value: 'all', label: '全部状态' },
                { value: 'active', label: 'Active' },
                { value: 'invited', label: 'Invited' },
              ]}
            />
            <Button type="primary" onClick={() => setInviteOpen(true)}>
              邀请成员
            </Button>
          </>
        }
      />
      <TableCard
        table={{
          rowKey: 'id',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 8 },
          columns: [
            { title: 'ID', dataIndex: 'id' },
            { title: '姓名', dataIndex: 'name' },
            { title: '邮箱', dataIndex: 'email' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => (
                <Tag color={value === 'active' ? 'green' : 'orange'}>{value}</Tag>
              ),
            },
            { title: '创建时间', dataIndex: 'created_at' },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Button
                    size="small"
                    onClick={() => {
                      setAssigningUserId(record.id)
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
        title="邀请成员"
        open={inviteOpen}
        onCancel={() => setInviteOpen(false)}
        onOk={handleInvite}
        okText="发送邀请"
      >
        <Form form={inviteForm} layout="vertical" initialValues={{ status: 'active' }}>
          <Form.Item label="姓名" name="name" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="成员姓名" />
          </Form.Item>
          <Form.Item label="邮箱" name="email" rules={[{ required: true, message: '请输入邮箱' }]}>
            <Input placeholder="name@company.com" />
          </Form.Item>
          <Form.Item label="手机号" name="phone">
            <Input placeholder="可选" />
          </Form.Item>
          <Form.Item label="状态" name="status">
            <Select
              options={[
                { value: 'active', label: 'Active' },
                { value: 'invited', label: 'Invited' },
              ]}
            />
          </Form.Item>
        </Form>
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
        </Form>
      </Modal>
    </div>
  )
}
