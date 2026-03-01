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

type KeyResponseShape = {
  api_key?: {
    id?: string
    public_chat_id?: string
    publicChatId?: string
  }
  apiKey?: {
    id?: string
    public_chat_id?: string
    publicChatId?: string
  }
  raw_key?: string
  rawKey?: string
}

function parseRawKeyAndChatID(input: KeyResponseShape, fallbackKeyID = '') {
  const raw = input.raw_key ?? input.rawKey ?? ''
  const keyID = input.api_key?.id ?? input.apiKey?.id ?? fallbackKeyID
  const chatID =
    input.api_key?.public_chat_id ??
    input.apiKey?.public_chat_id ??
    input.api_key?.publicChatId ??
    input.apiKey?.publicChatId ??
    ''
  return {
    rawKey: raw,
    keyID,
    publicChatID: chatID || keyID,
  }
}

export function ApiKeys() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [rawKeyOpen, setRawKeyOpen] = useState(false)
  const [rawKey, setRawKey] = useState('')
  const [publicChatID, setPublicChatID] = useState('')
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
      const parsed = parseRawKeyAndChatID(res)
      setRawKey(parsed.rawKey)
      setPublicChatID(parsed.publicChatID)
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
      const parsed = parseRawKeyAndChatID(res, id)
      setRawKey(parsed.rawKey)
      setPublicChatID(parsed.publicChatID)
      setRawKeyOpen(true)
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleRegenerateChatLink = async (id: string) => {
    try {
      const res = await consoleApi.regenerateApiKeyPublicChatID(id)
      const nextChatID =
        res.api_key?.public_chat_id ?? (res.api_key as any)?.publicChatId ?? ''
      uiMessage.success('已重置公开聊天链接')
      if (nextChatID) {
        await handleCopyChatLink(nextChatID)
      }
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

  const buildChatLink = (chatID?: string) =>
    chatID ? `${window.location.origin}/chat/${encodeURIComponent(chatID)}` : ''

  const handleCopyChatLink = async (chatID?: string) => {
    const link = buildChatLink(chatID)
    if (!link) {
      uiMessage.error('当前 Key 未配置公开聊天链接')
      return
    }
    try {
      await navigator.clipboard.writeText(link)
      uiMessage.success('已复制聊天链接')
    } catch {
      uiMessage.error('复制失败，请手动复制')
    }
  }

  const chatLink = buildChatLink(publicChatID)

  return (
    <div className="page">
      <PageHeader
        title="接口密钥"
        description="创建、轮换与管理接口密钥"
        extra={<DataSourceTag source={source} />}
      />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索密钥名称" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
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
              创建密钥
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
                      <Descriptions.Item label="创建时间">{formatDateTime(record.created_at)}</Descriptions.Item>
                      <Descriptions.Item label="最近使用">{formatDateTime(record.last_used_at)}</Descriptions.Item>
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
              render: (value: string) => (
                <Tag color={statusColors[value] || 'default'}>{statusLabels[value] || value}</Tag>
              ),
            },
            {
              title: '公开聊天',
              dataIndex: 'public_chat_enabled',
              render: (_: boolean, record) => (
                <Switch
                  size="small"
                  checked={record.public_chat_enabled !== false}
                  onChange={async (checked) => {
                    try {
                      await consoleApi.updateApiKey(record.id, { public_chat_enabled: checked })
                      uiMessage.success(checked ? '已启用公开聊天' : '已关闭公开聊天')
                      reload()
                    } catch (err) {
                      if (err instanceof Error) uiMessage.error(err.message)
                    }
                  }}
                />
              ),
            },
            {
              title: '权限范围',
              dataIndex: 'scopes',
              render: (scopes: string[]) => (Array.isArray(scopes) ? scopes.join(', ') : '-'),
            },
            {
              title: '接口版本',
              dataIndex: 'api_versions',
              render: (v: string[]) => (Array.isArray(v) ? v.join(', ') : '-'),
            },
            { title: '日配额', dataIndex: 'quota_daily' },
            { title: 'QPS', dataIndex: 'qps_limit' },
            { title: '最近使用', dataIndex: 'last_used_at', render: (value: string) => formatDateTime(value) },
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
                        public_chat_enabled: record.public_chat_enabled !== false,
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
                  <Button size="small" onClick={() => handleRegenerateChatLink(record.id)}>
                    重置链接
                  </Button>
                  <Button size="small" onClick={() => handleCopyChatLink(record.public_chat_id)}>
                    复制聊天链接
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
        title="创建接口密钥"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        okText="创建"
      >
        <Form
          form={createForm}
          layout="vertical"
          initialValues={{ scopes: [], api_versions: [], public_chat_enabled: true }}
        >
          <Form.Item label="名称" name="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="例如：客服 API Key" />
          </Form.Item>
          <Form.Item label="机器人" name="bot_id" rules={[{ required: true, message: '请选择机器人' }]}>
            <Select
              placeholder="选择要绑定的机器人"
              options={botOptions}
              showSearch
              optionFilterProp="label"
            />
          </Form.Item>
          <Form.Item label="权限范围" name="scopes">
            <Select mode="tags" placeholder="输入或选择权限范围" />
          </Form.Item>
          <Form.Item label="接口版本" name="api_versions">
            <Select mode="tags" placeholder="例如：v1" />
          </Form.Item>
          <Form.Item label="日配额" name="quota_daily" rules={[{ required: true, message: '请输入配额' }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="QPS Limit" name="qps_limit" rules={[{ required: true, message: '请输入 QPS' }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="公开聊天链接" name="public_chat_enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑接口密钥"
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
            <Select options={[{ value: 'active', label: '启用' }, { value: 'disabled', label: '停用' }]} />
          </Form.Item>
          <Form.Item label="公开聊天链接" name="public_chat_enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item label="权限范围" name="scopes">
            <Select mode="tags" />
          </Form.Item>
          <Form.Item label="接口版本" name="api_versions">
            <Select mode="tags" />
          </Form.Item>
          <Form.Item label="日配额" name="quota_daily">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="QPS Limit" name="qps_limit">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="新接口密钥"
        open={rawKeyOpen}
        onCancel={() => setRawKeyOpen(false)}
        footer={<Button onClick={() => setRawKeyOpen(false)}>关闭</Button>}
      >
        <p>这是唯一一次展示原始 Key，请妥善保存。</p>
        <Input.TextArea value={rawKey} readOnly rows={3} />
        <p style={{ marginTop: 12, marginBottom: 8 }}>面向客户的聊天链接：</p>
        <Input.Group compact>
          <Input value={chatLink} readOnly style={{ width: 'calc(100% - 84px)' }} />
          <Button onClick={() => handleCopyChatLink(publicChatID)}>
            复制链接
          </Button>
        </Input.Group>
        <Typography.Text className="muted" style={{ display: 'block', marginTop: 8 }}>
          建议仅用于公开客服场景，并配置合理的 QPS 与日配额。
        </Typography.Text>
      </Modal>
    </div>
  )
}

