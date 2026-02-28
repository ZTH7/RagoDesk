import { Button, Card, Descriptions, Form, Input, Modal, Select, Space, Switch, Tag, Typography } from 'antd'
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
  suspended: 'red',
}

export function Tenants() {
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [keyword, setKeyword] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [form] = Form.useForm()
  const { data, loading, source, error, reload } = useRequest(() => platformApi.listTenants(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (statusFilter && item.status !== statusFilter) return false
      if (keyword && !item.name.toLowerCase().includes(keyword.toLowerCase())) return false
      return true
    })
  }, [data.items, keyword, statusFilter])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await platformApi.createTenant(values)
      uiMessage.success('已创建租户')
      form.resetFields()
      setCreateOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="租户管理"
        description="平台租户列表与套餐状态"
        extra={<DataSourceTag source={source} />}
      />
      <RequestBanner error={error} />
      <Card>
        <FilterBar
          left={
            <Space>
              <Input.Search placeholder="按租户名称搜索" onSearch={setKeyword} allowClear style={{ width: 220 }} />
              <Button type="primary" onClick={() => setCreateOpen(true)}>
                新建租户
              </Button>
            </Space>
          }
          right={
            <Space>
              <Button
                onClick={() => setStatusFilter(statusFilter === 'active' ? '' : 'active')}
                type={statusFilter === 'active' ? 'primary' : 'default'}
              >
                仅显示启用
              </Button>
              <Space size={6}>
                <Typography.Text className="muted">高级列</Typography.Text>
                <Switch checked={showAdvanced} onChange={setShowAdvanced} />
              </Space>
            </Space>
          }
        />
      </Card>
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
                    <Descriptions.Item label="Tenant ID">{record.id}</Descriptions.Item>
                    <Descriptions.Item label="创建时间">{formatDateTime(record.created_at)}</Descriptions.Item>
                  </Descriptions>
                ),
              }
            : undefined,
          columns: [
            {
              title: '名称',
              dataIndex: 'name',
              render: (_: string, record) => <Link to={`/platform/tenants/${record.id}`}>{record.name}</Link>,
            },
            {
              title: '类型',
              dataIndex: 'type',
              render: (value: string) =>
                value === 'enterprise' ? '企业' : value === 'individual' ? '个人' : value,
            },
            { title: '套餐', dataIndex: 'plan' },
            {
              title: '状态',
              dataIndex: 'status',
              render: (status: string) => (
                <Tag color={statusColors[status] || 'default'}>{status === 'active' ? '启用' : '停用'}</Tag>
              ),
            },
            { title: '创建时间', dataIndex: 'created_at', render: (value: string) => formatDateTime(value) },
          ],
        }}
      />

      <Modal
        title="新建租户"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        okText="创建"
      >
        <Form form={form} layout="vertical" initialValues={{ status: 'active', type: 'enterprise' }}>
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入租户名称' }]}>
            <Input placeholder="例如：示例公司" />
          </Form.Item>
          <Form.Item label="类型" name="type" rules={[{ required: true, message: '请选择类型' }]}>
            <Select
              options={[
                { value: 'enterprise', label: '企业' },
                { value: 'individual', label: '个人' },
              ]}
            />
          </Form.Item>
          <Form.Item label="套餐" name="plan" rules={[{ required: true, message: '请输入套餐' }]}>
            <Input placeholder="例如：pro" />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={[{ value: 'active', label: '启用' }, { value: 'suspended', label: '停用' }]} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

