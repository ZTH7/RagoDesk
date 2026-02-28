import {
  Button,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Tag,
  Typography,
} from 'antd'
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

import { uiMessage } from '../../services/uiMessage'
const statusColors: Record<string, string> = {
  active: 'green',
  disabled: 'red',
}

export function ApiKeys() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [rawKeyOpen, setRawKeyOpen] = useState(false)
  const [rawKey, setRawKey] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()

  const { data, loading, source, error, reload } = useRequest(() => consoleApi.listApiKeys(), { items: [] })
  const { data: botData } = useRequest(() => consoleApi.listBots(), { items: [] })

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && item.status !== status) return false
      if (keyword && !item.name.toLowerCase().includes(keyword.toLowerCase())) return false
      return true
    })
  }, [data.items, keyword, status])

  const botOptions = useMemo(
    () =>
      botData.items.map((bot) => ({
        label: bot.name,
        value: bot.id,
      })),
    [botData.items],
  )

  const botNameById = useMemo(() => {
    const map = new Map<string, string>()
    botData.items.forEach((bot) => map.set(bot.id, bot.name))
    return map
  }, [botData.items])

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields()
      const res = await consoleApi.createApiKey(values)
      uiMessage.success('已创建 API Key')
      const raw = (res as any).raw_key ?? (res as any).rawKey ?? ''
      setRawKey(raw)
      setRawKeyOpen(true)
      setCreateOpen(false)
      createForm.resetFields()
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleEdit = async () => {
    if (!editingId) return
    try {
      const values = await editForm.validateFields()
      await consoleApi.updateApiKey(editingId, values)
      uiMessage.success('已更新 API Key')
      setEditOpen(false)
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleToggleStatus = async (id: string, nextStatus: string) => {
    try {
      await consoleApi.updateApiKey(id, { status: nextStatus })
      uiMessage.success('已更新状态')
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleRotate = async (id: string) => {
    try {
      const res = await consoleApi.rotateApiKey(id)
      uiMessage.success('已轮换 API Key')
      const raw = (res as any).raw_key ?? (res as any).rawKey ?? ''
      setRawKey(raw)
      setRawKeyOpen(true)
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await consoleApi.deleteApiKey(id)
      uiMessage.success('已删除 API Key')
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="API Keys"
        description="创建、轮换与管理 API Key"
        extra={<DataSourceTag source={source} />}
      />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索 Key" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
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
              创建 Key
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
                    <Descriptions.Item label="API Key ID">{record.id}</Descriptions.Item>
                    <Descriptions.Item label="Bot ID">{record.bot_id}</Descriptions.Item>
                    <Descriptions.Item label="创建时间">{record.created_at || '-'}</Descriptions.Item>
                    <Descriptions.Item label="最近使用">{record.last_used_at || '-'}</Descriptions.Item>
                  </Descriptions>
                ),
              }
            : undefined,
          columns: [
            {
              title: '名称',
              dataIndex: 'name',
              render: (_: string, record) => (
                <Link to={`/console/api-keys/${record.id}`}>{record.name}</Link>
              ),
            },
            {
              title: 'Bot',
              dataIndex: 'bot_id',
              render: (value: string) => botNameById.get(value) || value,
            },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => <Tag color={statusColors[value] || 'default'}>{value}</Tag>,
            },
            {
              title: 'Scopes',
              dataIndex: 'scopes',
              render: (scopes: string[]) => (Array.isArray(scopes) ? scopes.join(', ') : '-'),
            },
            {
              title: 'API Versions',
              dataIndex: 'api_versions',
              render: (v: string[]) => (Array.isArray(v) ? v.join(', ') : '-'),
            },
            { title: 'Quota', dataIndex: 'quota_daily' },
            { title: 'QPS', dataIndex: 'qps_limit' },
            { title: 'Last Used', dataIndex: 'last_used_at' },
            {
              title: '操作',
              key: 'actions',
              render: (_: unknown, record) => (
                <Space>
                  <Button
                    size="small"
                    onClick={() => {
                      setEditingId(record.id)
                      editForm.setFieldsValue({
                        name: record.name,
                        status: record.status,
                        scopes: record.scopes,
                        api_versions: record.api_versions,
                        quota_daily: record.quota_daily,
                        qps_limit: record.qps_limit,
                      })
                      setEditOpen(true)
                    }}
                  >
                    编辑
                  </Button>
                  <Button size="small" onClick={() => handleRotate(record.id)}>
                    轮换
                  </Button>
                  <Button
                    size="small"
                    onClick={() =>
                      handleToggleStatus(record.id, record.status === 'active' ? 'disabled' : 'active')
                    }
                  >
                    {record.status === 'active' ? '禁用' : '启用'}
                  </Button>
                  <Popconfirm title="确认删除该 Key？" onConfirm={() => handleDelete(record.id)}>
                    <Button size="small" danger>
                      删除
                    </Button>
                  </Popconfirm>
                </Space>
              ),
            },
          ],
        }}
      />

      <Modal
        title="创建 API Key"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        okText="创建"
      >
        <Form form={createForm} layout="vertical" initialValues={{ scopes: [], api_versions: [] }}>
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="例如：客服 API Key" />
          </Form.Item>
          <Form.Item label="Bot" name="bot_id" rules={[{ required: true, message: '请选择 Bot' }]}>
            <Select
              placeholder="选择要绑定的 Bot"
              options={botOptions}
              showSearch
              optionFilterProp="label"
            />
          </Form.Item>
          <Form.Item label="Scopes" name="scopes">
            <Select mode="tags" placeholder="输入或选择 scopes" />
          </Form.Item>
          <Form.Item label="API Versions" name="api_versions">
            <Select mode="tags" placeholder="例如：v1" />
          </Form.Item>
          <Form.Item label="Quota Daily" name="quota_daily" rules={[{ required: true, message: '请输入配额' }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="QPS Limit" name="qps_limit" rules={[{ required: true, message: '请输入 QPS' }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑 API Key"
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleEdit}
        okText="保存"
      >
        <Form form={editForm} layout="vertical">
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
            <Select options={[{ value: 'active', label: 'Active' }, { value: 'disabled', label: 'Disabled' }]} />
          </Form.Item>
          <Form.Item label="Scopes" name="scopes">
            <Select mode="tags" />
          </Form.Item>
          <Form.Item label="API Versions" name="api_versions">
            <Select mode="tags" />
          </Form.Item>
          <Form.Item label="Quota Daily" name="quota_daily">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="QPS Limit" name="qps_limit">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="新的 API Key"
        open={rawKeyOpen}
        onCancel={() => setRawKeyOpen(false)}
        footer={<Button onClick={() => setRawKeyOpen(false)}>关闭</Button>}
      >
        <p>这是唯一一次展示原始 Key，请妥善保存。</p>
        <Input.TextArea value={rawKey} readOnly rows={3} />
      </Modal>
    </div>
  )
}

