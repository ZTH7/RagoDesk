import { Button, Card, Col, DatePicker, Input, Row, Select, Space, Statistic, Skeleton } from 'antd'
import type { Dayjs } from 'dayjs'
import { useState } from 'react'
import { PageHeader } from '../../components/PageHeader'
import { RequestBanner } from '../../components/RequestBanner'
import { useRequest } from '../../hooks/useRequest'
import { consoleApi } from '../../services/console'

export function ApiUsageSummary() {
  const [botId, setBotId] = useState('all')
  const [apiKeyId, setApiKeyId] = useState('all')
  const [apiVersion, setApiVersion] = useState('')
  const [model, setModel] = useState('')
  const [range, setRange] = useState<[Dayjs | null, Dayjs | null] | null>(null)
  const [query, setQuery] = useState<{
    bot_id?: string
    api_key_id?: string
    api_version?: string
    model?: string
    start_time?: string
    end_time?: string
  }>({})

  const { data: botsData } = useRequest(() => consoleApi.listBots(), { items: [] })
  const { data: apiKeysData } = useRequest(() => consoleApi.listApiKeys({ limit: 200 }), { items: [] })

  const { data, loading, error } = useRequest(
    () =>
      consoleApi.getUsageSummary({
        bot_id: query.bot_id,
        api_key_id: query.api_key_id,
        api_version: query.api_version,
        model: query.model,
        start_time: query.start_time,
        end_time: query.end_time,
      }),
    {
      summary: {
        total: 0,
        error_count: 0,
        avg_latency_ms: 0,
        prompt_tokens: 0,
        completion_tokens: 0,
        total_tokens: 0,
      },
    },
    { deps: [query] },
  )

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

  return (
    <div className="page">
      <PageHeader title="调用汇总" description="按时间范围汇总的 API 使用情况" />
      <RequestBanner error={error} />
      <Card style={{ marginBottom: 16 }}>
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
            placeholder="API Version"
            value={apiVersion}
            onChange={(e) => setApiVersion(e.target.value)}
            style={{ width: 140 }}
          />
          <Input
            placeholder="Model"
            value={model}
            onChange={(e) => setModel(e.target.value)}
            style={{ width: 140 }}
          />
          <DatePicker.RangePicker value={range} onChange={(value) => setRange(value)} />
          <Button onClick={applyFilters}>应用筛选</Button>
          <Button onClick={resetFilters}>重置</Button>
        </Space>
      </Card>
      {loading ? (
        <Skeleton active paragraph={{ rows: 3 }} />
      ) : (
        <Row gutter={16}>
          <Col xs={24} md={6}>
            <Card>
              <Statistic title="总请求" value={data.summary.total} />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card>
              <Statistic title="错误数" value={data.summary.error_count} />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card>
              <Statistic title="平均延迟" value={data.summary.avg_latency_ms} suffix="ms" />
            </Card>
          </Col>
          <Col xs={24} md={6}>
            <Card>
              <Statistic title="总 Token" value={data.summary.total_tokens} />
            </Card>
          </Col>
        </Row>
      )}
    </div>
  )
}
