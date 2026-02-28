import { Button, Descriptions, Form, Input, Modal, Select, Space, Switch, Typography } from 'antd'
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
export function PlatformRoles() {
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [permOpen, setPermOpen] = useState(false)
  const [activeRoleId, setActiveRoleId] = useState('')
  const [permissionCodes, setPermissionCodes] = useState<string[]>([])
  const [permLoading, setPermLoading] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [form] = Form.useForm()

  const { data, loading, source, error, reload } = useRequest(() => platformApi.listRoles(), { items: [] })
  const { data: permissionData } = useRequest(() => platformApi.listPermissions(), { items: [] })

  const filtered = useMemo(() => {
    if (!keyword) return data.items
    return data.items.filter((item) => item.name.toLowerCase().includes(keyword.toLowerCase()))
  }, [data.items, keyword])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await platformApi.createRole(values)
      uiMessage.success('已创建角色')
      form.resetFields()
      setCreateOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const openPermissions = async (roleId: string) => {
    setActiveRoleId(roleId)
    setPermOpen(true)
    setPermLoading(true)
    try {
      const res = await platformApi.listRolePermissions(roleId)
      setPermissionCodes(res.items.map((item) => item.code))
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    } finally {
      setPermLoading(false)
    }
  }

  const handleAssignPermissions = async () => {
    try {
      await platformApi.assignRolePermissions(activeRoleId, permissionCodes)
      uiMessage.success('已更新角色权限')
      setPermOpen(false)
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader title="平台角色" description="平台角色与权限" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="按角色名称搜索" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <>
            <Button type="primary" onClick={() => setCreateOpen(true)}>
              新建角色
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
                    <Descriptions.Item label="Role ID">{record.id}</Descriptions.Item>
                    <Descriptions.Item label="创建时间">{formatDateTime(record.created_at)}</Descriptions.Item>
                  </Descriptions>
                ),
              }
            : undefined,
          columns: [
            {
              title: '名称',
              dataIndex: 'name',
              render: (_: string, record) => <Link to={`/platform/roles/${record.id}`}>{record.name}</Link>,
            },
            { title: '创建时间', dataIndex: 'created_at', render: (value: string) => formatDateTime(value) },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Button size="small" onClick={() => openPermissions(record.id)}>
                    配置权限
                  </Button>
                </Space>
              ),
            },
          ],
        }}
      />

      <Modal
        title="新建平台角色"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        okText="创建"
      >
        <Form form={form} layout="vertical">
          <Form.Item label="角色名称" name="name" rules={[{ required: true, message: '请输入角色名称' }]}>
            <Input placeholder="例如：Platform Admin" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="配置角色权限"
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
          {permissionData.items.length === 0 ? (
            <Typography.Text className="muted">暂无权限，请先前往「权限目录」创建权限。</Typography.Text>
          ) : null}
        </Form>
      </Modal>
    </div>
  )
}

