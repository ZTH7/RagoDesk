import { Button, DatePicker, Descriptions, Input, Select, Space, Switch, Tag, Typography } from 'antd'
import type { Dayjs } from 'dayjs'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { PageHeader } from '../../components/PageHeader'
import { FilterBar } from '../../components/FilterBar'
import { TableCard } from '../../components/TableCard'
import { DataSourceTag } from '../../components/DataSourceTag'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

import { uiMessage } from '../../services/uiMessage'

const PATH_LABELS: Record<string, string> = {
  '/api/v1/session': '创建会话',
  '/api/v1/message': '发送消息',
}

function resolvePathLabel(path: string) {
  return PATH_LABELS[path] || path || '-'
}

export function ApiUsage() {
  const [status, setStatus] = useState<string>('all')
  const [keyword, setKeyword] = useState('')
  const [botId, setBotId] = useState('all')
  const [apiKeyId, setApiKeyId] = useState('all')
  const [apiVersion, setApiVersion] = useState('')
  const [model, setModel] = useState('')
  const [range, setRange] = useState<[Dayjs | null, Dayjs | null] | null>(null)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [query, setQuery] = useState<{
    bot_id?: string
    api_key_id?: string
    api_version?: string
    model?: string
    start_time?: string
    end_time?: string
  }>({})
  const navigate = useNavigate()

  const { data: botsData } = useRequest(() => consoleApi.listBots(), { items: [] })
  const { data: apiKeysData } = useRequest(() => consoleApi.listApiKeys({ limit: 200 }), { items: [] })

  const { data, loading, source, error } = useRequest(
    () => consoleApi.listUsageLogs(query),
    { items: [] },
    { deps: [query] },
  )

  const filtered = useMemo(() => {
    return data.items.filter((item) => {
      if (status !== 'all' && String(item.status_code) !== status) return false
      const searchable = `${item.path || ''} ${item.model || ''} ${item.api_version || ''}`.toLowerCase()
      if (keyword && !searchable.includes(keyword.toLowerCase())) return false
      return true
    })
  }, [data.items, keyword, status])

  const applyFilters = () => {
    const next: typeof query = {}
    if (botId && botId !== 'all') next.bot_id = botId
    if (apiKeyId && apiKeyId !== 'all') next.api_key_id = apiKeyId
    if (apiVersion) next.api_version = apiVersion
    if (model) next.model = model
    if (range && range[0] && range[1]) {
      next.start_time = range[0].toISOString()
      next.end_time = range[1].toISOString()
    }
    setQuery(next)
  }

  const resetFilters = () => {
    setBotId('all')
    setApiKeyId('all')
    setApiVersion('')
    setModel('')
    setRange(null)
    setQuery({})
  }

  const handleExport = async () => {
    try {
      const res = await consoleApi.exportUsageLogs({
        api_key_id: query.api_key_id,
        bot_id: query.bot_id,
        api_version: query.api_version,
        model: query.model,
        start_time: query.start_time,
        end_time: query.end_time,
        format: 'csv',
      })

      if (res.download_url) {
        window.open(res.download_url, '_blank')
        return
      }

      if (res.content) {
        const blob = new Blob([res.content], { type: res.content_type || 'text/csv' })
        const url = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = res.filename || 'api-usage.csv'
        link.click()
        URL.revokeObjectURL(url)
        return
      }

      uiMessage.info('暂无可下载内容')
    } catch (err) {
      if (err instanceof Error) uiMessage.error(err.message)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="调用日志"
        description="API 使用明细与状态码"
        extra={
          <Space>
            <DataSourceTag source={source} />
            <Button type="default" onClick={() => navigate('/console/api-usage/summary')}>
              查看汇总
            </Button>
          </Space>
        }
      />
      <RequestBanner error={error} />
      <FilterBar
        left={
          <Space wrap>
            <Select
              value={botId}
              style={{ width: 200 }}
              onChange={setBotId}
              options={[{ label: '全部 Bot', value: 'all' }].concat(
                botsData.items.map((bot) => ({ label: bot.name, value: bot.id })),
              )}
              showSearch
              optionFilterProp="label"
            />
            <Select
              value={apiKeyId}
              style={{ width: 220 }}
              onChange={setApiKeyId}
              options={[{ label: '全部 API Key', value: 'all' }].concat(
                apiKeysData.items.map((key) => ({ label: key.name, value: key.id })),
              )}
              showSearch
              optionFilterProp="label"
            />
            <Input
              placeholder="接口版本"
              value={apiVersion}
              onChange={(e) => setApiVersion(e.target.value)}
              style={{ width: 140 }}
            />
            <Input
              placeholder="模型"
              value={model}
              onChange={(e) => setModel(e.target.value)}
              style={{ width: 140 }}
            />
            <DatePicker.RangePicker value={range} onChange={(value) => setRange(value)} />
            <Input.Search
              placeholder="搜索接口/模型"
              onSearch={setKeyword}
              allowClear
              style={{ width: 220 }}
            />
          </Space>
        }
        right={
          <Space>
            <Select
              value={status}
              style={{ width: 160 }}
              onChange={setStatus}
              options={[
                { value: 'all', label: '全部状态' },
                { value: '200', label: '成功(200)' },
                { value: '429', label: '限流(429)' },
                { value: '500', label: '错误(500)' },
              ]}
            />
            <Space size={6}>
              <Typography.Text className="muted">高级列</Typography.Text>
              <Switch checked={showAdvanced} onChange={setShowAdvanced} />
            </Space>
            <Button onClick={applyFilters}>应用筛选</Button>
            <Button onClick={resetFilters}>重置</Button>
            <Button type="primary" onClick={handleExport}>
              导出
            </Button>
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
                    <Descriptions.Item label="Log ID">{record.id}</Descriptions.Item>
                    <Descriptions.Item label="Path">{record.path || '-'}</Descriptions.Item>
                    <Descriptions.Item label="API Version">{record.api_version || '-'}</Descriptions.Item>
                    <Descriptions.Item label="Model">{record.model || '-'}</Descriptions.Item>
                    <Descriptions.Item label="Client IP">{record.client_ip || '-'}</Descriptions.Item>
                    <Descriptions.Item label="User Agent">{record.user_agent || '-'}</Descriptions.Item>
                  </Descriptions>
                ),
                rowExpandable: () => true,
              }
            : undefined,
          columns: [
            {
              title: '调用类型',
              dataIndex: 'path',
              render: (value: string) => resolvePathLabel(value),
            },
            {
              title: '结果',
              dataIndex: 'status_code',
              render: (code: number) => {
                const ok = code < 400
                return <Tag color={ok ? 'green' : 'red'}>{ok ? `成功 (${code})` : `失败 (${code})`}</Tag>
              },
            },
            { title: '响应耗时', dataIndex: 'latency_ms', render: (v: number) => `${v} ms` },
            { title: 'Token 消耗', dataIndex: 'total_tokens' },
            { title: '时间', dataIndex: 'created_at' },
          ],
        }}
      />
    </div>
  )
}
