import { Button, Form, Input, Modal, Select, Space, message } from 'antd'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function Roles() {
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [permOpen, setPermOpen] = useState(false)
  const [activeRoleId, setActiveRoleId] = useState('')
  const [permissionCodes, setPermissionCodes] = useState<string[]>([])
  const [permLoading, setPermLoading] = useState(false)
  const [form] = Form.useForm()

  const { data, loading, source, error, reload } = useRequest(() => consoleApi.listRoles(), { items: [] })
  const { data: permissionData } = useRequest(() => consoleApi.listPermissions(), { items: [] })

  const filtered = useMemo(() => {
    if (!keyword) return data.items
    return data.items.filter((item) => item.name.toLowerCase().includes(keyword.toLowerCase()))
  }, [data.items, keyword])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await consoleApi.createRole(values)
      message.success('已创建角色')
      form.resetFields()
      setCreateOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  const openPermissions = async (roleId: string) => {
    setActiveRoleId(roleId)
    setPermOpen(true)
    setPermLoading(true)
    try {
      const res = await consoleApi.listRolePermissions(roleId)
      setPermissionCodes(res.items.map((item) => item.code))
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    } finally {
      setPermLoading(false)
    }
  }

  const handleAssignPermissions = async () => {
    try {
      await consoleApi.assignRolePermissions(activeRoleId, permissionCodes)
      message.success('已更新角色权限')
      setPermOpen(false)
    } catch (err) {
      if (err instanceof Error) message.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader title="角色管理" description="定义与授权租户角色" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索角色" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <Button type="primary" onClick={() => setCreateOpen(true)}>
            新建角色
          </Button>
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
            { title: '名称', dataIndex: 'name' },
            { title: '创建时间', dataIndex: 'created_at' },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Button size="small" onClick={() => openPermissions(record.id)}>
                    编辑权限
                  </Button>
                </Space>
              ),
            },
          ],
        }}
      />

      <Modal
        title="新建角色"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        okText="创建"
      >
        <Form form={form} layout="vertical">
          <Form.Item label="角色名称" name="name" rules={[{ required: true, message: '请输入角色名称' }]}>
            <Input placeholder="例如：Knowledge Manager" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑角色权限"
        open={permOpen}
        onCancel={() => setPermOpen(false)}
        onOk={handleAssignPermissions}
        okText="保存"
        confirmLoading={permLoading}
      >
        <Form layout="vertical">
          <Form.Item label="权限">
            <Select
              mode="multiple"
              placeholder="选择权限"
              value={permissionCodes}
              onChange={setPermissionCodes}
              options={permissionData.items.map((item) => ({
                value: item.code,
                label: item.description ? `${item.code} · ${item.description}` : item.code,
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
