import { Button, Form, Input, Modal, Select, Space, Tag, message } from 'antd'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'

const statusColors: Record<string, string> = {
  active: 'green',
  disabled: 'red',
}

export function PlatformAdmins() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [assignOpen, setAssignOpen] = useState(false)
  const [activeAdminId, setActiveAdminId] = useState('')
  const [createForm] = Form.useForm()
  const [assignForm] = Form.useForm()

  const { data, loading, source, error, reload } = useRequest(() => platformApi.listAdmins(), { items: [] })
  const { data: roleData } = useRequest(() => platformApi.listRoles(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && item.status !== status) return false
      if (keyword && !item.name.toLowerCase().includes(keyword.toLowerCase())) return false
      return true
    })
  }, [data.items, keyword, status])

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields()
      await platformApi.createAdmin(values)
      message.success('已创建管理员')
      setCreateOpen(false)
      createForm.resetFields()
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  const handleAssign = async () => {
    try {
      const values = await assignForm.validateFields()
      await platformApi.assignAdminRole(activeAdminId, values.role_id)
      message.success('已分配角色')
      setAssignOpen(false)
      assignForm.resetFields()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
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
        left={<Input.Search placeholder="搜索管理员" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <>
            <Select
              value={status}
              style={{ width: 160 }}
              onChange={setStatus}
              options={[
                { value: 'all', label: '全部状态' },
                { value: 'active', label: 'Active' },
                { value: 'disabled', label: 'Disabled' },
              ]}
            />
            <Button type="primary" onClick={() => setCreateOpen(true)}>
              新建管理员
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
            {
              title: 'ID',
              dataIndex: 'id',
              render: (value: string) => <Link to={`/platform/admins/${value}`}>{value}</Link>,
            },
            { title: '姓名', dataIndex: 'name' },
            { title: '邮箱', dataIndex: 'email' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={statusColors[value] || 'default'}>{value}</Tag>,
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
        <Form form={createForm} layout="vertical" initialValues={{ status: 'active' }}>
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
            <Select options={[{ value: 'active', label: 'Active' }, { value: 'disabled', label: 'Disabled' }]} />
          </Form.Item>
          <Form.Item
            label="Password Hash"
            name="password_hash"
            rules={[{ required: true, message: '请输入密码哈希' }]}
          >
            <Input.Password placeholder="后端要求 password_hash" />
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
