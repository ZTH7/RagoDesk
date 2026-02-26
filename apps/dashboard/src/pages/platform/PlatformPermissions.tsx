import { Button, Form, Input, Modal, Select, Tag } from 'antd'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'

import { uiMessage } from '../../services/uiMessage'
export function PlatformPermissions() {
  const [scope, setScope] = useState<string>('all')
  const [createOpen, setCreateOpen] = useState(false)
  const [form] = Form.useForm()
  const { data, loading, source, error, reload } = useRequest(() => platformApi.listPermissions(), { items: [] })

  const filtered = useMemo(() => {
    if (scope === 'all') return data.items
    return data.items.filter((item) => item.scope === scope)
  }, [data.items, scope])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await platformApi.createPermission(values)
      uiMessage.success('已创建权限')
      setCreateOpen(false)
      form.resetFields()
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader title="权限目录" description="平台权限列表" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar
        left={
          <Select
            value={scope}
            style={{ width: 160 }}
            onChange={setScope}
            options={[
              { value: 'all', label: '全部 Scope' },
              { value: 'platform', label: 'Platform' },
              { value: 'tenant', label: 'Tenant' },
            ]}
          />
        }
        right={
          <Button type="primary" onClick={() => setCreateOpen(true)}>
            新建权限
          </Button>
        }
      />
      <TableCard
        table={{
          rowKey: 'code',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 8 },
          columns: [
            { title: 'Code', dataIndex: 'code' },
            { title: '描述', dataIndex: 'description' },
            { title: 'Scope', dataIndex: 'scope', render: (value: string) => <Tag>{value}</Tag> },
            { title: '创建时间', dataIndex: 'created_at' },
          ],
        }}
      />

      <Modal
        title="新建权限"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        okText="创建"
      >
        <Form form={form} layout="vertical">
          <Form.Item label="权限 Code" name="code" rules={[{ required: true, message: '请输入权限 code' }]}>
            <Input placeholder="例如：platform.tenant.read" />
          </Form.Item>
          <Form.Item label="描述" name="description" rules={[{ required: true, message: '请输入描述' }]}>
            <Input placeholder="描述该权限用途" />
          </Form.Item>
          <Form.Item label="Scope" name="scope" rules={[{ required: true, message: '请选择 scope' }]}>
            <Select options={[{ value: 'platform', label: 'Platform' }, { value: 'tenant', label: 'Tenant' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

