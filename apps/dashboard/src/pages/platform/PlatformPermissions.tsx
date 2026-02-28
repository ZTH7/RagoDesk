import { Button, Descriptions, Form, Input, Modal, Select, Space, Switch, Tag, Typography } from 'antd'
import { useMemo, useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { platformApi } from '../../services/platform'
import { formatDateTime } from '../../utils/datetime'

import { uiMessage } from '../../services/uiMessage'
export function PlatformPermissions() {
  const [scope, setScope] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [form] = Form.useForm()
  const { data, loading, source, error, reload } = useRequest(() => platformApi.listPermissions(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (scope !== 'all' && item.scope !== scope) return false
      if (
        keyword &&
        !item.code.toLowerCase().includes(keyword.toLowerCase()) &&
        !(item.description || '').toLowerCase().includes(keyword.toLowerCase())
      ) {
        return false
      }
      return true
    })
  }, [data.items, keyword, scope])

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
          <Space>
            <Input.Search placeholder="按权限名或描述搜索" onSearch={setKeyword} allowClear style={{ width: 220 }} />
            <Select
              value={scope}
              style={{ width: 160 }}
              onChange={setScope}
              options={[
                { value: 'all', label: '全部权限域' },
                { value: 'platform', label: '平台域' },
                { value: 'tenant', label: '租户域' },
              ]}
            />
          </Space>
        }
        right={
          <Space>
            <Button type="primary" onClick={() => setCreateOpen(true)}>
              新建权限
            </Button>
            <Space size={6}>
              <Typography.Text className="muted">高级列</Typography.Text>
              <Switch checked={showAdvanced} onChange={setShowAdvanced} />
            </Space>
          </Space>
        }
      />
      <TableCard
        table={{
          rowKey: 'code',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 8 },
          expandable: showAdvanced
            ? {
                expandedRowRender: (record) => (
                  <Descriptions column={2} bordered size="small">
                    <Descriptions.Item label="权限 Code">{record.code}</Descriptions.Item>
                    <Descriptions.Item label="创建时间">{formatDateTime(record.created_at)}</Descriptions.Item>
                  </Descriptions>
                ),
              }
            : undefined,
          columns: [
            { title: '权限描述', dataIndex: 'description', render: (value?: string) => value || '-' },
            {
              title: '权限域',
              dataIndex: 'scope',
              render: (value: string) => <Tag>{value === 'platform' ? '平台域' : value === 'tenant' ? '租户域' : value}</Tag>,
            },
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
          <Form.Item label="权限域" name="scope" rules={[{ required: true, message: '请选择权限域' }]}>
            <Select options={[{ value: 'platform', label: '平台域' }, { value: 'tenant', label: '租户域' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

