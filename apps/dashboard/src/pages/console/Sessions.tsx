import { Descriptions, Input, Select, Space, Switch, Tag, Typography } from 'antd'
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

export function Sessions() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const { data, loading, source, error } = useRequest(() => consoleApi.listSessions(), { items: [] })
  const { data: botData } = useRequest(() => consoleApi.listBots(), { items: [] })

  const botNameMap = useMemo(() => {
    const map = new Map<string, string>()
    botData.items.forEach((bot) => map.set(bot.id, bot.name))
    return map
  }, [botData.items])

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && item.status !== status) return false
      if (keyword && !`${item.id} ${item.user_external_id || ''} ${item.close_reason || ''}`.toLowerCase().includes(keyword.toLowerCase())) {
        return false
      }
      return true
    })
  }, [data.items, keyword, status])

  return (
    <div className="page">
      <PageHeader title="会话管理" description="会话列表与状态" extra={<DataSourceTag source={source} />} />
      <RequestBanner error={error} />
      <FilterBar
        left={<Input.Search placeholder="搜索用户 / 会话" onSearch={setKeyword} allowClear style={{ width: 220 }} />}
        right={
          <Space>
            <Select
              value={status}
              style={{ width: 160 }}
              onChange={setStatus}
              options={[
                { value: 'all', label: '全部状态' },
                { value: 'bot', label: '进行中' },
                { value: 'closed', label: '已关闭' },
              ]}
            />
            <Space size={6}>
              <Typography.Text className="muted">高级列</Typography.Text>
              <Switch checked={showAdvanced} onChange={setShowAdvanced} />
            </Space>
          </Space>
        }
      />
      <TableCard
        table={{
          rowKey: 'id',
          dataSource: filtered,
          loading,
          pagination: { pageSize: 10 },
          expandable: showAdvanced
            ? {
                expandedRowRender: (record) => (
                    <Descriptions column={2} bordered size="small">
                      <Descriptions.Item label="Session ID">{record.id}</Descriptions.Item>
                      <Descriptions.Item label="Bot ID">{record.bot_id || '-'}</Descriptions.Item>
                      <Descriptions.Item label="关闭原因">{record.close_reason || '-'}</Descriptions.Item>
                      <Descriptions.Item label="创建时间">{formatDateTime(record.created_at)}</Descriptions.Item>
                    </Descriptions>
                ),
                rowExpandable: () => true,
              }
            : undefined,
          columns: [
            {
              title: '会话',
              dataIndex: 'user_external_id',
              render: (_: string, record) => (
                <Space direction="vertical" size={0}>
                  <Link to={`/console/sessions/${record.id}`}>{record.user_external_id || '匿名用户'}</Link>
                  <Typography.Text type="secondary">{botNameMap.get(record.bot_id) || record.bot_id || '-'}</Typography.Text>
                </Space>
              ),
            },
            {
              title: '状态',
              dataIndex: 'status',
              render: (value: string) => (
                <Tag color={value === 'closed' ? 'default' : 'blue'}>{value === 'closed' ? '已关闭' : '进行中'}</Tag>
              ),
            },
            { title: '关闭原因', dataIndex: 'close_reason', render: (value: string) => value || '-' },
            { title: '创建时间', dataIndex: 'created_at', render: (value: string) => formatDateTime(value) },
          ],
        }}
      />
    </div>
  )
}
