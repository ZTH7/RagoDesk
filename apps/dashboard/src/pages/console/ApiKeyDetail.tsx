import { Button, Card, Descriptions, Empty, Input, Modal, Space, Tag, Popconfirm, Skeleton } from 'antd'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { TechnicalMeta } from '../../components/TechnicalMeta'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'
import { formatDateTime } from '../../utils/datetime'

import { uiMessage } from '../../services/uiMessage'
export function ApiKeyDetail() {
  const { id } = useParams()
  const keyId = id ?? ''
  const [rawKey, setRawKey] = useState('')
  const [rawKeyOpen, setRawKeyOpen] = useState(false)

  const { data, loading, error, reload } = useRequest(() => consoleApi.listApiKeys(), { items: [] })
  const key = data.items.find((item) => item.id === keyId)

  const handleToggle = async () => {
    if (!key) return
    try {
      const nextStatus = key.status === 'active' ? 'disabled' : 'active'
      await consoleApi.updateApiKey(key.id, { status: nextStatus })
      uiMessage.success('已更新状态')
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  const handleRotate = async () => {
    try {
      const res = await consoleApi.rotateApiKey(keyId)
      uiMessage.success('已触发轮换')
      setRawKey(res.raw_key)
      setRawKeyOpen(true)
      reload()
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="接口密钥详情"
        description="查看密钥配置与限制"
        extra={
          <Space>
            <Popconfirm title="确认变更该 Key 状态？" onConfirm={handleToggle}>
              <Button>{key?.status === 'active' ? '禁用' : '启用'}</Button>
            </Popconfirm>
            <Popconfirm title="确认轮换该 Key？" onConfirm={handleRotate}>
              <Button type="primary">轮换</Button>
            </Popconfirm>
          </Space>
        }
      />
      <RequestBanner error={error} />
      <Card>
        {loading ? (
          <Skeleton active paragraph={{ rows: 3 }} />
        ) : !key ? (
          <Empty description="未找到该密钥" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        ) : (
          <Descriptions column={1} bordered size="middle">
            <Descriptions.Item label="状态">
              <Tag color={key.status === 'active' ? 'green' : 'red'}>
                {key.status === 'active' ? '启用' : '停用'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="权限范围">{key.scopes?.join(', ') || '-'}</Descriptions.Item>
            <Descriptions.Item label="接口版本">{key.api_versions?.join(', ') || '-'}</Descriptions.Item>
            <Descriptions.Item label="日配额">{key.quota_daily ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="QPS 限制">{key.qps_limit ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(key.created_at)}</Descriptions.Item>
            <Descriptions.Item label="最近使用">{formatDateTime(key.last_used_at)}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>
      <Card>
        <TechnicalMeta
          items={[
            { key: 'api-key-id', label: 'API Key ID', value: key?.id || keyId },
            { key: 'bot-id', label: 'Bot ID', value: key?.bot_id },
          ]}
        />
      </Card>

      <Modal
        title="新接口密钥"
        open={rawKeyOpen}
        onCancel={() => setRawKeyOpen(false)}
        footer={<Button onClick={() => setRawKeyOpen(false)}>关闭</Button>}
      >
        <p>这是唯一一次展示原始密钥，请妥善保存。</p>
        <Input.TextArea value={rawKey} readOnly rows={3} />
      </Modal>
    </div>
  )
}

